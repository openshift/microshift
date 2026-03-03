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
extractjsontags is a helper package that extracts JSON tags from a struct field.

It returns data behind the interface [StructFieldTags] which is used to find information about JSON tags on fields within a struct.

Data about json tags, for a field within a struct can be accessed by calling the `FieldTags` method on the interface.

Example:

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	jsonTags := pass.ResultOf[extractjsontags.Analyzer].(extractjsontags.StructFieldTags)

	// Filter to fields so that we can iterate over fields in a struct.
	nodeFilter := []ast.Node{
		(*ast.Field)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		field, ok := n.(*ast.Field)
		if !ok {
			return
		}

		tagInfo := jsonTags.FieldTags(field)

		...

	})

For each field, tag information is returned as a [FieldTagInfo] struct.
This can be used to determine the name of the field, as per the json tag, whether the
field is inline, has omitempty or is missing completely.
*/
package extractjsontags
