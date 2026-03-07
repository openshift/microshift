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
The `namingconventions` linter ensures that field names adhere to a set of defined naming conventions.

By default, `namingconventions` is not enabled.

When enabled, it must be configured with at least one naming convention.

Naming conventions must have:
- A unique human-readable name.
- A human-readable message to be included in violation errors.
- A regular expression that will match text within the field name that violates the convention.
- A defined "operation". Allowed operations are "Inform", "Drop", "DropField", and "Replacement".

The "Inform" operation will simply inform via a linter error when a field name violates the naming convention.
The "Drop" operation will suggest a fix that drops violating text from the field name.
The "DropField" operation will suggest a fix that removes the field in it's entirety.
The "Replacement" operation will suggest a fix that replaces the violating text in the field name with a defined replacement value.

Some example configurations:

**Scenario:** Inform that any variations of the word 'fruit' in field names is not allowed
```yaml
linterConfig:

	namingconventions:
	  conventions:
	    - name: nofruit
	      violationMatcher: (?i)fruit
	      operation: Inform
	      message: fields should not contain any variation of the word 'fruit' in their names

```

**Scenario:** Drop any variations of the word 'fruit' in field names
```yaml
linterConfig:

	namingconventions:
	  conventions:
	    - name: nofruit
	      violationMatcher: (?i)fruit
	      operation: Drop
	      message: fields should not contain any variation of the word 'fruit' in their names

```

**Scenario:** Do not allow fields with any variations of the word 'fruit' in their name
```yaml
linterConfig:

	namingconventions:
	  conventions:
	    - name: nofruit
	      violationMatcher: (?i)fruit
	      operation: DropField
	      message: fields should not contain any variation of the word 'fruit' in their names

```

**Scenario:** Replace any variations of the word 'color' with 'colour' in field names
```yaml
linterConfig:

	namingconventions:
	  conventions:
	    - name: BritishEnglishColour
	      violationMatcher: (?i)color
	      operation: Replacement
	      replacement: colour
	      message: prefer 'colour' over 'color' when referring to colours in field names

```
*/
package namingconventions
