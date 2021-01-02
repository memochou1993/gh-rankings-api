package resource

import (
	"github.com/memochou1993/github-rankings/util"
	"strings"
)

var (
	Languages []Language
	Locations []Location
)

type Language struct {
	Name string
}

type Location struct {
	Name    string
	Aliases []Location
	Cities  []Location
	Unique  bool
}

func (l Location) is(name string) bool {
	return strings.ToUpper(l.Name) == strings.ToUpper(name)
}

func (l Location) isSimilar(name string) bool {
	return strings.Contains(strings.ToUpper(name), strings.ToUpper(l.Name))
}

func (l Location) isUnique() bool {
	return l.Unique
}

func Init() {
	util.LoadAsset("languages", &Languages)
	util.LoadAsset("locations", &Locations)
}

func Locate(text string) (locations []string) {
	for _, location := range Locations {
		if location.is(text) {
			return append(locations, location.Name)
		}
		for _, alias := range location.Aliases {
			if alias.is(text) && alias.isUnique() {
				return append(locations, location.Name)
			}
		}
	}
	for _, location := range Locations {
		if location.isSimilar(text) {
			locations = append(locations, location.Name)
			for _, city := range location.Cities {
				if city.isSimilar(text) {
					return append(locations, city.Name)
				}
				for _, alias := range city.Aliases {
					if alias.isSimilar(text) {
						return append(locations, city.Name)
					}
				}
			}
			return
		}
		for _, alias := range location.Aliases {
			if alias.isSimilar(text) && alias.isUnique() {
				locations = append(locations, location.Name)
				for _, city := range location.Cities {
					if city.isSimilar(text) {
						return append(locations, city.Name)
					}
					for _, alias := range city.Aliases {
						if alias.isSimilar(text) {
							return append(locations, city.Name)
						}
					}
				}
				return
			}
		}
	}
	for _, location := range Locations {
		for _, city := range location.Cities {
			if isFuzzy(city.Name) {
				continue
			}
			if city.isSimilar(text) && city.isUnique() {
				return append(locations, location.Name, city.Name)
			}
			for _, alias := range city.Aliases {
				if alias.isSimilar(text) {
					return append(locations, location.Name, city.Name)
				}
			}
		}
	}
	return
}

func isFuzzy(text string) bool {
	return len(text) <= 5
}
