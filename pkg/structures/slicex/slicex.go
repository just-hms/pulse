package slicex

import "slices"

// MinsFunc returns a slice of minum values inside a given slice
func MinsFunc[S ~[]E, E any](slice S, cmp func(a, b E) int) []E {
	min := slices.MinFunc(slice, cmp)

	count := 0
	for _, val := range slice {
		if cmp(val, min) == 0 {
			count++
		}
	}

	res := make(S, 0, count)
	for _, val := range slice {
		if cmp(val, min) == 0 {
			res = append(res, val)
		}
	}
	return res
}
