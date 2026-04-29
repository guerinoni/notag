package globally

type Inner struct {
	X int
}

type WithEmbedded struct {
	Inner `json:"inner"` // want "field 'Inner' contains denied tags: 'json'"
}
