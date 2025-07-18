package globally 

type ExampleXML struct { // want "field 'Name' contains denied tags: 'xml'"
	Name string `xml:"name"`
}
