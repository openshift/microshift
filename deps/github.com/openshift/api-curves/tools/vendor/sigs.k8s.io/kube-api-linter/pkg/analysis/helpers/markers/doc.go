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
markers is a helper used to extract marker information from types.
A marker is a comment line preceded with `+` that indicates to a generator something about the field or type.

The package returns a [Markers] interface, which can be used to access markers associated with a struct or a field within a struct.

Example:

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	markersAccess := pass.ResultOf[markers.Analyzer].(markers.Markers)

	// Filter to structs so that we can iterate over fields in a struct.
	nodeFilter := []ast.Node{
		(*ast.StructType)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		sTyp, ok := n.(*ast.StructType)
		if !ok {
			return
		}

		if sTyp.Fields == nil {
			return
		}

		for _, field := range sTyp.Fields.List {
			if field == nil || len(field.Names) == 0 {
				continue
			}

			structMarkers := markersAccess.StructMarkers(sTyp)
			fieldMarkers := markersAccess.FieldMarkers(field)

			...
		}
	})

The result of StructMarkers or StructFieldMarkers is a [MarkerSet] which can be used to determine the presence of a marker, and the value of the marker.
The MarkerSet is indexed based on the value of the marker, once the prefix `+` is removed.

Additional information about the marker can be found in the [Marker] struct, for each marker on the field.

Example:

	fieldMarkers := markersAccess.FieldMarkers(field)

	if fieldMarkers.Has("required") {
		requiredMarker := fieldMarkers["required"]
		...
	}
*/
package markers
