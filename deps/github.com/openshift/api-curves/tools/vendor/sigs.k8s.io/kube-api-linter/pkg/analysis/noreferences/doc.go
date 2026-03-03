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
The `noreferences` linter ensures that field names use 'Ref'/'Refs' instead of 'Reference'/'References'.
By default, `noreferences` is enabled and enforces this naming convention.
The linter checks that 'Reference' is present at the beginning or end of the field name, and replaces it with 'Ref'.
Similarly, 'References' anywhere in field names is replaced with 'Refs'.

Example configuration:
Default behavior (allow Ref/Refs in field names):

	  lintersConfig:
		noreferences:
		  policy: PreferAbbreviatedReference

Strict mode (forbid Ref/Refs in field names):

	lintersConfig:
		noreferences:
	  		policy: NoReferences

When `policy` is set to `PreferAbbreviatedReference` (the default), fields containing 'Ref' or 'Refs' are allowed.
The policy can be set to `NoReferences` to also report errors for 'Ref' or 'Refs' in field names.
*/
package noreferences
