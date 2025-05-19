/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package plugin

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/tools/cache"
	credentialproviderapi "k8s.io/kubelet/pkg/apis/credentialprovider"
	credentialproviderv1 "k8s.io/kubelet/pkg/apis/credentialprovider/v1"
	credentialproviderv1alpha1 "k8s.io/kubelet/pkg/apis/credentialprovider/v1alpha1"
	credentialproviderv1beta1 "k8s.io/kubelet/pkg/apis/credentialprovider/v1beta1"
	"k8s.io/kubernetes/pkg/credentialprovider"
	kubeletconfig "k8s.io/kubernetes/pkg/kubelet/apis/config"
	"k8s.io/utils/clock"
	testingclock "k8s.io/utils/clock/testing"
)

type fakeExecPlugin struct {
	cacheKeyType  credentialproviderapi.PluginCacheKeyType
	cacheDuration time.Duration

	auth map[string]credentialproviderapi.AuthConfig
}

func (f *fakeExecPlugin) ExecPlugin(ctx context.Context, image string) (*credentialproviderapi.CredentialProviderResponse, error) {
	return &credentialproviderapi.CredentialProviderResponse{
		CacheKeyType: f.cacheKeyType,
		CacheDuration: &metav1.Duration{
			Duration: f.cacheDuration,
		},
		Auth: f.auth,
	}, nil
}

func Test_Provide(t *testing.T) {
	tclock := clock.RealClock{}
	testcases := []struct {
		name           string
		pluginProvider *pluginProvider
		image          string
		dockerconfig   credentialprovider.DockerConfig
	}{
		{
			name: "exact image match, with Registry cache key",
			pluginProvider: &pluginProvider{
				clock:          tclock,
				lastCachePurge: tclock.Now(),
				matchImages:    []string{"test.registry.io"},
				cache:          cache.NewExpirationStore(cacheKeyFunc, &cacheExpirationPolicy{clock: tclock}),
				plugin: &fakeExecPlugin{
					cacheKeyType: credentialproviderapi.RegistryPluginCacheKeyType,
					auth: map[string]credentialproviderapi.AuthConfig{
						"test.registry.io": {
							Username: "user",
							Password: "password",
						},
					},
				},
			},
			image: "test.registry.io/foo/bar",
			dockerconfig: credentialprovider.DockerConfig{
				"test.registry.io": credentialprovider.DockerConfigEntry{
					Username: "user",
					Password: "password",
				},
			},
		},
		{
			name: "exact image match, with Image cache key",
			pluginProvider: &pluginProvider{
				clock:          tclock,
				lastCachePurge: tclock.Now(),
				matchImages:    []string{"test.registry.io/foo/bar"},
				cache:          cache.NewExpirationStore(cacheKeyFunc, &cacheExpirationPolicy{clock: tclock}),
				plugin: &fakeExecPlugin{
					cacheKeyType: credentialproviderapi.ImagePluginCacheKeyType,
					auth: map[string]credentialproviderapi.AuthConfig{
						"test.registry.io/foo/bar": {
							Username: "user",
							Password: "password",
						},
					},
				},
			},
			image: "test.registry.io/foo/bar",
			dockerconfig: credentialprovider.DockerConfig{
				"test.registry.io/foo/bar": credentialprovider.DockerConfigEntry{
					Username: "user",
					Password: "password",
				},
			},
		},
		{
			name: "exact image match, with Global cache key",
			pluginProvider: &pluginProvider{
				clock:          tclock,
				lastCachePurge: tclock.Now(),
				matchImages:    []string{"test.registry.io"},
				cache:          cache.NewExpirationStore(cacheKeyFunc, &cacheExpirationPolicy{clock: tclock}),
				plugin: &fakeExecPlugin{
					cacheKeyType: credentialproviderapi.GlobalPluginCacheKeyType,
					auth: map[string]credentialproviderapi.AuthConfig{
						"test.registry.io": {
							Username: "user",
							Password: "password",
						},
					},
				},
			},
			image: "test.registry.io",
			dockerconfig: credentialprovider.DockerConfig{
				"test.registry.io": credentialprovider.DockerConfigEntry{
					Username: "user",
					Password: "password",
				},
			},
		},
		{
			name: "wild card image match, with Registry cache key",
			pluginProvider: &pluginProvider{
				clock:          tclock,
				lastCachePurge: tclock.Now(),
				matchImages:    []string{"*.registry.io:8080"},
				cache:          cache.NewExpirationStore(cacheKeyFunc, &cacheExpirationPolicy{clock: tclock}),
				plugin: &fakeExecPlugin{
					cacheKeyType: credentialproviderapi.RegistryPluginCacheKeyType,
					auth: map[string]credentialproviderapi.AuthConfig{
						"*.registry.io:8080": {
							Username: "user",
							Password: "password",
						},
					},
				},
			},
			image: "test.registry.io:8080/foo",
			dockerconfig: credentialprovider.DockerConfig{
				"*.registry.io:8080": credentialprovider.DockerConfigEntry{
					Username: "user",
					Password: "password",
				},
			},
		},
		{
			name: "wild card image match, with Image cache key",
			pluginProvider: &pluginProvider{
				clock:          tclock,
				lastCachePurge: tclock.Now(),
				matchImages:    []string{"*.*.registry.io"},
				cache:          cache.NewExpirationStore(cacheKeyFunc, &cacheExpirationPolicy{clock: tclock}),
				plugin: &fakeExecPlugin{
					cacheKeyType: credentialproviderapi.ImagePluginCacheKeyType,
					auth: map[string]credentialproviderapi.AuthConfig{
						"*.*.registry.io": {
							Username: "user",
							Password: "password",
						},
					},
				},
			},
			image: "foo.bar.registry.io/foo/bar",
			dockerconfig: credentialprovider.DockerConfig{
				"*.*.registry.io": credentialprovider.DockerConfigEntry{
					Username: "user",
					Password: "password",
				},
			},
		},
		{
			name: "wild card image match, with Global cache key",
			pluginProvider: &pluginProvider{
				clock:          tclock,
				lastCachePurge: tclock.Now(),
				matchImages:    []string{"*.registry.io"},
				cache:          cache.NewExpirationStore(cacheKeyFunc, &cacheExpirationPolicy{clock: tclock}),
				plugin: &fakeExecPlugin{
					cacheKeyType: credentialproviderapi.GlobalPluginCacheKeyType,
					auth: map[string]credentialproviderapi.AuthConfig{
						"*.registry.io": {
							Username: "user",
							Password: "password",
						},
					},
				},
			},
			image: "test.registry.io",
			dockerconfig: credentialprovider.DockerConfig{
				"*.registry.io": credentialprovider.DockerConfigEntry{
					Username: "user",
					Password: "password",
				},
			},
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()
			dockerconfig := testcase.pluginProvider.Provide(testcase.image)
			if !reflect.DeepEqual(dockerconfig, testcase.dockerconfig) {
				t.Logf("actual docker config: %v", dockerconfig)
				t.Logf("expected docker config: %v", testcase.dockerconfig)
				t.Error("unexpected docker config")
			}
		})
	}
}

// This test calls Provide in parallel for different registries and images
// The purpose of this is to detect any race conditions while cache rw.
func Test_ProvideParallel(t *testing.T) {
	tclock := clock.RealClock{}

	testcases := []struct {
		name     string
		registry string
	}{
		{
			name:     "provide for registry 1",
			registry: "test1.registry.io",
		},
		{
			name:     "provide for registry 2",
			registry: "test2.registry.io",
		},
		{
			name:     "provide for registry 3",
			registry: "test3.registry.io",
		},
		{
			name:     "provide for registry 4",
			registry: "test4.registry.io",
		},
	}

	pluginProvider := &pluginProvider{
		clock:          tclock,
		lastCachePurge: tclock.Now(),
		matchImages:    []string{"test1.registry.io", "test2.registry.io", "test3.registry.io", "test4.registry.io"},
		cache:          cache.NewExpirationStore(cacheKeyFunc, &cacheExpirationPolicy{clock: tclock}),
		plugin: &fakeExecPlugin{
			cacheDuration: time.Minute * 1,
			cacheKeyType:  credentialproviderapi.RegistryPluginCacheKeyType,
			auth: map[string]credentialproviderapi.AuthConfig{
				"test.registry.io": {
					Username: "user",
					Password: "password",
				},
			},
		},
	}

	dockerconfig := credentialprovider.DockerConfig{
		"test.registry.io": credentialprovider.DockerConfigEntry{
			Username: "user",
			Password: "password",
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()
			var wg sync.WaitGroup
			wg.Add(5)

			for i := 0; i < 5; i++ {
				go func(w *sync.WaitGroup) {
					image := fmt.Sprintf(testcase.registry+"/%s", rand.String(5))
					dockerconfigResponse := pluginProvider.Provide(image)
					if !reflect.DeepEqual(dockerconfigResponse, dockerconfig) {
						t.Logf("actual docker config: %v", dockerconfigResponse)
						t.Logf("expected docker config: %v", dockerconfig)
						t.Error("unexpected docker config")
					}
					w.Done()
				}(&wg)
			}
			wg.Wait()

		})
	}
}

func Test_getCachedCredentials(t *testing.T) {
	fakeClock := testingclock.NewFakeClock(time.Now())
	p := &pluginProvider{
		clock:          fakeClock,
		lastCachePurge: fakeClock.Now(),
		cache:          cache.NewExpirationStore(cacheKeyFunc, &cacheExpirationPolicy{clock: fakeClock}),
		plugin:         &fakeExecPlugin{},
	}

	testcases := []struct {
		name             string
		step             time.Duration
		cacheEntry       cacheEntry
		expectedResponse credentialprovider.DockerConfig
		keyLength        int
		getKey           string
	}{
		{
			name:      "It should return not expired credential",
			step:      1 * time.Second,
			keyLength: 1,
			getKey:    "image1",
			expectedResponse: map[string]credentialprovider.DockerConfigEntry{
				"image1": {
					Username: "user1",
					Password: "pass1",
				},
			},
			cacheEntry: cacheEntry{
				key:       "image1",
				expiresAt: fakeClock.Now().Add(1 * time.Minute),
				credentials: map[string]credentialprovider.DockerConfigEntry{
					"image1": {
						Username: "user1",
						Password: "pass1",
					},
				},
			},
		},

		{
			name:      "It should not return expired credential",
			step:      2 * time.Minute,
			getKey:    "image2",
			keyLength: 1,
			cacheEntry: cacheEntry{
				key:       "image2",
				expiresAt: fakeClock.Now(),
				credentials: map[string]credentialprovider.DockerConfigEntry{
					"image2": {
						Username: "user2",
						Password: "pass2",
					},
				},
			},
		},

		{
			name:      "It should delete expired credential during purge",
			step:      18 * time.Minute,
			keyLength: 0,
			// while get call for random, cache purge will be called and it will delete expired
			// image3 credentials. We cannot use image3 as getKey here, as it will get deleted during
			// get only, we will not be able verify the purge call.
			getKey: "random",
			cacheEntry: cacheEntry{
				key:       "image3",
				expiresAt: fakeClock.Now().Add(2 * time.Minute),
				credentials: map[string]credentialprovider.DockerConfigEntry{
					"image3": {
						Username: "user3",
						Password: "pass3",
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p.cache.Add(&tc.cacheEntry)
			fakeClock.Step(tc.step)

			// getCachedCredentials returns unexpired credentials.
			res, _, err := p.getCachedCredentials(tc.getKey)
			if err != nil {
				t.Errorf("Unexpected error %v", err)
			}
			if !reflect.DeepEqual(res, tc.expectedResponse) {
				t.Logf("response %v", res)
				t.Logf("expected response %v", tc.expectedResponse)
				t.Errorf("Unexpected response")
			}

			// Listkeys returns all the keys present in cache including expired keys.
			if len(p.cache.ListKeys()) != tc.keyLength {
				t.Errorf("Unexpected cache key length")
			}
		})
	}
}

func Test_encodeRequest(t *testing.T) {
	testcases := []struct {
		name         string
		apiVersion   schema.GroupVersion
		request      *credentialproviderapi.CredentialProviderRequest
		expectedData []byte
		expectedErr  bool
	}{
		{
			name:       "successful with v1alpha1",
			apiVersion: credentialproviderv1alpha1.SchemeGroupVersion,
			request: &credentialproviderapi.CredentialProviderRequest{
				Image: "test.registry.io/foobar",
			},
			expectedData: []byte(`{"kind":"CredentialProviderRequest","apiVersion":"credentialprovider.kubelet.k8s.io/v1alpha1","image":"test.registry.io/foobar"}
`),
			expectedErr: false,
		},
		{
			name:       "successful with v1beta1",
			apiVersion: credentialproviderv1beta1.SchemeGroupVersion,
			request: &credentialproviderapi.CredentialProviderRequest{
				Image: "test.registry.io/foobar",
			},
			expectedData: []byte(`{"kind":"CredentialProviderRequest","apiVersion":"credentialprovider.kubelet.k8s.io/v1beta1","image":"test.registry.io/foobar"}
`),
			expectedErr: false,
		},
		{
			name:       "successful with v1",
			apiVersion: credentialproviderv1.SchemeGroupVersion,
			request: &credentialproviderapi.CredentialProviderRequest{
				Image: "test.registry.io/foobar",
			},
			expectedData: []byte(`{"kind":"CredentialProviderRequest","apiVersion":"credentialprovider.kubelet.k8s.io/v1","image":"test.registry.io/foobar"}
`),
			expectedErr: false,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			mediaType := "application/json"
			info, ok := runtime.SerializerInfoForMediaType(codecs.SupportedMediaTypes(), mediaType)
			if !ok {
				t.Fatalf("unsupported media type: %s", mediaType)
			}

			e := &execPlugin{
				encoder: codecs.EncoderForVersion(info.Serializer, testcase.apiVersion),
			}

			data, err := e.encodeRequest(testcase.request)
			if err != nil && !testcase.expectedErr {
				t.Fatalf("unexpected error: %v", err)
			}

			if err == nil && testcase.expectedErr {
				t.Fatalf("expected error %v but got nil", testcase.expectedErr)
			}

			if !reflect.DeepEqual(data, testcase.expectedData) {
				t.Errorf("actual encoded data: %v", string(data))
				t.Errorf("expected encoded data: %v", string(testcase.expectedData))
				t.Errorf("unexpected encoded response")
			}
		})
	}
}

func Test_decodeResponse(t *testing.T) {
	testcases := []struct {
		name             string
		data             []byte
		expectedResponse *credentialproviderapi.CredentialProviderResponse
		expectedErr      bool
	}{
		{
			name: "success with v1",
			data: []byte(`{"kind":"CredentialProviderResponse","apiVersion":"credentialprovider.kubelet.k8s.io/v1","cacheKeyType":"Registry","cacheDuration":"1m","auth":{"*.registry.io":{"username":"user","password":"password"}}}`),
			expectedResponse: &credentialproviderapi.CredentialProviderResponse{
				CacheKeyType: credentialproviderapi.RegistryPluginCacheKeyType,
				CacheDuration: &metav1.Duration{
					Duration: time.Minute,
				},
				Auth: map[string]credentialproviderapi.AuthConfig{
					"*.registry.io": {
						Username: "user",
						Password: "password",
					},
				},
			},
			expectedErr: false,
		},
		{
			name: "success with v1beta1",
			data: []byte(`{"kind":"CredentialProviderResponse","apiVersion":"credentialprovider.kubelet.k8s.io/v1beta1","cacheKeyType":"Registry","cacheDuration":"1m","auth":{"*.registry.io":{"username":"user","password":"password"}}}`),
			expectedResponse: &credentialproviderapi.CredentialProviderResponse{
				CacheKeyType: credentialproviderapi.RegistryPluginCacheKeyType,
				CacheDuration: &metav1.Duration{
					Duration: time.Minute,
				},
				Auth: map[string]credentialproviderapi.AuthConfig{
					"*.registry.io": {
						Username: "user",
						Password: "password",
					},
				},
			},
			expectedErr: false,
		},
		{
			name: "success with v1alpha1",
			data: []byte(`{"kind":"CredentialProviderResponse","apiVersion":"credentialprovider.kubelet.k8s.io/v1alpha1","cacheKeyType":"Registry","cacheDuration":"1m","auth":{"*.registry.io":{"username":"user","password":"password"}}}`),
			expectedResponse: &credentialproviderapi.CredentialProviderResponse{
				CacheKeyType: credentialproviderapi.RegistryPluginCacheKeyType,
				CacheDuration: &metav1.Duration{
					Duration: time.Minute,
				},
				Auth: map[string]credentialproviderapi.AuthConfig{
					"*.registry.io": {
						Username: "user",
						Password: "password",
					},
				},
			},
			expectedErr: false,
		},
		{
			name:             "wrong Kind",
			data:             []byte(`{"kind":"WrongKind","apiVersion":"credentialprovider.kubelet.k8s.io/v1beta1","cacheKeyType":"Registry","cacheDuration":"1m","auth":{"*.registry.io":{"username":"user","password":"password"}}}`),
			expectedResponse: nil,
			expectedErr:      true,
		},
		{
			name:             "wrong Group",
			data:             []byte(`{"kind":"CredentialProviderResponse","apiVersion":"foobar.kubelet.k8s.io/v1beta1","cacheKeyType":"Registry","cacheDuration":"1m","auth":{"*.registry.io":{"username":"user","password":"password"}}}`),
			expectedResponse: nil,
			expectedErr:      true,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			e := &execPlugin{}

			decodedResponse, err := e.decodeResponse(testcase.data)
			if err != nil && !testcase.expectedErr {
				t.Fatalf("unexpected error: %v", err)
			}

			if err == nil && testcase.expectedErr {
				t.Fatalf("expected error %v but not nil", testcase.expectedErr)
			}

			if !reflect.DeepEqual(decodedResponse, testcase.expectedResponse) {
				t.Logf("actual decoded response: %#v", decodedResponse)
				t.Logf("expected decoded response: %#v", testcase.expectedResponse)
				t.Errorf("unexpected decoded response")
			}
		})
	}
}

func Test_RegistryCacheKeyType(t *testing.T) {
	tclock := clock.RealClock{}
	pluginProvider := &pluginProvider{
		clock:          tclock,
		lastCachePurge: tclock.Now(),
		matchImages:    []string{"*.registry.io"},
		cache:          cache.NewExpirationStore(cacheKeyFunc, &cacheExpirationPolicy{clock: tclock}),
		plugin: &fakeExecPlugin{
			cacheKeyType:  credentialproviderapi.RegistryPluginCacheKeyType,
			cacheDuration: time.Hour,
			auth: map[string]credentialproviderapi.AuthConfig{
				"*.registry.io": {
					Username: "user",
					Password: "password",
				},
			},
		},
	}

	expectedDockerConfig := credentialprovider.DockerConfig{
		"*.registry.io": credentialprovider.DockerConfigEntry{
			Username: "user",
			Password: "password",
		},
	}

	dockerConfig := pluginProvider.Provide("test.registry.io/foo/bar")
	if !reflect.DeepEqual(dockerConfig, expectedDockerConfig) {
		t.Logf("actual docker config: %v", dockerConfig)
		t.Logf("expected docker config: %v", expectedDockerConfig)
		t.Fatal("unexpected docker config")
	}

	expectedCacheKeys := []string{"test.registry.io"}
	cacheKeys := pluginProvider.cache.ListKeys()

	if !reflect.DeepEqual(cacheKeys, expectedCacheKeys) {
		t.Logf("actual cache keys: %v", cacheKeys)
		t.Logf("expected cache keys: %v", expectedCacheKeys)
		t.Error("unexpected cache keys")
	}

	// nil out the exec plugin, this will test whether credentialproviderapi are fetched
	// from cache, otherwise Provider should panic
	pluginProvider.plugin = nil
	dockerConfig = pluginProvider.Provide("test.registry.io/foo/bar")
	if !reflect.DeepEqual(dockerConfig, expectedDockerConfig) {
		t.Logf("actual docker config: %v", dockerConfig)
		t.Logf("expected docker config: %v", expectedDockerConfig)
		t.Fatal("unexpected docker config")
	}
}

func Test_ImageCacheKeyType(t *testing.T) {
	tclock := clock.RealClock{}
	pluginProvider := &pluginProvider{
		clock:          tclock,
		lastCachePurge: tclock.Now(),
		matchImages:    []string{"*.registry.io"},
		cache:          cache.NewExpirationStore(cacheKeyFunc, &cacheExpirationPolicy{clock: tclock}),
		plugin: &fakeExecPlugin{
			cacheKeyType:  credentialproviderapi.ImagePluginCacheKeyType,
			cacheDuration: time.Hour,
			auth: map[string]credentialproviderapi.AuthConfig{
				"*.registry.io": {
					Username: "user",
					Password: "password",
				},
			},
		},
	}

	expectedDockerConfig := credentialprovider.DockerConfig{
		"*.registry.io": credentialprovider.DockerConfigEntry{
			Username: "user",
			Password: "password",
		},
	}

	dockerConfig := pluginProvider.Provide("test.registry.io/foo/bar")
	if !reflect.DeepEqual(dockerConfig, expectedDockerConfig) {
		t.Logf("actual docker config: %v", dockerConfig)
		t.Logf("expected docker config: %v", expectedDockerConfig)
		t.Fatal("unexpected docker config")
	}

	expectedCacheKeys := []string{"test.registry.io/foo/bar"}
	cacheKeys := pluginProvider.cache.ListKeys()

	if !reflect.DeepEqual(cacheKeys, expectedCacheKeys) {
		t.Logf("actual cache keys: %v", cacheKeys)
		t.Logf("expected cache keys: %v", expectedCacheKeys)
		t.Error("unexpected cache keys")
	}

	// nil out the exec plugin, this will test whether credentialproviderapi are fetched
	// from cache, otherwise Provider should panic
	pluginProvider.plugin = nil
	dockerConfig = pluginProvider.Provide("test.registry.io/foo/bar")
	if !reflect.DeepEqual(dockerConfig, expectedDockerConfig) {
		t.Logf("actual docker config: %v", dockerConfig)
		t.Logf("expected docker config: %v", expectedDockerConfig)
		t.Fatal("unexpected docker config")
	}
}

func Test_GlobalCacheKeyType(t *testing.T) {
	tclock := clock.RealClock{}
	pluginProvider := &pluginProvider{
		clock:          tclock,
		lastCachePurge: tclock.Now(),
		matchImages:    []string{"*.registry.io"},
		cache:          cache.NewExpirationStore(cacheKeyFunc, &cacheExpirationPolicy{clock: tclock}),
		plugin: &fakeExecPlugin{
			cacheKeyType:  credentialproviderapi.GlobalPluginCacheKeyType,
			cacheDuration: time.Hour,
			auth: map[string]credentialproviderapi.AuthConfig{
				"*.registry.io": {
					Username: "user",
					Password: "password",
				},
			},
		},
	}

	expectedDockerConfig := credentialprovider.DockerConfig{
		"*.registry.io": credentialprovider.DockerConfigEntry{
			Username: "user",
			Password: "password",
		},
	}

	dockerConfig := pluginProvider.Provide("test.registry.io/foo/bar")
	if !reflect.DeepEqual(dockerConfig, expectedDockerConfig) {
		t.Logf("actual docker config: %v", dockerConfig)
		t.Logf("expected docker config: %v", expectedDockerConfig)
		t.Fatal("unexpected docker config")
	}

	expectedCacheKeys := []string{"global"}
	cacheKeys := pluginProvider.cache.ListKeys()

	if !reflect.DeepEqual(cacheKeys, expectedCacheKeys) {
		t.Logf("actual cache keys: %v", cacheKeys)
		t.Logf("expected cache keys: %v", expectedCacheKeys)
		t.Error("unexpected cache keys")
	}

	// nil out the exec plugin, this will test whether credentialproviderapi are fetched
	// from cache, otherwise Provider should panic
	pluginProvider.plugin = nil
	dockerConfig = pluginProvider.Provide("test.registry.io/foo/bar")
	if !reflect.DeepEqual(dockerConfig, expectedDockerConfig) {
		t.Logf("actual docker config: %v", dockerConfig)
		t.Logf("expected docker config: %v", expectedDockerConfig)
		t.Fatal("unexpected docker config")
	}
}

func Test_NoCacheResponse(t *testing.T) {
	tclock := clock.RealClock{}
	pluginProvider := &pluginProvider{
		clock:          tclock,
		lastCachePurge: tclock.Now(),
		matchImages:    []string{"*.registry.io"},
		cache:          cache.NewExpirationStore(cacheKeyFunc, &cacheExpirationPolicy{clock: tclock}),
		plugin: &fakeExecPlugin{
			cacheKeyType:  credentialproviderapi.GlobalPluginCacheKeyType,
			cacheDuration: 0, // no cache
			auth: map[string]credentialproviderapi.AuthConfig{
				"*.registry.io": {
					Username: "user",
					Password: "password",
				},
			},
		},
	}

	expectedDockerConfig := credentialprovider.DockerConfig{
		"*.registry.io": credentialprovider.DockerConfigEntry{
			Username: "user",
			Password: "password",
		},
	}

	dockerConfig := pluginProvider.Provide("test.registry.io/foo/bar")
	if !reflect.DeepEqual(dockerConfig, expectedDockerConfig) {
		t.Logf("actual docker config: %v", dockerConfig)
		t.Logf("expected docker config: %v", expectedDockerConfig)
		t.Fatal("unexpected docker config")
	}

	expectedCacheKeys := []string{}
	cacheKeys := pluginProvider.cache.ListKeys()
	if !reflect.DeepEqual(cacheKeys, expectedCacheKeys) {
		t.Logf("actual cache keys: %v", cacheKeys)
		t.Logf("expected cache keys: %v", expectedCacheKeys)
		t.Error("unexpected cache keys")
	}
}

func Test_ExecPluginEnvVars(t *testing.T) {
	testcases := []struct {
		name            string
		systemEnvVars   []string
		execPlugin      *execPlugin
		expectedEnvVars []string
	}{
		{
			name:          "positive append system env vars",
			systemEnvVars: []string{"HOME=/home/foo", "PATH=/usr/bin"},
			execPlugin: &execPlugin{
				envVars: []kubeletconfig.ExecEnvVar{
					{
						Name:  "SUPER_SECRET_STRONG_ACCESS_KEY",
						Value: "123456789",
					},
				},
			},
			expectedEnvVars: []string{
				"HOME=/home/foo",
				"PATH=/usr/bin",
				"SUPER_SECRET_STRONG_ACCESS_KEY=123456789",
			},
		},
		{
			name:          "positive no env vars provided in plugin",
			systemEnvVars: []string{"HOME=/home/foo", "PATH=/usr/bin"},
			execPlugin:    &execPlugin{},
			expectedEnvVars: []string{
				"HOME=/home/foo",
				"PATH=/usr/bin",
			},
		},
		{
			name: "positive no system env vars but env vars are provided in plugin",
			execPlugin: &execPlugin{
				envVars: []kubeletconfig.ExecEnvVar{
					{
						Name:  "SUPER_SECRET_STRONG_ACCESS_KEY",
						Value: "123456789",
					},
				},
			},
			expectedEnvVars: []string{
				"SUPER_SECRET_STRONG_ACCESS_KEY=123456789",
			},
		},
		{
			name:            "positive no system or plugin provided env vars",
			execPlugin:      &execPlugin{},
			expectedEnvVars: nil,
		},
		{
			name:          "positive plugin provided vars takes priority",
			systemEnvVars: []string{"HOME=/home/foo", "PATH=/usr/bin", "SUPER_SECRET_STRONG_ACCESS_KEY=1111"},
			execPlugin: &execPlugin{
				envVars: []kubeletconfig.ExecEnvVar{
					{
						Name:  "SUPER_SECRET_STRONG_ACCESS_KEY",
						Value: "123456789",
					},
				},
			},
			expectedEnvVars: []string{
				"HOME=/home/foo",
				"PATH=/usr/bin",
				"SUPER_SECRET_STRONG_ACCESS_KEY=1111",
				"SUPER_SECRET_STRONG_ACCESS_KEY=123456789",
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			testcase.execPlugin.environ = func() []string {
				return testcase.systemEnvVars
			}

			var configVars []string
			for _, envVar := range testcase.execPlugin.envVars {
				configVars = append(configVars, fmt.Sprintf("%s=%s", envVar.Name, envVar.Value))
			}
			merged := mergeEnvVars(testcase.systemEnvVars, configVars)

			err := validate(testcase.expectedEnvVars, merged)
			if err != nil {
				t.Logf("unexpecged error %v", err)
			}
		})
	}
}

func validate(expected, actual []string) error {
	if len(actual) != len(expected) {
		return fmt.Errorf("actual env var length [%d] and expected env var length [%d] don't match",
			len(actual), len(expected))
	}

	for i := range actual {
		if actual[i] != expected[i] {
			return fmt.Errorf("mismatch in expected env var %s and actual env var %s", actual[i], expected[i])
		}
	}

	return nil
}
