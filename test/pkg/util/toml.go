package util

import (
	"fmt"
	"strings"
)

// GetTOMLFieldValue obtains value of first variable named `field`.
// It does not use the TOML parser as it is expected to be used on
// TOML files that are yet to be templated.
//
// Primary goal is to stop relying on assumption that filename without extension is the name/id of Blueprint/Source.
// Because Blueprints/Sources are removed prior re-adding, the name/id must be precise.
func GetTOMLFieldValue(data, field string) (string, error) {
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, field) {
			fields := strings.Split(line, "\"")
			if len(fields) != 3 {
				return "", fmt.Errorf("found matching field in TOML but splitting with double quotes gave unexpected results, line=%q, after split=%q", line, fields)
			}
			// name = "VALUE"
			// ------- ----- -
			//    0      1   2
			return fields[1], nil
		}
	}
	return "", nil
}
