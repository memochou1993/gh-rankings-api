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

func Init() {
	util.LoadAsset("languages", &Languages)
	util.LoadAsset("locations", &Locations)
}

func Locate(text string) (locations []string) {
	for _, location := range Locations {
		if text == location.Name {
			return append(locations, location.Name)
		}
		for _, alias := range location.Aliases {
			if text == alias.Name && alias.Unique {
				return append(locations, location.Name)
			}
		}
	}
	for _, location := range Locations {
		if strings.Contains(text, location.Name) {
			locations = append(locations, location.Name)
			for _, city := range location.Cities {
				if strings.Contains(text, city.Name) {
					return append(locations, city.Name)
				}
				for _, alias := range city.Aliases {
					if strings.Contains(text, alias.Name) {
						return append(locations, city.Name)
					}
				}
			}
			return
		}
		for _, alias := range location.Aliases {
			if strings.Contains(text, alias.Name) && alias.Unique {
				locations = append(locations, location.Name)
				for _, city := range location.Cities {
					if strings.Contains(text, city.Name) {
						return append(locations, city.Name)
					}
					for _, alias := range city.Aliases {
						if strings.Contains(text, alias.Name) {
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
			if strings.Contains(text, city.Name) && city.Unique {
				return append(locations, location.Name, city.Name)
			}
			for _, alias := range city.Aliases {
				if strings.Contains(text, alias.Name) {
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
