package sources

import (
	"fmt"
	"reflect"
	"testing"
	"testing/fstest"

	"github.com/openshift/microshift/test/pkg/compose/templatingdata"
	"github.com/openshift/microshift/test/pkg/testutil"
	"github.com/openshift/microshift/test/pkg/testutil/mocks"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_SuccessfulFlow(t *testing.T) {
	sourcesFS := fstest.MapFS{
		"microshift-base.toml": &fstest.MapFile{
			Data: []byte(`id = "microshift-base"
name = "MicroShift {{ .Base.Repository }} Repo"
type = "yum-baseurl"
url = "file://{{ .Base.Repository }}/"
check_gpg = false
check_ssl = false
system = false`)},

		"microshift-crel.toml": &fstest.MapFile{
			Data: []byte(`{{- if hasPrefix .Current.Repository "http" -}}
id = "microshift-crel"
name = "Repository with already existing RPMs for current release"
type = "yum-baseurl"
url = "{{ .Current.Repository }}"
check_gpg = false
check_ssl = true
system = false
{{- end -}}`)},

		"microshift-external.toml": &fstest.MapFile{
			Data: []byte(`{{- if .External.Repository -}}
id = "microshift-external"
name = "Repository with externally supplied RPMs"
type = "yum-baseurl"
url = "{{ .External.Repository }}"
check_gpg = false
check_ssl = true
system = false
{{- end -}}`)},

		"microshift-local.toml": &fstest.MapFile{
			Data: []byte(`id = "microshift-local"
name = "MicroShift Local Repo"
type = "yum-baseurl"
url = "file://{{ .Source.Repository }}/"
check_gpg = false
check_ssl = false
system = false`)},

		"microshift-prel.toml": &fstest.MapFile{
			Data: []byte(`{{- if hasPrefix .Previous.Repository "http" -}}
id = "microshift-prel"
name = "Repository with RPMs for previous release"
type = "yum-baseurl"
url = "{{ .Previous.Repository }}"
check_gpg = false
check_ssl = true
system = false
{{- end -}}`)},

		"rhocp-y.toml": &fstest.MapFile{
			Data: []byte(`{{- if .RHOCPMinorY -}}
id = "rhocp-y"
name = "Red Hat OpenShift Container Platform 4.{{ .RHOCPMinorY }} for RHEL 9"
type = "yum-baseurl"
url = "https://cdn.redhat.com/content/dist/layered/rhel9/{{ .Arch }}/rhocp/4.{{ .RHOCPMinorY }}/os"
check_gpg = true
check_ssl = true
system = false
rhsm = true
{{- end -}}`)},

		"rhocp-y1.toml": &fstest.MapFile{
			Data: []byte(`{{- if .RHOCPMinorY1 -}}
id = "rhocp-y1"
name = "Red Hat OpenShift Container Platform 4.{{ .RHOCPMinorY1 }} for RHEL 9"
type = "yum-baseurl"
url = "https://cdn.redhat.com/content/dist/layered/rhel9/{{ .Arch }}/rhocp/4.{{ .RHOCPMinorY1 }}/os"
check_gpg = true
check_ssl = true
system = false
rhsm = true
{{- end -}}`)},
		"rhocp-y2.toml": &fstest.MapFile{
			Data: []byte(`{{- if .RHOCPMinorY2 -}}
id = "rhocp-y2"
name = "Red Hat OpenShift Container Platform 4.{{ .RHOCPMinorY2 }} for RHEL 9"
type = "yum-baseurl"
url = "https://cdn.redhat.com/content/dist/layered/rhel9/{{ .Arch }}/rhocp/4.{{ .RHOCPMinorY2 }}/os"
check_gpg = true
check_ssl = true
system = false
rhsm = true
{{- end -}}`)},
	}

	composer := new(mocks.ComposerMock)
	events := new(mocks.EventManagerMock)

	baseRepo := "/path/to/base/rpms"
	crelRepo := "http://fake-repository.com"
	prevRepo := "rhocp-4.15"
	srcRepo := "/home/microshift/_output/local"
	extRepo := "/some/path"
	arch := "x86_64"
	rhocpY := 0
	rhocpY1 := 15
	rhocpY2 := 14

	opts := &SourceConfigurerOpts{
		Composer: composer,
		TplData: &templatingdata.TemplatingData{
			Arch:         arch,
			RHOCPMinorY:  rhocpY,
			RHOCPMinorY1: rhocpY1,
			RHOCPMinorY2: rhocpY2,
			Current: templatingdata.Release{
				Repository: crelRepo,
			},
			Previous: templatingdata.Release{
				Repository: prevRepo, // because it doesn't start with `http`, the rendered Source will be empty
			},
			Base: templatingdata.Release{
				Repository: baseRepo,
			},
			Source: templatingdata.Release{
				Repository: srcRepo,
			},
			External: templatingdata.Release{
				// non-empty value will cause microshift-ext.toml to be templated successfully and added to the composer
				Repository: extRepo,
			},
		},
		Events:    events,
		SourcesFS: sourcesFS,
	}

	getRHCOPUrl := func(minor int) string {
		return fmt.Sprintf("https://cdn.redhat.com/content/dist/layered/rhel9/%s/rhocp/4.%d/os", arch, minor)
	}

	composer.On("ListSources").Return([]string{"rhocp-y", "rhocp-y2"}, nil)

	// microshift-base
	composer.On("AddSource",
		mock.MatchedBy(And(
			IdShouldBe(t, "microshift-base"),
			UrlShouldBe(t, fmt.Sprintf("file://%s/", baseRepo)),
		))).Return(nil).Once()
	events.On("AddEvent", mock.MatchedBy(EventShouldBe("microshift-base", &testutil.Event{}))).Once()

	// microshift-crel
	composer.On("AddSource",
		mock.MatchedBy(And(
			IdShouldBe(t, "microshift-crel"),
			UrlShouldBe(t, crelRepo),
		))).Return(nil).Once()
	events.On("AddEvent", mock.MatchedBy(EventShouldBe("microshift-crel", &testutil.Event{}))).Once()

	// microshift-external
	composer.On("AddSource",
		mock.MatchedBy(And(
			IdShouldBe(t, "microshift-external"),
			UrlShouldBe(t, extRepo),
		))).Return(nil).Once()
	events.On("AddEvent", mock.MatchedBy(EventShouldBe("microshift-external", &testutil.Event{}))).Once()

	// microshift-local
	composer.On("AddSource",
		mock.MatchedBy(And(
			IdShouldBe(t, "microshift-local"),
			UrlShouldBe(t, fmt.Sprintf("file://%s/", srcRepo)),
		))).Return(nil).Once()
	events.On("AddEvent", mock.MatchedBy(EventShouldBe("microshift-local", &testutil.Event{}))).Once()

	// microshift-prel - should not cause AddSource to be called, but AddEvent with SkippedEvent should be called
	events.On("AddEvent", mock.MatchedBy(EventShouldBe("microshift-prel", &testutil.SkippedEvent{}))).Once()

	// rhocp-y
	// Present in ListSources(), but templated to empty string because of the .RHOCPY=0: Delete, don't Add
	composer.On("DeleteSource", "rhocp-y").Return(nil).Once()

	// rhocp-y1
	composer.On("AddSource",
		mock.MatchedBy(And(
			IdShouldBe(t, "rhocp-y1"),
			UrlShouldBe(t, getRHCOPUrl(rhocpY1)),
		))).Return(nil).Once()
	events.On("AddEvent", mock.MatchedBy(EventShouldBe("rhocp-y1", &testutil.Event{}))).Once()

	// rhocp-y2
	// Although it's already present in the composer (present in slice returned from ListSources()), it's not deleted, just updated with AddEvent call
	composer.On("AddSource",
		mock.MatchedBy(And(
			IdShouldBe(t, "rhocp-y2"),
			UrlShouldBe(t, getRHCOPUrl(rhocpY2)),
		))).Return(nil).Once()
	events.On("AddEvent", mock.MatchedBy(EventShouldBe("rhocp-y2", &testutil.Event{}))).Once()

	err := (&SourceConfigurer{Opts: opts}).ConfigureSources()
	assert.NoError(t, err)
	composer.AssertExpectations(t)
	events.AssertExpectations(t)
}

// Verify that error from AddSource() is both: registered in the Events and propagated back to the caller.
func Test_AddSourceFailed(t *testing.T) {
	sourcesFS := fstest.MapFS{
		"microshift-base.toml": &fstest.MapFile{
			Data: []byte(`id = "microshift-base"
name = "MicroShift {{ .Base.Repository }} Repo"
type = "yum-baseurl"
url = "file://{{ .Base.Repository }}/"
check_gpg = false
check_ssl = false
system = false`)},
		"microshift-crel.toml": &fstest.MapFile{
			Data: []byte(`{{- if hasPrefix .Current.Repository "http" -}}
id = "microshift-crel"
name = "Repository with already existing RPMs for current release"
type = "yum-baseurl"
url = "{{ .Current.Repository }}"
check_gpg = false
check_ssl = true
system = false
{{- end -}}`)},
	}

	composer := new(mocks.ComposerMock)
	events := new(mocks.EventManagerMock)

	baseRepo := "/path/to/base/rpms"
	crelRepo := "http://fake-repository.com"

	opts := &SourceConfigurerOpts{
		Composer: composer,
		TplData: &templatingdata.TemplatingData{
			Base: templatingdata.Release{
				Repository: baseRepo,
			},
			Current: templatingdata.Release{
				Repository: crelRepo,
			},
		},
		Events:    events,
		SourcesFS: sourcesFS,
	}

	composer.On("ListSources").Return([]string{"rhocp-y", "rhocp-y2"}, nil)

	// microshift-base
	composer.On("AddSource",
		mock.MatchedBy(And(
			IdShouldBe(t, "microshift-base"),
			UrlShouldBe(t, fmt.Sprintf("file://%s/", baseRepo)),
		))).Return(fmt.Errorf("some error")).Once()
	events.On("AddEvent", mock.MatchedBy(EventShouldBe("microshift-base", &testutil.FailedEvent{}))).Once()

	// microshift-crel
	composer.On("AddSource",
		mock.MatchedBy(And(
			IdShouldBe(t, "microshift-crel"),
			UrlShouldBe(t, crelRepo),
		))).Return(fmt.Errorf("some other error")).Once()
	events.On("AddEvent", mock.MatchedBy(EventShouldBe("microshift-crel", &testutil.FailedEvent{}))).Once()

	err := (&SourceConfigurer{Opts: opts}).ConfigureSources()
	assert.Error(t, err)
	assert.ErrorContains(t, err, "some error")
	assert.ErrorContains(t, err, "some other error")
	composer.AssertExpectations(t)
	events.AssertExpectations(t)
}

func EventShouldBe(expectedName string, expectedType testutil.IEvent) func(testutil.IEvent) bool {
	return func(ev testutil.IEvent) bool {
		if ev.GetName() != expectedName {
			return false
		}

		if reflect.TypeOf(ev) != reflect.TypeOf(expectedType) {
			return false
		}

		return true
	}
}

func And(preds ...func(string) bool) func(string) bool {
	return func(s string) bool {
		for _, pred := range preds {
			if !pred(s) {
				return false
			}
		}
		return true
	}
}

func IdShouldBe(t *testing.T, id string) func(string) bool {
	return func(tomlSource string) bool {
		src := Source{}
		_, err := toml.Decode(tomlSource, &src)
		assert.NoError(t, err)
		return src.Id == id
	}
}

func UrlShouldBe(t *testing.T, url string) func(string) bool {
	return func(tomlSource string) bool {
		src := Source{}
		_, err := toml.Decode(tomlSource, &src)
		assert.NoError(t, err)
		return src.Url == url
	}
}

// Source is a brief representation of composer's Source.
// Added here to not import more dependencies as we know exactly what fields we use and care about in tests.
type Source struct {
	Id   string
	Name string
	Url  string
}
