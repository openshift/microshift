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
* The `forbiddenmarkers` linter ensures that types and fields do not contain any markers
* that are forbidden.
*
* By default, `forbiddenmarkers` is not enabled.
*
* It can be configured with a list of marker identifiers and optionally their attributes and values that are forbidden.
*
* Some examples configurations explained:
*
* **Scenario:** forbid all instances of the marker with the identifier `forbidden:marker`
*
* ```yaml
* linterConfig:
*   forbiddenmarkers:
*     markers:
*       - identifier: "forbidden:marker"
* ```
*
* **Scenario:** forbid all instances of the marker with the identifier `forbidden:marker` containing the attribute 'fruit'
*
* ```yaml
* linterConfig:
*   forbiddenmarkers:
*     markers:
*       - identifier: "forbidden:marker"
*         ruleSets:
*           - attributes:
*               - name: "fruit"
* ```
*
* **Scenario:** forbid all instances of the marker with the identifier `forbidden:marker` containing the 'fruit' AND 'color' attributes
*
* ```yaml
* linterConfig:
*   forbiddenmarkers:
*     markers:
*       - identifier: "forbidden:marker"
*         ruleSets:
*           - attributes:
*               - name: "fruit"
*               - name: "color"
* ```
*
* **Scenario:** forbid all instances of the marker with the identifier `forbidden:marker` where the 'fruit' attribute is set to one of 'apple', 'banana', or 'orange'
*
* ```yaml
* linterConfig:
*   forbiddenmarkers:
*     markers:
*       - identifier: "forbidden:marker"
*         ruleSets:
*           - attributes:
*               - name: "fruit"
*                 values:
*                   - "apple"
*                   - "banana"
*                   - "orange"
* ```
*
* **Scenario:** forbid all instances of the marker with the identifier `forbidden:marker` where the 'fruit' attribute is set to one of 'apple', 'banana', or 'orange' AND the 'color' attribute is set to one of 'red', 'green', or 'blue'
*
* ```yaml
* linterConfig:
*   forbiddenmarkers:
*     markers:
*       - identifier: "forbidden:marker"
*         ruleSets:
*           - attributes:
*               - name: "fruit"
*                 values:
*                   - "apple"
*                   - "banana"
*                   - "orange"
*               - name: "color"
*                 values:
*                   - "red"
*                   - "blue"
*                   - "green"
* ```
*
* **Scenario:** forbid all instances of the marker with the identifier `forbidden:marker` where:
* - The `fruit` attribute is set to `apple` and the `color` attribute is set to one of `blue` or `orange` (allow any other color apple)
* OR
* - The `fruit` attribute is set to `orange` and the `color` attribute is set to one of `blue`, `red`, or `green` (allow any other color orange)
* OR
* - The `fruit` attribute is set to `banana` (no bananas allowed)
*
* ```yaml
* linterConfig:
*   forbiddenmarkers:
*     markers:
*       - identifier: "forbidden:marker"
*         ruleSets:
*           - attributes:
*               - name: "fruit"
*                 values:
*                   - "apple"
*               - name: "color"
*                 values:
*                   - "blue"
*                   - "orange"
*           - attributes:
*               - name: "fruit"
*                 values:
*                   - "orange"
*               - name: "color"
*                 values:
*                   - "blue"
*                   - "red"
*                   - "green"
*           - attributes:
*               - name: "fruit"
*                 values:
*                   - "banana"
* ```
*
*
* Fixes are suggested to remove all markers that are forbidden.
*
 */
package forbiddenmarkers
