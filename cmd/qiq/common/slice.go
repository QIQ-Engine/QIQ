package common

import "strings"

func ImplodeSlice(slice []string, separator string) string {
	var result strings.Builder
	for i, s := range slice {
		if i > 0 {
			result.WriteString(", ")
		}
		result.WriteString(s)
	}
	return result.String()
}

func ImplodeStrSlice(slice []string) string {
	var result strings.Builder
	for i, s := range slice {
		if i > 0 {
			result.WriteString(", ")
		}
		result.WriteString(`"` + s + `"`)
	}
	return result.String()
}

func RemoveIndex[T any](s []T, index int) []T {
	return append(append(make([]T, 0), s[:index]...), s[index+1:]...)
}
