package slicex

func MinsFunc[S ~[]E, E any](x S, cmp func(a, b E) int) []E {
	if len(x) < 1 {
		panic("slicex.MinsFunc: empty list")
	}

	// get the minumum value
	m := x[0]
	for i := 1; i < len(x); i++ {
		if cmp(x[i], m) < 0 {
			m = x[i]
		}
	}

	// get all the values
	res := make(S, 0)
	for _, val := range x {
		if cmp(val, m) == 0 {
			res = append(res, val)
		}
	}
	return res
}

func Cap[S ~[]E, E any](x S, maxlen int) S {
	return x[:min(len(x), maxlen)]
}
