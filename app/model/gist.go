package model

type Gist struct {
	Forks      *Items `json:"forks,omitempty" bson:"forks,omitempty"`
	Name       string `json:"name,omitempty" bson:"name,omitempty"`
	Stargazers *Items `json:"stargazers,omitempty" bson:"stargazers,omitempty"`
}
