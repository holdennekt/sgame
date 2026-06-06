package sets

func Delta[K comparable](a, b map[K]struct{}) map[K]struct{} {
	delta := make(map[K]struct{})
	if a == nil {
		return delta
	}
	if b == nil {
		return a
	}
	for k := range a {
		if _, ok := b[k]; !ok {
			delta[k] = struct{}{}
		}
	}
	return delta
}
