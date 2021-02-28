package resource

import (
	"encoding/json"
	"fmt"
	"github.com/memochou1993/gh-rankings/util"
	"io/ioutil"
	"log"
	"strings"
)

var (
	SpecifiedOrganizations []SpecifiedOwner
	SpecifiedUsers         []SpecifiedOwner
	Languages              []Language
	Locations              []Location
)

type SpecifiedOwner struct {
	Login string
}

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

func init() {
	read("specified_organization", &SpecifiedOrganizations)
	read("specified_user", &SpecifiedUsers)
	read("language", &Languages)
	read("location", &Locations)
}

func Locate(text string) (location, city string) {
	for _, location := range Locations {
		if location.is(text) {
			return location.Name, ""
		}
		for _, alias := range location.Aliases {
			if alias.is(text) && alias.isUnique() {
				return location.Name, ""
			}
		}
	}
	for _, location := range Locations {
		if location.isSimilar(text) {
			for _, city := range location.Cities {
				if city.isSimilar(text) {
					return location.Name, fmt.Sprintf("%s, %s", city.Name, location.Name)
				}
				for _, alias := range city.Aliases {
					if alias.isSimilar(text) {
						return location.Name, fmt.Sprintf("%s, %s", city.Name, location.Name)
					}
				}
			}
			return location.Name, ""
		}
		for _, alias := range location.Aliases {
			if alias.isSimilar(text) && alias.isUnique() {
				for _, city := range location.Cities {
					if city.isSimilar(text) {
						return location.Name, fmt.Sprintf("%s, %s", city.Name, location.Name)
					}
					for _, alias := range city.Aliases {
						if alias.isSimilar(text) {
							return location.Name, fmt.Sprintf("%s, %s", city.Name, location.Name)
						}
					}
				}
				return location.Name, ""
			}
		}
	}
	for _, location := range Locations {
		for _, city := range location.Cities {
			if isFuzzy(city.Name) {
				continue
			}
			if city.isSimilar(text) && city.isUnique() {
				return location.Name, fmt.Sprintf("%s, %s", city.Name, location.Name)
			}
			for _, alias := range city.Aliases {
				if alias.isSimilar(text) {
					return location.Name, fmt.Sprintf("%s, %s", city.Name, location.Name)
				}
			}
		}
	}
	return
}

func isFuzzy(text string) bool {
	return len(text) <= 5
}

func read(name string, v interface{}) {
	b, err := ioutil.ReadFile(fmt.Sprintf("%s/assets/%s/index.json", util.Root(), name))
	if err != nil {
		log.Fatal(err.Error())
	}
	if err = json.Unmarshal(b, &v); err != nil {
		log.Fatal(err.Error())
	}
}
