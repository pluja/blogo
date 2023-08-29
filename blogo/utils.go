package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func StringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

// Given a path, return the slug and extension
func ParseFilePath(fPath string) (string, string) {
	filenameWithExt := filepath.Base(fPath)
	extension := filepath.Ext(filenameWithExt)
	filename := strings.TrimSuffix(filenameWithExt, extension)
	return filename, extension
}

// Given a map, returns the value of a key as a string
func GetMapStringValue(metadata map[string]interface{}, key string) string {
	if value, ok := metadata[key].(string); ok {
		return value
	} else if value, ok := metadata[key].(bool); ok {
		return fmt.Sprintf("%v", value)
	} else if value, ok := metadata[key].(int); ok {
		return fmt.Sprintf("%v", value)
	}
	return ""
}

// Difference returns the difference between two collections.
// The first value is the collection of element absent of list2.
// The second value is the collection of element absent of list1.
func Difference[T comparable](list1 []T, list2 []T) ([]T, []T) {
	left := []T{}
	right := []T{}

	seenLeft := map[T]struct{}{}
	seenRight := map[T]struct{}{}

	for _, elem := range list1 {
		seenLeft[elem] = struct{}{}
	}

	for _, elem := range list2 {
		seenRight[elem] = struct{}{}
	}

	for _, elem := range list1 {
		if _, ok := seenRight[elem]; !ok {
			left = append(left, elem)
		}
	}

	for _, elem := range list2 {
		if _, ok := seenLeft[elem]; !ok {
			right = append(right, elem)
		}
	}

	return left, right
}
