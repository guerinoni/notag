package globally

type MultiViolations struct {
	Name string `json:"name"` // want "field 'Name' contains denied tags: 'json'"
	Age  int    `xml:"age"`   // want "field 'Age' contains denied tags: 'xml'"
}
