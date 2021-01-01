package model

import (
	"strings"
)

type Location struct {
	Name    string
	Aliases []Location
	Unique  bool
}

type Locations []struct {
	Name    string
	Aliases []Location
	Cities  []Location
}

func (l Locations) Locate(text string) (locations []string) {
	for _, location := range l {
		if text == location.Name {
			return append(locations, location.Name)
		}
		for _, alias := range location.Aliases {
			if text == alias.Name && alias.Unique {
				return append(locations, location.Name)
			}
		}
	}
	for _, location := range l {
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
	for _, location := range l {
		for _, city := range location.Cities {
			if l.isFuzzy(city.Name) {
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

func (l Locations) isFuzzy(text string) bool {
	return len(text) <= 5
}
