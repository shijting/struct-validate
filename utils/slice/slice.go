package slice

func Contains[T comparable](src []T, dst T) bool {
	for _, v := range src {
		if v == dst {
			return true
		}
	}
	return false
}

type equalFunc[T any] func(src, dst T) bool

func ContainsFunc[T any](src []T, dst T, equal equalFunc[T]) bool {
	for _, v := range src {
		if equal(v, dst) {
			return true
		}
	}
	return false
}
