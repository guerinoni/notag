package globally

type Outer struct {
	Inner struct {
		Foo string `json:"foo"` // want "field 'Foo' contains denied tags: 'json'"
	}
}
