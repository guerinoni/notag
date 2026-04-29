package allowval

type Mixed struct {
	JSONDash    string `json:"-"`
	JSONName    string `json:"name"` // want "field 'JSONName' contains denied tags: 'json'"
	JSONIgnore  string `json:"ignore"`
	BothExempt  string `json:"-" xml:"-"`
	JSONExempt  string `json:"-" xml:"name"`   // want "field 'JSONExempt' contains denied tags: 'xml'"
	XMLExempt   string `json:"name" xml:"-"`   // want "field 'XMLExempt' contains denied tags: 'json'"
	BothDenied  string `json:"a" xml:"b"`      // want "field 'BothDenied' contains denied tags: 'json,xml'"
	DBOnly      string `db:"d"`
	DBWithJSON  string `db:"d" json:"-"`
	WithComma   string `json:"name,omitempty"` // want "field 'WithComma' contains denied tags: 'json'"
}
