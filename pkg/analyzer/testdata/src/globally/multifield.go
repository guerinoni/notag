package globally

type MultiViolations struct { // want "field 'Name' contains denied tags: 'json'" "field 'Age' contains denied tags: 'xml'"
	Name string `json:"name"`
	Age  int    `xml:"age"`
}
