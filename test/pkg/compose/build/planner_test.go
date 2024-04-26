package build

import (
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var _ Proxy = (*proxyMock)(nil)

type proxyMock struct {
	mock.Mock
}

func (f *proxyMock) NewBlueprintBuild(path string, opts *PlannerOpts) (Build, error) {
	f.Called(path, opts)
	return &BlueprintBuild{
		build: build{
			Name: filepath.Base(path),
			Path: path,
		},
	}, nil
}

func (f *proxyMock) NewContainerfileBuild(path string, opts *PlannerOpts) (Build, error) {
	f.Called(path, opts)
	return &ContainerfileBuild{
		build: build{
			Name: filepath.Base(path),
			Path: path,
		},
	}, nil
}

func (f *proxyMock) NewImageFetcher(path string, opts *PlannerOpts) (Build, error) {
	f.Called(path, opts)
	return &ImageFetcher{
		build: build{
			Name: filepath.Base(path),
			Path: path,
		},
	}, nil
}

var blueprintsFS = fstest.MapFS{
	"layer1-base/group1/centos9.image-fetcher":                            &fstest.MapFile{},
	"layer1-base/group1/rhel92.image-installer":                           &fstest.MapFile{},
	"layer1-base/group1/rhel92.toml":                                      &fstest.MapFile{},
	"layer1-base/group1/rhel93.image-installer":                           &fstest.MapFile{},
	"layer1-base/group1/rhel93.toml":                                      &fstest.MapFile{},
	"layer1-base/group2/rhel92-microshift-previous-minor.image-installer": &fstest.MapFile{},
	"layer1-base/group2/rhel92-microshift-previous-minor.toml":            &fstest.MapFile{},
	"layer1-base/group2/rhel92-microshift-yminus2.toml":                   &fstest.MapFile{},
	"layer2-presubmit/group1/rhel92-crel.toml":                            &fstest.MapFile{},
	"layer2-presubmit/group1/rhel92-source-base.toml":                     &fstest.MapFile{},
	"layer2-presubmit/group1/rhel92-source-fake-next-minor.toml":          &fstest.MapFile{},
	"layer2-presubmit/group1/rhel92-source.alias":                         &fstest.MapFile{},
	"layer2-presubmit/group1/rhel92-source.toml":                          &fstest.MapFile{},
	"layer3-periodic/group1/rhel92-crel-with-optionals.toml":              &fstest.MapFile{},
	"layer3-periodic/group1/rhel92-source-isolated.image-installer":       &fstest.MapFile{},
	"layer3-periodic/group1/rhel92-source-isolated.toml":                  &fstest.MapFile{},
	"layer3-periodic/group1/rhel92-source-with-optionals.toml":            &fstest.MapFile{},
	"layer3-periodic/group1/rhel93-crel.toml":                             &fstest.MapFile{},
	"layer3-periodic/group1/rhel93-source.toml":                           &fstest.MapFile{},
	"layer4-ext/group1/rhel92_microshift-ext.image-installer":             &fstest.MapFile{},
	"layer4-ext/group1/rhel92_microshift-ext.toml":                        &fstest.MapFile{},
	"layer4-ext/group1/rhel93_microshift-ext.image-installer":             &fstest.MapFile{},
	"layer4-ext/group1/rhel93_microshift-ext.toml":                        &fstest.MapFile{},
	"layer5-bootc/group1/cos9-bootc-source.containerfile":                 &fstest.MapFile{},
	"layer5-bootc/group1/rhel94-bootc-source.containerfile":               &fstest.MapFile{},
	"layer5-bootc/group2/cos9-bootc-source-optionals.containerfile":       &fstest.MapFile{},
	"layer5-bootc/group2/rhel94-bootc-source-optionals.containerfile":     &fstest.MapFile{},
}

func Test_Planner_InstallersAndAliasesAreNotGeneratingNewBuild(t *testing.T) {
	proxy := &proxyMock{}
	opts := &PlannerOpts{
		SourceOnly:   false,
		Proxy:        proxy,
		BlueprintsFS: blueprintsFS,
	}

	proxy.On("NewBlueprintBuild", "layer1-base/group1/rhel92.toml", mock.Anything)
	proxy.On("NewImageFetcher", "layer1-base/group1/centos9.image-fetcher", mock.Anything)
	proxy.On("NewBlueprintBuild", "layer1-base/group1/rhel92.toml", mock.Anything)
	proxy.On("NewBlueprintBuild", "layer1-base/group1/rhel93.toml", mock.Anything)
	proxy.On("NewBlueprintBuild", "layer1-base/group2/rhel92-microshift-previous-minor.toml", mock.Anything)
	proxy.On("NewBlueprintBuild", "layer1-base/group2/rhel92-microshift-yminus2.toml", mock.Anything)

	planner := &Planner{Opts: opts}
	buildPlan, err := planner.CreateBuildPlan([]string{"layer1-base"})
	assert.NoError(t, err)
	assert.Len(t, buildPlan, 2)
	proxy.AssertExpectations(t)
}

func Test_Planner_BuildPlanShouldBeDividedByGroupsNotLayers(t *testing.T) {
	// Verifies if layers with multiple groups end up as multiple BuildGroups to correctly build in sequence.
	proxy := &proxyMock{}
	opts := &PlannerOpts{
		SourceOnly:   false,
		Proxy:        proxy,
		BlueprintsFS: blueprintsFS,
	}

	proxy.On("NewBlueprintBuild", mock.Anything, mock.Anything)
	proxy.On("NewImageFetcher", mock.Anything, mock.Anything)
	proxy.On("NewContainerfileBuild", mock.Anything, mock.Anything)

	planner := &Planner{Opts: opts}
	buildPlan, err := planner.CreateBuildPlan([]string{"layer1-base", "layer2-presubmit", "layer3-periodic", "layer4-ext", "layer5-bootc"})
	assert.NoError(t, err)
	assert.Len(t, buildPlan, 7)
	proxy.AssertExpectations(t)
}

func Test_Planner_SourceOnly(t *testing.T) {
	proxy := &proxyMock{}
	opts := &PlannerOpts{
		SourceOnly:   true,
		Proxy:        proxy,
		BlueprintsFS: blueprintsFS,
	}

	// Builds should not be created for following files:
	// - layer2-presubmit/group1/rhel92-crel.toml - not source
	// - layer2-presubmit/group1/rhel92-source.alias - aliases are handled by NewBlueprintBuild for matching .toml
	// - layer3-periodic/group1/rhel92-crel-with-optionals.toml - not source
	// - layer3-periodic/group1/rhel92-source-isolated.image-installer - installers are handled by NewBlueprintBuild for matching .toml
	// - layer3-periodic/group1/rhel93-crel.toml - not source

	// Only following calls are expected:
	proxy.On("NewBlueprintBuild", "layer2-presubmit/group1/rhel92-source-base.toml", mock.Anything)
	proxy.On("NewBlueprintBuild", "layer2-presubmit/group1/rhel92-source-fake-next-minor.toml", mock.Anything)
	proxy.On("NewBlueprintBuild", "layer2-presubmit/group1/rhel92-source.toml", mock.Anything)
	proxy.On("NewBlueprintBuild", "layer3-periodic/group1/rhel92-source-isolated.toml", mock.Anything)
	proxy.On("NewBlueprintBuild", "layer3-periodic/group1/rhel92-source-with-optionals.toml", mock.Anything)
	proxy.On("NewBlueprintBuild", "layer3-periodic/group1/rhel93-source.toml", mock.Anything)

	planner := &Planner{Opts: opts}
	buildPlan, err := planner.CreateBuildPlan([]string{"layer2-presubmit", "layer3-periodic"})
	assert.NoError(t, err)
	assert.Len(t, buildPlan, 2)
	proxy.AssertExpectations(t)
}

func Test_Planner_SingleGroup(t *testing.T) {
	proxy := &proxyMock{}
	opts := &PlannerOpts{
		Proxy:        proxy,
		BlueprintsFS: blueprintsFS,
	}

	// Only following calls are expected:
	proxy.On("NewBlueprintBuild", "layer4-ext/group1/rhel92_microshift-ext.toml", mock.Anything)
	proxy.On("NewBlueprintBuild", "layer4-ext/group1/rhel93_microshift-ext.toml", mock.Anything)

	planner := &Planner{Opts: opts}
	buildPlan, err := planner.CreateBuildPlan([]string{"layer4-ext/group1"})
	assert.NoError(t, err)
	assert.Len(t, buildPlan, 1)
	proxy.AssertExpectations(t)
}

func Test_Planner_BlueprintDirectly(t *testing.T) {
	proxy := &proxyMock{}
	opts := &PlannerOpts{
		Proxy:        proxy,
		BlueprintsFS: blueprintsFS,
	}

	proxy.On("NewBlueprintBuild", "layer1-base/group1/rhel92.toml", mock.Anything)
	proxy.On("NewBlueprintBuild", "layer1-base/group1/rhel93.toml", mock.Anything)

	planner := &Planner{Opts: opts}
	buildPlan, err := planner.CreateBuildPlan([]string{"layer1-base/group1/rhel92.toml", "layer1-base/group1/rhel93.toml"})
	assert.NoError(t, err)
	assert.Len(t, buildPlan, 2)
	proxy.AssertExpectations(t)
}

func Test_Planner_ImageFetcher(t *testing.T) {
	proxy := &proxyMock{}
	opts := &PlannerOpts{
		Proxy:        proxy,
		BlueprintsFS: blueprintsFS,
	}

	proxy.On("NewImageFetcher", "layer1-base/group1/centos9.image-fetcher", mock.Anything)

	planner := &Planner{Opts: opts}
	buildPlan, err := planner.CreateBuildPlan([]string{"layer1-base/group1/centos9.image-fetcher"})
	assert.NoError(t, err)
	assert.Len(t, buildPlan, 1)
	proxy.AssertExpectations(t)
}

func Test_Planner_Containerfile(t *testing.T) {
	proxy := &proxyMock{}
	opts := &PlannerOpts{
		Proxy:        proxy,
		BlueprintsFS: blueprintsFS,
	}

	proxy.On("NewContainerfileBuild", "layer5-bootc/group1/cos9-bootc-source.containerfile", mock.Anything)

	planner := &Planner{Opts: opts}
	buildPlan, err := planner.CreateBuildPlan([]string{"layer5-bootc/group1/cos9-bootc-source.containerfile"})
	assert.NoError(t, err)
	assert.Len(t, buildPlan, 1)
	proxy.AssertExpectations(t)
}

func Test_Planner_PassingJustInstallerIsNotSupported(t *testing.T) {
	proxy := &proxyMock{}
	opts := &PlannerOpts{
		Proxy:        proxy,
		BlueprintsFS: blueprintsFS,
	}

	planner := &Planner{Opts: opts}
	_, err := planner.CreateBuildPlan([]string{"layer1-base/group1/rhel92.image-installer"})
	assert.Error(t, err)
	proxy.AssertExpectations(t)
}
