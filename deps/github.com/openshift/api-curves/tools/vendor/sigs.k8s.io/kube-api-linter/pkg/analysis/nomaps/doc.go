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
nomaps provides a linter to ensure that fields do not use map types.

Maps are discouraged in Kubernetes APIs. It is hard to distinguish between structs and maps in JSON/YAML and as such, lists of named subobjects are preferred over plain map types.

Instead of

	ports:
	  www:
	    containerPort: 80

use

	ports:
	  - name: www
	    containerPort: 80

Lists should use the `+listType=map` and `+listMapKey=name` markers, or equivalent.
*/
package nomaps
