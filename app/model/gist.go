package model

type Gist struct {
	Name       string    `json:"name" bson:"name"`
	Stargazers Directory `json:"stargazers" bson:"stargazers"`
}
