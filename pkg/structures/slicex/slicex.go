package slicex

import "slices"

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

func Cap[S ~[]E, E any](x S, maxlen int) S {
	return x[:min(len(x), maxlen)]
}

func Flatten[S ~[]E, E any](input []S) []E {
	var result []E

	for _, row := range input {
		result = append(result, row...)
	}

	return result
}
