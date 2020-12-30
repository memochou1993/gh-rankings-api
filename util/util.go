package util

import (
	"bytes"
	"encoding/json"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func LoadEnv() {
	viper.AddConfigPath("./")
	viper.SetConfigName(os.Getenv("APP_ENV"))
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalln(err.Error())
	}
}

func ParseStruct(v interface{}, sep string) string {
	b := bytes.Buffer{}
	encoder := json.NewEncoder(&b)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(v); err != nil {
		log.Fatalln(err.Error())
	}

	s := b.String()
	s = strings.Replace(s, "\n", "", -1)
	s = strings.Replace(s, "\\\"", "#", -1)
	s = strings.Replace(s, "\"", "", -1)
	s = strings.Replace(s, "#", "\"", -1)
	s = strings.Replace(s, ",", sep, -1)
	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")

	return s
}

func Languages() (languages []string) {
	b, err := ioutil.ReadFile("./assets/languages.json")
	if err != nil {
		log.Fatalln(err.Error())
	}
	if err = json.Unmarshal(b, &languages); err != nil {
		log.Fatalln(err.Error())
	}
	return languages
}

func Locations() (locations []string) {
	b, err := ioutil.ReadFile("./assets/locations.json")
	if err != nil {
		log.Fatalln(err.Error())
	}
	if err = json.Unmarshal(b, &locations); err != nil {
		log.Fatalln(err.Error())
	}
	return locations
}
