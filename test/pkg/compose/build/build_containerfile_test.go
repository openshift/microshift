package build

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/openshift/microshift/test/pkg/compose/templatingdata"
	"github.com/openshift/microshift/test/pkg/testutil"
	"github.com/openshift/microshift/test/pkg/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_NewContainerfileBuild(t *testing.T) {
	events := &mocks.EventManagerMock{}
	opts := &PlannerOpts{
		TplData: &templatingdata.TemplatingData{
			Source: templatingdata.Release{
				Version: "4.17.0",
			},
			Previous: templatingdata.Release{
				Version: "4.16.0",
				Minor:   16,
			},
		},
		Events: events,
	}
	path := "image-blueprints/layer5-bootc/group0/centos9.containerfile"
	expectedName := "centos9"

	events.On("AddEvent", mock.MatchedBy(EventShouldBe(expectedName, "containerfile", &testutil.Event{}))).Once()

	cb, err := NewContainerfileBuild(path, opts)

	assert.NoError(t, err)
	assert.Equal(t, expectedName, cb.Name)
	assert.Equal(t, path, cb.Path)
}

func Test_ContainerfileBuild_Execute(t *testing.T) {
	bootcimages := "_output/boot-ci-images"
	blueprints := "image-blueprints/"
	rpmRepos := "_output/rpm-repos"
	name := "centos9"
	path := "layer5-bootc/group0/centos9.containerfile"
	expectedDest := filepath.Join(bootcimages, name)

	cb := ContainerfileBuild{
		build: build{
			Name: name,
			Path: path,
		},
	}

	t.Run("First run", func(t *testing.T) {
		events := &mocks.EventManagerMock{}
		podman := &mocks.PodmanMock{}
		utils := &utilProxyMock{}
		events.On("AddEvent", mock.MatchedBy(EventShouldBe(name, "containerfile", &testutil.Event{}))).Once()
		utils.On("PathExistsAndIsNotEmpty", expectedDest).Return(false, nil).Once()
		podman.On("BuildAndSave", mock.Anything, name, filepath.Join(blueprints, path), rpmRepos, expectedDest).Return(nil)

		opts := &Opts{
			Paths: &testutil.Paths{
				BootCImages:         bootcimages,
				ImageBlueprintsPath: blueprints,
				RPMRepos:            rpmRepos,
			},
			Events: events,
			Podman: podman,
			Utils:  utils,
			Force:  false,
		}

		err := cb.Execute(context.TODO(), opts)
		assert.NoError(t, err)

		events.AssertExpectations(t)
		utils.AssertExpectations(t)
		podman.AssertExpectations(t)
	})

	t.Run("Image already exists", func(t *testing.T) {
		events := &mocks.EventManagerMock{}
		utils := &utilProxyMock{}
		events.On("AddEvent",
			mock.MatchedBy(EventShouldBe(name, "containerfile", &testutil.SkippedEvent{}))).Once()
		utils.On("PathExistsAndIsNotEmpty", expectedDest).Return(true, nil).Once()

		opts := &Opts{
			Paths: &testutil.Paths{
				BootCImages:         bootcimages,
				ImageBlueprintsPath: blueprints,
				RPMRepos:            rpmRepos,
			},
			Events: events,
			Utils:  utils,
			Force:  false,
		}
		err := cb.Execute(context.TODO(), opts)
		assert.NoError(t, err)

		events.AssertExpectations(t)
		utils.AssertExpectations(t)
	})

	t.Run("Image already exists but rebuild is forced", func(t *testing.T) {
		events := &mocks.EventManagerMock{}
		podman := &mocks.PodmanMock{}
		utils := &utilProxyMock{}
		events.On("AddEvent",
			mock.MatchedBy(EventShouldBe(name, "containerfile", &testutil.Event{}))).Once()
		utils.On("PathExistsAndIsNotEmpty", expectedDest).Return(true, nil).Once()
		utils.On("RemoveAll", expectedDest).Return(nil).Once()
		podman.On("BuildAndSave", mock.Anything, name, filepath.Join(blueprints, path), rpmRepos, expectedDest).Return(nil)
		opts := &Opts{
			Paths: &testutil.Paths{
				BootCImages:         bootcimages,
				ImageBlueprintsPath: blueprints,
				RPMRepos:            rpmRepos,
			},
			Events: events,
			Utils:  utils,
			Podman: podman,
			Force:  true,
		}

		err := cb.Execute(context.TODO(), opts)
		assert.NoError(t, err)

		events.AssertExpectations(t)
		utils.AssertExpectations(t)
		podman.AssertExpectations(t)
	})

	t.Run("Podman fails", func(t *testing.T) {
		events := &mocks.EventManagerMock{}
		podman := &mocks.PodmanMock{}
		utils := &utilProxyMock{}
		events.On("AddEvent",
			mock.MatchedBy(EventShouldBe(name, "containerfile", &testutil.FailedEvent{}))).Once()
		utils.On("PathExistsAndIsNotEmpty", expectedDest).Return(false, nil).Once()
		podman.On("BuildAndSave", mock.Anything, name, filepath.Join(blueprints, path), rpmRepos, expectedDest).Return(fmt.Errorf("podman failed"))

		opts := &Opts{
			Paths: &testutil.Paths{
				BootCImages:         bootcimages,
				ImageBlueprintsPath: blueprints,
				RPMRepos:            rpmRepos,
			},
			Events: events,
			Utils:  utils,
			Podman: podman,
			Force:  false,
		}

		err := cb.Execute(context.TODO(), opts)
		assert.ErrorContains(t, err, "podman failed")

		events.AssertExpectations(t)
		utils.AssertExpectations(t)
		podman.AssertExpectations(t)
	})

}
