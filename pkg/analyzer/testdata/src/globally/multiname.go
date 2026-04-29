package globally

type MultiName struct {
	A, B string `json:"shared"` // want "field 'A' contains denied tags: 'json'" "field 'B' contains denied tags: 'json'"
}
