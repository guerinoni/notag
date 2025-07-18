package globally

type Both struct { // want "field 'Name' contains denied tags: 'json,xml'"
	Name string `xml:"name" json:"name"`
}
