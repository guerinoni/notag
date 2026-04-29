package globally

type Bar struct{}

type Container[T any] struct {
	V T
}

type WrapPointer struct {
	*Bar `json:"bar"` // want "field 'Bar' contains denied tags: 'json'"
}

type WrapGeneric struct {
	Container[int] `json:"c"` // want "field 'Container' contains denied tags: 'json'"
}
