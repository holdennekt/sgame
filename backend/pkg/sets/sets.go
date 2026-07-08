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

func Union[K comparable](a, b map[K]struct{}) map[K]struct{} {
	union := make(map[K]struct{}, len(a)+len(b))
	for k := range a {
		union[k] = struct{}{}
	}
	for k := range b {
		union[k] = struct{}{}
	}
	return union
}
