package build

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"
	"testing/fstest"
	"time"

	"github.com/openshift/microshift/test/pkg/compose/templatingdata"
	"github.com/openshift/microshift/test/pkg/testutil"
	"github.com/openshift/microshift/test/pkg/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_NewBlueprintBuild(t *testing.T) {
	var blueprintBuildFS = fstest.MapFS{
		"layer1-base/group1/rhel92.image-installer": &fstest.MapFile{Data: []byte{}},
		"layer1-base/group1/rhel92.toml": &fstest.MapFile{
			Data: []byte(`name = "rhel-9.2"
description = "Basic RHEL 9.2 edge image without MicroShift."
version = "0.0.1"
distro = "rhel-92"
modules = []
groups = []

[[packages]]
name = "microshift-test-agent"
version = "*"

[[packages]]
name = "iproute-tc"
version = "*"

[customizations.services]
enabled = ["microshift-test-agent"]

[customizations.firewall]
ports = ["22:tcp"]

[customizations.firewall.services]
enabled = ["ssh"]
`)},

		"layer1-base/group2/rhel92-microshift-previous-minor.image-installer": &fstest.MapFile{},
		"layer1-base/group2/rhel92-microshift-previous-minor.toml": &fstest.MapFile{
			Data: []byte(`name = "rhel-9.2-microshift-4.{{ .Previous.Minor }}"
description = "RHEL 9.2 with MicroShift from the previous minor version (y-stream) installed."
version = "0.0.1"
modules = []
groups = []
distro = "rhel-92"

[[packages]]
name = "microshift"
version = "{{ .Previous.Version }}"

[[packages]]
name = "microshift-greenboot"
version = "{{ .Previous.Version }}"

[[packages]]
name = "microshift-networking"
version = "{{ .Previous.Version }}"

[[packages]]
name = "microshift-selinux"
version = "{{ .Previous.Version }}"

[[packages]]
name = "microshift-test-agent"
version = "*"

[customizations.services]
enabled = ["microshift", "microshift-test-agent"]

[customizations.firewall]
ports = ["22:tcp", "80:tcp", "443:tcp", "5353:udp", "6443:tcp", "30000-32767:tcp", "30000-32767:udp"]

[customizations.firewall.services]
enabled = ["mdns", "ssh", "http", "https"]

[[customizations.firewall.zones]]
name = "trusted"
sources = ["10.42.0.0/16", "169.254.169.1"]
`),
		},

		"layer2-presubmit/group1/rhel92-source.alias": &fstest.MapFile{
			Data: []byte(`rhel-9.2-microshift-source-aux
		
		`)}, // intentional empty lines and tabs
		"layer2-presubmit/group1/rhel92-source.toml": &fstest.MapFile{
			Data: []byte(`name = "rhel-9.2-microshift-source"
description = "A RHEL 9.2 image with the RPMs built from source."
version = "0.0.1"
modules = []
groups = []
distro = "rhel-92"

[[packages]]
name = "microshift"
version = "{{ .Source.Version }}"

[[packages]]
name = "microshift-greenboot"
version = "{{ .Source.Version }}"

[[packages]]
name = "microshift-networking"
version = "{{ .Source.Version }}"

[[packages]]
name = "microshift-selinux"
version = "{{ .Source.Version }}"

[[packages]]
name = "microshift-test-agent"
version = "*"

[[packages]]
name = "systemd-resolved"
version = "*"

[customizations.services]
enabled = ["microshift", "microshift-test-agent"]

[customizations.firewall]
ports = ["22:tcp", "80:tcp", "443:tcp", "5353:udp", "6443:tcp", "30000-32767:tcp", "30000-32767:udp"]

[customizations.firewall.services]
enabled = ["mdns", "ssh", "http", "https"]

[[customizations.firewall.zones]]
name = "trusted"
sources = ["10.42.0.0/16", "169.254.169.1"]
`)},

		"layer2-presubmit/group1/rhel92-source-2.alias": &fstest.MapFile{
			Data: []byte(`rhel-9.2-microshift-source-aux
` + `	` + `
rhel-9.2-microshift-source-aux-2`)}, // intentional empty lines and tabs
		"layer2-presubmit/group1/rhel92-source-2.toml": &fstest.MapFile{
			Data: []byte(`name = "rhel-9.2-microshift-source"
description = "A RHEL 9.2 image with the RPMs built from source."
version = "0.0.1"
modules = []
groups = []
distro = "rhel-92"

[[packages]]
name = "microshift"
version = "{{ .Source.Version }}"

[[packages]]
name = "microshift-greenboot"
version = "{{ .Source.Version }}"

[[packages]]
name = "microshift-networking"
version = "{{ .Source.Version }}"

[[packages]]
name = "microshift-selinux"
version = "{{ .Source.Version }}"

[[packages]]
name = "microshift-test-agent"
version = "*"

[[packages]]
name = "systemd-resolved"
version = "*"

[customizations.services]
enabled = ["microshift", "microshift-test-agent"]

[customizations.firewall]
ports = ["22:tcp", "80:tcp", "443:tcp", "5353:udp", "6443:tcp", "30000-32767:tcp", "30000-32767:udp"]

[customizations.firewall.services]
enabled = ["mdns", "ssh", "http", "https"]

[[customizations.firewall.zones]]
name = "trusted"
sources = ["10.42.0.0/16", "169.254.169.1"]
`)},
	}

	events := &mocks.EventManagerMock{}
	storageDir := "iso-dest-dir"
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
		BuildInstallers: true,
		BlueprintsFS:    blueprintBuildFS,
		Paths:           &testutil.Paths{VMStorageDir: storageDir},
		Events:          events,
	}

	testData := []struct {
		TestName        string
		Name            string
		Path            string
		Installer       bool
		Parent          string
		Aliases         []string
		ContentContains string
	}{
		{
			TestName:  "+installer -parent -aliases",
			Name:      "rhel-9.2",
			Path:      "layer1-base/group1/rhel92.toml",
			Installer: true,
			Parent:    "",
			Aliases:   nil,
		},
		{
			TestName:        "+parent +alias -installer",
			Name:            "rhel-9.2-microshift-source",
			Path:            "layer2-presubmit/group1/rhel92-source.toml",
			Installer:       false,
			Parent:          "rhel-9.2",
			Aliases:         []string{"rhel-9.2-microshift-source-aux"},
			ContentContains: `version = "4.17.0"`,
		},
		{
			TestName:        "+parent +multiple_aliases -installer",
			Name:            "rhel-9.2-microshift-source",
			Path:            "layer2-presubmit/group1/rhel92-source-2.toml",
			Installer:       false,
			Parent:          "rhel-9.2",
			Aliases:         []string{"rhel-9.2-microshift-source-aux", "rhel-9.2-microshift-source-aux-2"},
			ContentContains: `version = "4.17.0"`,
		},
		{
			TestName:        "+parent +installer +templated-name",
			Name:            "rhel-9.2-microshift-4.16",
			Path:            "layer1-base/group2/rhel92-microshift-previous-minor.toml",
			Installer:       true,
			Parent:          "rhel-9.2",
			Aliases:         nil,
			ContentContains: `version = "4.16.0"`,
		},
	}

	for _, td := range testData {
		t.Run(td.TestName, func(t *testing.T) {
			events.On("AddEvent", mock.MatchedBy(EventShouldBe(td.Name, "", &testutil.Event{}))).Once()

			bb, err := NewBlueprintBuild(td.Path, opts)

			assert.NoError(t, err)
			assert.NotNil(t, bb)
			assert.Equal(t, td.Name, bb.Name)
			assert.Equal(t, td.Path, bb.Path)

			if td.Installer {
				assert.True(t, bb.Installer)
				assert.Equal(t, filepath.Join(storageDir, fmt.Sprintf("%s.iso", td.Name)), bb.InstallerDestination)
			} else {
				assert.False(t, bb.Installer)
			}

			assert.Equal(t, td.Parent, bb.Parent)
			assert.Equal(t, td.Aliases, bb.Aliases)

			if td.ContentContains != "" {
				assert.Contains(t, bb.Contents, td.ContentContains)
			}
		})
	}
}

func Test_Prepare(t *testing.T) {
	composer := &mocks.ComposerMock{}
	events := &mocks.EventManagerMock{}
	name := "rhel-9.2"
	contents := "dummy"
	blueprint := &BlueprintBuild{
		build: build{
			Name: name,
		},
		Contents: contents,
	}
	opts := &Opts{
		Composer: composer,
		Events:   events,
	}

	t.Run("AddBlueprint fails", func(t *testing.T) {
		e := errors.New("some error")
		composer.On("AddBlueprint", contents).Return(e).Once()
		events.On("AddEvent", mock.MatchedBy(EventShouldBe(name, "add", &testutil.FailedEvent{}))).Once()
		err := blueprint.Prepare(opts)
		assert.Error(t, err)
		composer.AssertExpectations(t)
	})

	t.Run("DepsolveBlueprint fails", func(t *testing.T) {
		e := errors.New("some error")
		composer.On("AddBlueprint", contents).Return(nil).Once()
		composer.On("DepsolveBlueprint", name).Return(e).Once()
		events.On("AddEvent", mock.MatchedBy(EventShouldBe(name, "depsolve", &testutil.FailedEvent{}))).Once()
		err := blueprint.Prepare(opts)
		assert.Error(t, err)
		composer.AssertExpectations(t)
	})

	t.Run("No errors", func(t *testing.T) {
		composer.On("AddBlueprint", contents).Return(nil).Once()
		composer.On("DepsolveBlueprint", name).Return(nil).Once()
		events.On("AddEvent", mock.MatchedBy(EventShouldBe(name, "depsolve", &testutil.Event{}))).Once()
		err := blueprint.Prepare(opts)
		assert.NoError(t, err)
		composer.AssertExpectations(t)
	})
}

func Test_Execute(t *testing.T) {
	name := "rhel-9.2-microshift-source"
	parent := "rhel-9.2"
	contents := "dummy"
	commitID := "1-2-3-4"
	installerID := "5-6-7-8"
	installerDest := "storage-dir/rhel-9.2-microshift-source.iso"

	blueprint := &BlueprintBuild{
		build: build{
			Name: name,
		},
		Contents:             contents,
		Parent:               parent,
		Installer:            true,
		InstallerDestination: installerDest,
		Aliases:              []string{"dummy-alias", "another-alias"},
	}

	testData := []struct {
		TestName  string
		RefExists bool
		IsoExists bool
		Force     bool
	}{
		{
			TestName:  "First build on clean hypervisor",
			RefExists: false,
			IsoExists: false,
			Force:     false,
		},
		{
			TestName:  "Ref exists no-force",
			RefExists: true,
			IsoExists: false,
			Force:     false,
		},
		{
			TestName:  "ISO exists no-force",
			RefExists: false,
			IsoExists: true,
			Force:     false,
		},
		{
			TestName:  "Ref and ISO exist no-force",
			RefExists: true,
			IsoExists: true,
			Force:     false,
		},
		{
			TestName:  "Ref and ISO exist with force",
			RefExists: true,
			IsoExists: true,
			Force:     true,
		},
	}

	for _, td := range testData {
		t.Run(td.TestName, func(t *testing.T) {
			composer := &mocks.ComposerMock{}
			events := &mocks.EventManagerMock{}
			ostree := &mocks.OstreeMock{}
			utilProxy := &utilProxyMock{}
			opts := &Opts{Composer: composer, Events: events, Ostree: ostree, Utils: utilProxy,
				Force:   td.Force,
				Retries: 1, RetryInterval: time.Millisecond,
				Paths: &testutil.Paths{VMStorageDir: "storage-dir/"},
			}

			ostree.On("DoesRefExists", name).Return(td.RefExists, nil)
			commitShouldBeSkipped := td.RefExists && !td.Force
			if commitShouldBeSkipped {
				events.On("AddEvent", mock.MatchedBy(EventShouldBe(name, "commit", &testutil.SkippedEvent{}))).Once()
			} else {
				events.On("AddEvent", mock.MatchedBy(EventShouldBe(name, "commit", &testutil.Event{}))).Once()
				composer.On("StartOSTreeCompose", name, "edge-commit", name, parent).Return(commitID, nil).Once()
				composer.On("WaitForCompose", mock.Anything, commitID, mock.Anything, mock.Anything).Return(nil).Once()
				composer.On("SaveComposeMetadata", commitID, mock.Anything).Return(nil).Once()
				composer.On("SaveComposeLogs", commitID, mock.Anything).Return(nil).Once()
				composer.On("SaveComposeImage", commitID, mock.Anything, ".tar").Return("some-commit-path.tar", nil).Once()
				ostree.On("ExtractCommit", "some-commit-path.tar").Return(nil).Once()
			}

			utilProxy.On("PathExistsAndIsNotEmpty", installerDest).Return(td.IsoExists, nil)
			isoShouldBeSkipped := td.IsoExists && !td.Force
			if isoShouldBeSkipped {
				events.On("AddEvent", mock.MatchedBy(EventShouldBe(name, "installer", &testutil.SkippedEvent{}))).Once()
			} else {
				events.On("AddEvent", mock.MatchedBy(EventShouldBe(name, "installer", &testutil.Event{}))).Once()
				composer.On("StartCompose", name, "image-installer").Return(installerID, nil).Once()
				composer.On("WaitForCompose", mock.Anything, installerID, mock.Anything, mock.Anything).Return(nil).Once()
				composer.On("SaveComposeMetadata", installerID, mock.Anything).Return(nil).Once()
				composer.On("SaveComposeLogs", installerID, mock.Anything).Return(nil).Once()
				composer.On("SaveComposeImage", installerID, mock.Anything, ".iso").Return("some-installer.iso", nil).Once()
				utilProxy.On("Rename", "some-installer.iso", installerDest).Return(nil).Once()
			}

			if !commitShouldBeSkipped {
				ostree.On("CreateAlias", name, blueprint.Aliases).Return(nil)
				events.On("AddEvent", mock.MatchedBy(EventShouldBe(name, "alias", &testutil.Event{}))).Once()
			}

			err := blueprint.Execute(context.TODO(), opts)
			assert.NoError(t, err)

			composer.AssertExpectations(t)
			events.AssertExpectations(t)
			ostree.AssertExpectations(t)
			utilProxy.AssertExpectations(t)
		})
	}
}

func Test_Execute_AllErrorsShouldBeJoinedAndReturned(t *testing.T) {
	name := "rhel-9.2-microshift-source"
	parent := "rhel-9.2"
	contents := "dummy"
	commitID := "1-2-3-4"
	installerID := "5-6-7-8"
	retries := 3
	installerDest := "storage-dir/rhel-9.2-microshift-source.iso"

	blueprint := &BlueprintBuild{
		build: build{
			Name: name,
		},
		Contents:             contents,
		Parent:               parent,
		Installer:            true,
		InstallerDestination: installerDest,
		Aliases:              []string{"dummy-alias", "another-alias"},
	}

	t.Run("Start*Compose repeatedly fails", func(t *testing.T) {
		composer := &mocks.ComposerMock{}
		events := &mocks.EventManagerMock{}
		ostree := &mocks.OstreeMock{}
		utilProxy := &utilProxyMock{}
		opts := &Opts{Composer: composer, Events: events, Ostree: ostree, Utils: utilProxy,
			Force:         false,
			Retries:       retries,
			RetryInterval: time.Millisecond,
			Paths:         &testutil.Paths{VMStorageDir: "storage-dir/"},
		}

		// commit
		ostree.On("DoesRefExists", name).Return(false, nil)
		composer.On("StartOSTreeCompose", name, "edge-commit", name, parent).Return("", errors.New("commit start compose error")).Times(retries)
		events.On("AddEvent", mock.MatchedBy(EventShouldBe(name, "commit", &testutil.FailedEvent{}))).Once()

		// installer
		utilProxy.On("PathExistsAndIsNotEmpty", installerDest).Return(false, nil)
		composer.On("StartCompose", name, "image-installer").Return(installerID, errors.New("installer start compose error"))
		events.On("AddEvent", mock.MatchedBy(EventShouldBe(name, "installer", &testutil.FailedEvent{}))).Once()

		err := blueprint.Execute(context.TODO(), opts)
		assert.Error(t, err)
		// Verify that errors are joined and returned together
		assert.ErrorContains(t, err, "commit start compose error")
		assert.ErrorContains(t, err, "installer start compose error")

		composer.AssertExpectations(t)
		events.AssertExpectations(t)
	})

	t.Run("Waiting for compose fails", func(t *testing.T) {
		composer := &mocks.ComposerMock{}
		events := &mocks.EventManagerMock{}
		ostree := &mocks.OstreeMock{}
		utilProxy := &utilProxyMock{}
		opts := &Opts{Composer: composer, Events: events, Ostree: ostree, Utils: utilProxy,
			Force:         false,
			Retries:       retries,
			RetryInterval: time.Millisecond,
			Paths:         &testutil.Paths{VMStorageDir: "storage-dir/"},
		}

		// commit
		ostree.On("DoesRefExists", name).Return(false, nil)
		composer.On("StartCompose", name, "image-installer").Return(installerID, nil).Times(retries)
		composer.On("StartOSTreeCompose", name, "edge-commit", name, parent).Return(commitID, nil).Times(retries)
		composer.On("WaitForCompose", mock.Anything, commitID, mock.Anything, mock.Anything).Return(errors.New("commit waiting failed")).Times(retries)
		// Even though waiting failed, calls to save metadata and logs are expected to gather debug info
		composer.On("SaveComposeMetadata", commitID, mock.Anything).Return(nil).Times(retries)
		composer.On("SaveComposeLogs", commitID, mock.Anything).Return(nil).Times(retries)
		events.On("AddEvent", mock.MatchedBy(EventShouldBe(name, "commit", &testutil.FailedEvent{}))).Once()

		// installer
		utilProxy.On("PathExistsAndIsNotEmpty", installerDest).Return(false, nil)
		composer.On("StartCompose", name, "image-installer").Return(installerID, nil)
		composer.On("WaitForCompose", mock.Anything, installerID, mock.Anything, mock.Anything).Return(errors.New("installer waiting failed")).Times(retries)
		// Even though waiting failed, calls to save metadata and logs are expected to gather debug info
		composer.On("SaveComposeMetadata", installerID, mock.Anything).Return(nil).Times(retries)
		composer.On("SaveComposeLogs", installerID, mock.Anything).Return(nil).Times(retries)
		events.On("AddEvent", mock.MatchedBy(EventShouldBe(name, "installer", &testutil.FailedEvent{}))).Once()

		err := blueprint.Execute(context.TODO(), opts)
		assert.Error(t, err)
		// Verify that errors are joined and returned together
		assert.ErrorContains(t, err, "commit waiting failed")
		assert.ErrorContains(t, err, "installer waiting failed")

		composer.AssertExpectations(t)
		events.AssertExpectations(t)
	})

	t.Run("Waiting for compose, and saving metadata and logs fails", func(t *testing.T) {
		composer := &mocks.ComposerMock{}
		events := &mocks.EventManagerMock{}
		ostree := &mocks.OstreeMock{}
		utilProxy := &utilProxyMock{}
		opts := &Opts{Composer: composer, Events: events, Ostree: ostree, Utils: utilProxy,
			Force:         false,
			Retries:       retries,
			RetryInterval: time.Millisecond,
			Paths:         &testutil.Paths{VMStorageDir: "storage-dir/"},
		}

		// commit
		ostree.On("DoesRefExists", name).Return(false, nil)
		composer.On("StartOSTreeCompose", name, "edge-commit", name, parent).Return(commitID, nil).Times(retries)
		composer.On("WaitForCompose", mock.Anything, commitID, mock.Anything, mock.Anything).Return(errors.New("commit waiting failed")).Times(retries)
		// Even though waiting failed, calls to save metadata and logs are expected to gather debug info
		composer.On("SaveComposeMetadata", commitID, mock.Anything).Return(errors.New("commit saving metadata failed")).Times(retries)
		composer.On("SaveComposeLogs", commitID, mock.Anything).Return(errors.New("commit saving logs failed")).Times(retries)
		events.On("AddEvent", mock.MatchedBy(EventShouldBe(name, "commit", &testutil.FailedEvent{}))).Once()

		// installer
		utilProxy.On("PathExistsAndIsNotEmpty", installerDest).Return(false, nil)
		composer.On("StartCompose", name, "image-installer").Return(installerID, nil)
		composer.On("WaitForCompose", mock.Anything, installerID, mock.Anything, mock.Anything).Return(errors.New("installer waiting failed")).Times(retries)
		// Even though waiting failed, calls to save metadata and logs are expected to gather debug info
		composer.On("SaveComposeMetadata", installerID, mock.Anything).Return(errors.New("installer saving metadata failed")).Times(retries)
		composer.On("SaveComposeLogs", installerID, mock.Anything).Return(errors.New("installer saving logs failed")).Times(retries)
		events.On("AddEvent", mock.MatchedBy(EventShouldBe(name, "installer", &testutil.FailedEvent{}))).Once()

		err := blueprint.Execute(context.TODO(), opts)
		assert.Error(t, err)
		// Verify that errors are joined and returned together
		assert.ErrorContains(t, err, "commit saving metadata failed")
		assert.ErrorContains(t, err, "commit saving logs failed")
		assert.ErrorContains(t, err, "commit waiting failed")
		assert.ErrorContains(t, err, "installer saving metadata failed")
		assert.ErrorContains(t, err, "installer saving logs failed")
		assert.ErrorContains(t, err, "installer waiting failed")

		composer.AssertExpectations(t)
		events.AssertExpectations(t)
	})

}

func EventShouldBe(expectedName string, expectedClass string, expectedType testutil.IEvent) func(testutil.IEvent) bool {
	return func(ev testutil.IEvent) bool {
		if ev.GetName() != expectedName {
			return false
		}

		if expectedClass != "" && ev.GetClass() != expectedClass {
			return false
		}

		if reflect.TypeOf(ev) != reflect.TypeOf(expectedType) {
			return false
		}

		return true
	}
}

var _ UtilProxy = (*utilProxyMock)(nil)

type utilProxyMock struct {
	mock.Mock
}

func (u *utilProxyMock) Rename(oldpath string, newpath string) error {
	args := u.Called(oldpath, newpath)
	return args.Error(0)
}

func (u *utilProxyMock) PathExistsAndIsNotEmpty(path string) (bool, error) {
	args := u.Called(path)
	return args.Bool(0), args.Error(1)
}
