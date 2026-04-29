package globally

type ExampleXML struct {
	Name string `xml:"name"` // want "field 'Name' contains denied tags: 'xml'"
}
