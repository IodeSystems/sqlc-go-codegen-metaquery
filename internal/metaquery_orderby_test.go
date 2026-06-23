package golang

import "testing"

func TestHasTopLevelOrderBy(t *testing.T) {
	cases := []struct {
		name string
		sql  string
		want bool
	}{
		{"trailing", "SELECT id, name FROM authors ORDER BY name", true},
		{"trailing semicolon", "SELECT id FROM authors ORDER BY name;", true},
		{"lowercase", "select id from authors order by name", true},
		{"with limit", "SELECT id FROM authors ORDER BY name LIMIT 10", true},
		{"union trailing", "SELECT id FROM a UNION SELECT id FROM b ORDER BY id", true},
		{"newline before", "SELECT id\nFROM authors\nORDER BY\n  name DESC", true},

		{"none", "SELECT id, name FROM authors", false},
		{"subquery only", "SELECT * FROM (SELECT id FROM a ORDER BY id) t", false},
		{"order in string literal", "SELECT 'order by x' AS note FROM authors", false},
		{"order in line comment", "SELECT id FROM authors -- order by name\n", false},
		{"order in block comment", "SELECT id FROM authors /* order by name */", false},
		{"window order not top level", "SELECT row_number() OVER (ORDER BY id) FROM a", false},
		{"identifier prefix", "SELECT reorder_by FROM authors", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := hasTopLevelOrderBy(c.sql); got != c.want {
				t.Errorf("hasTopLevelOrderBy(%q) = %v, want %v", c.sql, got, c.want)
			}
		})
	}
}
