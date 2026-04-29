package globally

type Example struct {
	Name string `json:"name"` // want "field 'Name' contains denied tags: 'json'"
}
