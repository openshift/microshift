/*
Copyright 2025 The Kubernetes Authors.

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

/*
commentstart is a simple analysis tool that checks if the first line of a comment is the same as the json tag.

By convention in Go, comments typically start with the name of the item being described.
In the case of field names, this would mean for a field Foo, the comment should look like:

	// Foo is a field that does something.
	Foo string `json:"foo"`

However, in Kubernetes API types, the json tag is often used to generate documentation.
In this case, the comment should start with the json tag, like so:

	// foo is a field that does something.
	Foo string `json:"foo"`

This ensures that for any generated documentation, the documentation refers to the serialized field name.
We expect most readers of Kubernetes API documentation will be more familiar with the serialized field names than the Go field names.
*/
package commentstart
