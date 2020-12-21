package model

type Repository struct {
	Forks           Directory `json:"forks" bson:"forks"`
	Name            string    `json:"name" bson:"name"`
	PrimaryLanguage struct {
		Name string `json:"name" bson:"name"`
	} `json:"primaryLanguage" bson:"primary_language"`
	Stargazers Directory `json:"stargazers" bson:"stargazers"`
	Watchers   Directory `json:"watchers" bson:"watchers"`
}
