package globally 

type Example struct { // want "field 'Name' contains denied tags: 'json'"
	Name string `json:"name"`
}
