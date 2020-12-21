package model

type Gist struct {
	Forks      Directory `json:"forks" bson:"forks"`
	Name       string    `json:"name" bson:"name"`
	Stargazers Directory `json:"stargazers" bson:"stargazers"`
}
