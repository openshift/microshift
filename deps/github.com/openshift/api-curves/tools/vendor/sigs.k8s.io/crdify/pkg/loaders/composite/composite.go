// Copyright 2025 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package composite

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// Loader is used to load a CustomResourceDefinition from a source location.
type Loader interface {
	// Load uses the provided context and URL to determine how to
	// source a CustomResourceDefinition.
	// Upon successful sourcing, a non-nil CustomResourceDefinition and a nil error should be returned.
	// Upon failed sourcing, a nil CustomResourceDefinition and a non-nil error should be returned.
	Load(context.Context, *url.URL) (*apiextensionsv1.CustomResourceDefinition, error)
}

// Composite is a utility type that is used to encapsulate
// the behavior of multiple loaders into a single implementation.
// It uses the scheme of a URL as the key for which encapsulated Loader
// to execute.
type Composite struct {
	// loaders is the set of loaders than can be used
	loaders map[string]Loader
}

// NewComposite creates a new Composite loader configured with the provided
// loaders.
func NewComposite(loaders map[string]Loader) *Composite {
	composite := &Composite{
		loaders: loaders,
	}

	return composite
}

// Load is used to source a CustomResourceDefinition using the provided context and source string.
// The source string is expected to be a parseable URL using Go's net/url.Parse() function.
// Depending on the scheme of the parsed URL, Load will call a nested Loader implementation
// to source the CustomResourceDefinition.
func (c *Composite) Load(ctx context.Context, location string) (*apiextensionsv1.CustomResourceDefinition, error) {
	locationURL, err := url.Parse(location)
	if err != nil {
		log.Fatalf("parsing source: %v", err)
	}

	loader, ok := c.loaders[locationURL.Scheme]
	if !ok {
		return nil, fmt.Errorf("%w : %q", errNoLoader, locationURL.Scheme)
	}

	crd, err := loader.Load(ctx, locationURL)
	if err != nil {
		return nil, fmt.Errorf("loading CustomResourceDefinition: %w", err)
	}

	return crd, nil
}

var errNoLoader = errors.New("no loader found for provided scheme")
