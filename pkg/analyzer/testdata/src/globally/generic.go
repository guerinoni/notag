package globally

type Generic[T any] struct {
	Value T `json:"value"` // want "field 'Value' contains denied tags: 'json'"
}
