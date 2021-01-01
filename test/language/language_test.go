package language

import (
	"fmt"
	"github.com/memochou1993/github-rankings/app/model"
	"github.com/memochou1993/github-rankings/test"
	"github.com/memochou1993/github-rankings/util"
	"os"
	"reflect"
	"strconv"
	"testing"
)

var (
	locations *model.Locations
)

func TestMain(m *testing.M) {
	setUp()
	code := m.Run()
	tearDown()
	os.Exit(code)
}

func setUp() {
	test.ChangeDirectory()
	util.LoadAsset("locations", &locations)
}

func TestLocate(t *testing.T) {
	tests := []struct {
		name     string
		expected []string
		actual   []string
	}{
		{
			name:     "Taipei, Taiwan",
			expected: []string{"Taiwan", "Taipei"},
			actual:   locations.Locate("Taipei, Taiwan"),
		},
		// TODO: lowercase
		// {
		// 	name: "taipei, taiwan",
		// 	expected: []string{},
		// 	actual:   locations.Locate("taipei, taiwan"),
		// },
		// TODO: uppercase
		// {
		// 	name: "TAIPEI, TAIWAN",
		// 	expected: []string{},
		// 	actual:   locations.Locate("TAIPEI, TAIWAN"),
		// },
		{
			name:     "Taiwan",
			expected: []string{"Taiwan"},
			actual:   locations.Locate("Taiwan"),
		},
		{
			name:     "Taipei",
			expected: []string{"Taiwan", "Taipei"},
			actual:   locations.Locate("Taipei"),
		},
		{
			name:     "Congo",
			expected: []string{"Congo"},
			actual:   locations.Locate("Congo"),
		},
		{
			name:     "Congo (DRC)",
			expected: []string{"Congo (DRC)"},
			actual:   locations.Locate("Congo (DRC)"),
		},
		{
			name:     "Lubumbashi, Congo (DRC)",
			expected: []string{"Congo (DRC)"},
			actual:   locations.Locate("Lubumbashi, Congo (DRC)"),
		},
		{
			name:     "Democratic Republic of the Congo",
			expected: []string{"Congo (DRC)"},
			actual:   locations.Locate("Democratic Republic of the Congo"),
		},
		{
			name:     "Niger",
			expected: []string{"Niger"},
			actual:   locations.Locate("Niger"),
		},
		{
			name:     "Nigeria",
			expected: []string{"Nigeria"},
			actual:   locations.Locate("Nigeria"),
		},
		{
			name:     "Lagos, Nigeria",
			expected: []string{"Nigeria", "Lagos"},
			actual:   locations.Locate("Lagos, Nigeria"),
		},
		{
			name:     "Netherlands Antilles",
			expected: []string{"Curacao"},
			actual:   locations.Locate("Netherlands Antilles"),
		},
		{
			name:     "Formosa",
			expected: []string{},
			actual:   locations.Locate("Formosa"),
		},
		{
			name:     "Taipei, Formosa",
			expected: []string{"Taiwan", "Taipei"},
			actual:   locations.Locate("Taipei, Formosa"),
		},
		{
			name:     "Formosa, Argentina",
			expected: []string{"Argentina", "Formosa"},
			actual:   locations.Locate("Formosa, Argentina"),
		},
		{
			name:     "Ilan",
			expected: []string{},
			actual:   locations.Locate("Ilan"),
		},
		{
			name:     "Ilan, Taiwan",
			expected: []string{"Taiwan", "Ilan"},
			actual:   locations.Locate("Ilan, Taiwan"),
		},
		{
			name:     "Central",
			expected: []string{},
			actual:   locations.Locate("Central"),
		},
		{
			name:     "Taioan, Taiwan",
			expected: []string{"Taiwan", "Tainan"},
			actual:   locations.Locate("Taioan, Taiwan"),
		},
		{
			name:     "Taioan",
			expected: []string{"Taiwan", "Tainan"},
			actual:   locations.Locate("Taioan"),
		},
		{
			name:     "Takao, Taiwan",
			expected: []string{"Taiwan"},
			actual:   locations.Locate("Takao, Taiwan"),
		},
		{
			name:     "Takao",
			expected: []string{},
			actual:   locations.Locate("Takao"),
		},
	}

	for _, test := range tests {
		if len(test.expected) == 0 && len(test.actual) == 0 {
			continue
		}
		if !reflect.DeepEqual(test.expected, test.actual) {
			t.Error(fmt.Sprintf("Test: %s, Expected: %s, Actual: %s", strconv.Quote(test.name), test.expected, test.actual))
		}
	}
}

func tearDown() {
	//
}
