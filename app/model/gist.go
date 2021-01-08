package model

type Gist struct {
	Forks      *Directory `json:"forks,omitempty" bson:"forks,omitempty"`
	Name       string     `json:"name,omitempty" bson:"name,omitempty"`
	Stargazers *Directory `json:"stargazers,omitempty" bson:"stargazers,omitempty"`
}
