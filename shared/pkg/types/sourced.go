package types

type Origin uint8

const (
	OriginDatabase Origin = iota
	OriginCache
)

func (o Origin) String() string {
	if o == OriginCache {
		return "cache"
	}

	return "database"
}

type Sourced[T any] struct {
	Value  T
	Origin Origin
}

func FromCache[T any](value T) Sourced[T] {
	return Sourced[T]{Value: value, Origin: OriginCache}
}

func FromDatabase[T any](value T) Sourced[T] {
	return Sourced[T]{Value: value, Origin: OriginDatabase}
}
