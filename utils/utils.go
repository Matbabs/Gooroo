package utils

import (
	"fmt"

	"github.com/lithammer/shortuuid"
)

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func GenerateShortId() string {
	return shortuuid.New()
}

func MapInit[T any](key string, _map map[string]T, value T) {
	_, isPresent := _map[key]
	if !isPresent {
		_map[key] = value
	}
}

func CallerToKey(file string, no int) string {
	return fmt.Sprintf("%s#%d", file, no)
}
