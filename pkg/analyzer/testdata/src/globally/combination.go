package globally

type Both struct {
	Name string `xml:"name" json:"name"` // want "field 'Name' contains denied tags: 'json,xml'"
}
