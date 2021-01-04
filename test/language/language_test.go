package language

import (
	"fmt"
	"github.com/memochou1993/github-rankings/app/resource"
	"github.com/memochou1993/github-rankings/test"
	"os"
	"reflect"
	"strconv"
	"testing"
)

func TestMain(m *testing.M) {
	setUp()
	code := m.Run()
	tearDown()
	os.Exit(code)
}

func setUp() {
	test.ChangeDirectory()
	resource.Init()
}

func TestLocate(t *testing.T) {
	cases := []struct {
		name     string
		expected []string
		actual   []string
	}{
		{
			name:     "Taipei, Taiwan",
			expected: []string{"Taiwan", "Taipei"},
			actual:   resource.Locate("Taipei, Taiwan"),
		},
		{
			name:     "taipei, taiwan",
			expected: []string{"Taiwan", "Taipei"},
			actual:   resource.Locate("taipei, taiwan"),
		},
		{
			name:     "TAIPEI, TAIWAN",
			expected: []string{"Taiwan", "Taipei"},
			actual:   resource.Locate("TAIPEI, TAIWAN"),
		},
		{
			name:     "Taiwan",
			expected: []string{"Taiwan"},
			actual:   resource.Locate("Taiwan"),
		},
		{
			name:     "Taipei",
			expected: []string{"Taiwan", "Taipei"},
			actual:   resource.Locate("Taipei"),
		},
		{
			name:     "Congo",
			expected: []string{"Congo"},
			actual:   resource.Locate("Congo"),
		},
		{
			name:     "Congo (DRC)",
			expected: []string{"Congo (DRC)"},
			actual:   resource.Locate("Congo (DRC)"),
		},
		{
			name:     "Lubumbashi, Congo (DRC)",
			expected: []string{"Congo (DRC)"},
			actual:   resource.Locate("Lubumbashi, Congo (DRC)"),
		},
		{
			name:     "Democratic Republic of the Congo",
			expected: []string{"Congo (DRC)"},
			actual:   resource.Locate("Democratic Republic of the Congo"),
		},
		{
			name:     "Niger",
			expected: []string{"Niger"},
			actual:   resource.Locate("Niger"),
		},
		{
			name:     "Nigeria",
			expected: []string{"Nigeria"},
			actual:   resource.Locate("Nigeria"),
		},
		{
			name:     "Lagos, Nigeria",
			expected: []string{"Nigeria", "Lagos"},
			actual:   resource.Locate("Lagos, Nigeria"),
		},
		{
			name:     "Netherlands Antilles",
			expected: []string{"Curacao"},
			actual:   resource.Locate("Netherlands Antilles"),
		},
		{
			name:     "Brasil",
			expected: []string{"Brazil"},
			actual:   resource.Locate("Brasil"),
		},
		{
			name:     "Formosa",
			expected: []string{},
			actual:   resource.Locate("Formosa"),
		},
		{
			name:     "Taipei, Formosa",
			expected: []string{"Taiwan", "Taipei"},
			actual:   resource.Locate("Taipei, Formosa"),
		},
		{
			name:     "Formosa, Argentina",
			expected: []string{"Argentina", "Formosa"},
			actual:   resource.Locate("Formosa, Argentina"),
		},
		{
			name:     "Ilan",
			expected: []string{},
			actual:   resource.Locate("Ilan"),
		},
		{
			name:     "Ilan, Taiwan",
			expected: []string{"Taiwan", "Ilan"},
			actual:   resource.Locate("Ilan, Taiwan"),
		},
		{
			name:     "Central",
			expected: []string{},
			actual:   resource.Locate("Central"),
		},
		{
			name:     "Taioan, Taiwan",
			expected: []string{"Taiwan", "Tainan"},
			actual:   resource.Locate("Taioan, Taiwan"),
		},
		{
			name:     "Taioan",
			expected: []string{"Taiwan", "Tainan"},
			actual:   resource.Locate("Taioan"),
		},
		{
			name:     "Takao, Taiwan",
			expected: []string{"Taiwan"},
			actual:   resource.Locate("Takao, Taiwan"),
		},
		{
			name:     "Takao",
			expected: []string{},
			actual:   resource.Locate("Takao"),
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

func tearDown() {
	//
}
