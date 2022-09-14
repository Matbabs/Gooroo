// This package describes the private utility and generic functions to the other methods of the main gooroo package.
package utils

import (
	"fmt"

	"github.com/lithammer/shortuuid"
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

// Generates a uuid.
func GenerateShortId() string {
	return shortuuid.New()
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
	return fmt.Sprintf("%s#%d", file, no)
}
