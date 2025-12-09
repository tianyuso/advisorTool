// Package utils is a utility library for advisorTool.
//
//nolint:revive
package utils

import (
	"unicode"
)

// Uniq removes duplicate elements from a slice.
func Uniq[T comparable](array []T) []T {
	res := make([]T, 0, len(array))
	seen := make(map[T]struct{}, len(array))

	for _, e := range array {
		if _, ok := seen[e]; ok {
			continue
		}
		seen[e] = struct{}{}
		res = append(res, e)
	}

	return res
}

// IsSpaceOrSemicolon checks if the rune is a space or a semicolon.
func IsSpaceOrSemicolon(r rune) bool {
	if ok := unicode.IsSpace(r); ok {
		return true
	}
	return r == ';'
}
