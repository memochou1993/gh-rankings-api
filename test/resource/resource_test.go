package resource

import (
	"fmt"
	"github.com/memochou1993/gh-rankings/app/resource"
	"reflect"
	"strconv"
	"testing"
)

func TestLocate(t *testing.T) {
	cases := []struct {
		name     string
		expected []string
		actual   []string
	}{
		{
			name:     "Taipei, Taiwan",
			expected: []string{"Taiwan", "Taipei, Taiwan"},
			actual:   join(resource.Locate("Taipei, Taiwan")),
		},
		{
			name:     "taipei, taiwan",
			expected: []string{"Taiwan", "Taipei, Taiwan"},
			actual:   join(resource.Locate("taipei, taiwan")),
		},
		{
			name:     "TAIPEI, TAIWAN",
			expected: []string{"Taiwan", "Taipei, Taiwan"},
			actual:   join(resource.Locate("TAIPEI, TAIWAN")),
		},
		{
			name:     "Taiwan",
			expected: []string{"Taiwan", ""},
			actual:   join(resource.Locate("Taiwan")),
		},
		{
			name:     "Taipei",
			expected: []string{"Taiwan", "Taipei, Taiwan"},
			actual:   join(resource.Locate("Taipei")),
		},
		{
			name:     "Congo",
			expected: []string{"Congo", ""},
			actual:   join(resource.Locate("Congo")),
		},
		{
			name:     "Congo (DRC)",
			expected: []string{"Congo (DRC)", ""},
			actual:   join(resource.Locate("Congo (DRC)")),
		},
		{
			name:     "Lubumbashi, Congo (DRC)",
			expected: []string{"Congo (DRC)", ""},
			actual:   join(resource.Locate("Lubumbashi, Congo (DRC)")),
		},
		{
			name:     "Democratic Republic of the Congo",
			expected: []string{"Congo (DRC)", ""},
			actual:   join(resource.Locate("Democratic Republic of the Congo")),
		},
		{
			name:     "Niger",
			expected: []string{"Niger", ""},
			actual:   join(resource.Locate("Niger")),
		},
		{
			name:     "Nigeria",
			expected: []string{"Nigeria", ""},
			actual:   join(resource.Locate("Nigeria")),
		},
		{
			name:     "Lagos, Nigeria",
			expected: []string{"Nigeria", "Lagos, Nigeria"},
			actual:   join(resource.Locate("Lagos, Nigeria")),
		},
		{
			name:     "Netherlands Antilles",
			expected: []string{"Curacao", ""},
			actual:   join(resource.Locate("Netherlands Antilles")),
		},
		{
			name:     "Brasil",
			expected: []string{"Brazil", ""},
			actual:   join(resource.Locate("Brasil")),
		},
		{
			name:     "Formosa",
			expected: []string{"", ""},
			actual:   join(resource.Locate("Formosa")),
		},
		{
			name:     "Taipei, Formosa",
			expected: []string{"Taiwan", "Taipei, Taiwan"},
			actual:   join(resource.Locate("Taipei, Formosa")),
		},
		{
			name:     "Formosa, Argentina",
			expected: []string{"Argentina", "Formosa, Argentina"},
			actual:   join(resource.Locate("Formosa, Argentina")),
		},
		{
			name:     "Ilan",
			expected: []string{"", ""},
			actual:   join(resource.Locate("Ilan")),
		},
		{
			name:     "Ilan, Taiwan",
			expected: []string{"Taiwan", "Ilan, Taiwan"},
			actual:   join(resource.Locate("Ilan, Taiwan")),
		},
		{
			name:     "Central",
			expected: []string{"", ""},
			actual:   join(resource.Locate("Central")),
		},
		{
			name:     "Taioan, Taiwan",
			expected: []string{"Taiwan", "Tainan, Taiwan"},
			actual:   join(resource.Locate("Taioan, Taiwan")),
		},
		{
			name:     "Taioan",
			expected: []string{"Taiwan", "Tainan, Taiwan"},
			actual:   join(resource.Locate("Taioan")),
		},
		{
			name:     "Takao, Taiwan",
			expected: []string{"Taiwan", ""},
			actual:   join(resource.Locate("Takao, Taiwan")),
		},
		{
			name:     "Takao",
			expected: []string{"", ""},
			actual:   join(resource.Locate("Takao")),
		},
	}

	for _, c := range cases {
		if len(c.expected) == 0 && len(c.actual) == 0 {
			continue
		}
		if !reflect.DeepEqual(c.expected, c.actual) {
			t.Error(fmt.Sprintf("Test: %s, Expected: %s, Actual: %s", strconv.Quote(c.name), c.expected, c.actual))
		}
	}
}

func join(location, city string) []string {
	return []string{location, city}
}
