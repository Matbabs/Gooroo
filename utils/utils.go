// This package describes the private utility and generic functions to the other methods of the main gooroo package.
package utils

import (
	"fmt"
	"strings"
)

// Check for the presence of a 'string' in a '[]string'.
func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// Initializes a value in a map if and only if it does not exist.
func MapInit[T any](key string, _map map[string]T, value T) {
	_, isPresent := _map[key]
	if !isPresent {
		_map[key] = value
	}
}

// Initializes a function (value) in a map if and only if it does not exist.
func MapInitCallback[T any](key string, _map map[string]T, callback func() T) {
	_, isPresent := _map[key]
	if !isPresent {
		_map[key] = callback()
	}
}

// Formatting a string for key management
func CallerToKey(file string, no int) string {
	splitFile := strings.Split(file, "/")
	return fmt.Sprintf("%s#%d", splitFile[len(splitFile)-1], no)
}

// Convert 'any' type to 'string'.
func AnyStr(v any) string {
	return fmt.Sprintf("%v", v)
}
