package sets

func Delta[K comparable](a, b map[K]struct{}) map[K]struct{} {
	delta := make(map[K]struct{})
	for k := range a {
		if _, ok := b[k]; !ok {
			delta[k] = struct{}{}
		}
	}
	return delta
}
