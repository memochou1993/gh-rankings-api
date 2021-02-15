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
			actual:   concat(resource.Locate("Taipei, Taiwan")),
		},
		{
			name:     "taipei, taiwan",
			expected: []string{"Taiwan", "Taipei, Taiwan"},
			actual:   concat(resource.Locate("taipei, taiwan")),
		},
		{
			name:     "TAIPEI, TAIWAN",
			expected: []string{"Taiwan", "Taipei, Taiwan"},
			actual:   concat(resource.Locate("TAIPEI, TAIWAN")),
		},
		{
			name:     "Taiwan",
			expected: []string{"Taiwan", ""},
			actual:   concat(resource.Locate("Taiwan")),
		},
		{
			name:     "Taipei",
			expected: []string{"Taiwan", "Taipei, Taiwan"},
			actual:   concat(resource.Locate("Taipei")),
		},
		{
			name:     "Congo",
			expected: []string{"Congo", ""},
			actual:   concat(resource.Locate("Congo")),
		},
		{
			name:     "Congo (DRC)",
			expected: []string{"Congo (DRC)", ""},
			actual:   concat(resource.Locate("Congo (DRC)")),
		},
		{
			name:     "Lubumbashi, Congo (DRC)",
			expected: []string{"Congo (DRC)", ""},
			actual:   concat(resource.Locate("Lubumbashi, Congo (DRC)")),
		},
		{
			name:     "Democratic Republic of the Congo",
			expected: []string{"Congo (DRC)", ""},
			actual:   concat(resource.Locate("Democratic Republic of the Congo")),
		},
		{
			name:     "Niger",
			expected: []string{"Niger", ""},
			actual:   concat(resource.Locate("Niger")),
		},
		{
			name:     "Nigeria",
			expected: []string{"Nigeria", ""},
			actual:   concat(resource.Locate("Nigeria")),
		},
		{
			name:     "Lagos, Nigeria",
			expected: []string{"Nigeria", "Lagos, Nigeria"},
			actual:   concat(resource.Locate("Lagos, Nigeria")),
		},
		{
			name:     "Netherlands Antilles",
			expected: []string{"Curacao", ""},
			actual:   concat(resource.Locate("Netherlands Antilles")),
		},
		{
			name:     "Brasil",
			expected: []string{"Brazil", ""},
			actual:   concat(resource.Locate("Brasil")),
		},
		{
			name:     "Formosa",
			expected: []string{"", ""},
			actual:   concat(resource.Locate("Formosa")),
		},
		{
			name:     "Taipei, Formosa",
			expected: []string{"Taiwan", "Taipei, Taiwan"},
			actual:   concat(resource.Locate("Taipei, Formosa")),
		},
		{
			name:     "Formosa, Argentina",
			expected: []string{"Argentina", "Formosa, Argentina"},
			actual:   concat(resource.Locate("Formosa, Argentina")),
		},
		{
			name:     "Ilan",
			expected: []string{"", ""},
			actual:   concat(resource.Locate("Ilan")),
		},
		{
			name:     "Ilan, Taiwan",
			expected: []string{"Taiwan", "Ilan, Taiwan"},
			actual:   concat(resource.Locate("Ilan, Taiwan")),
		},
		{
			name:     "Central",
			expected: []string{"", ""},
			actual:   concat(resource.Locate("Central")),
		},
		{
			name:     "Taioan, Taiwan",
			expected: []string{"Taiwan", "Tainan, Taiwan"},
			actual:   concat(resource.Locate("Taioan, Taiwan")),
		},
		{
			name:     "Taioan",
			expected: []string{"Taiwan", "Tainan, Taiwan"},
			actual:   concat(resource.Locate("Taioan")),
		},
		{
			name:     "Takao, Taiwan",
			expected: []string{"Taiwan", ""},
			actual:   concat(resource.Locate("Takao, Taiwan")),
		},
		{
			name:     "Takao",
			expected: []string{"", ""},
			actual:   concat(resource.Locate("Takao")),
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

func concat(location, city string) []string {
	return []string{location, city}
}
