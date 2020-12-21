package model

type Repository struct {
	Forks           Directory `json:"forks" bson:"forks"`
	Name            string    `json:"name" bson:"name"`
	NameWithOwner   string    `json:"nameWithOwner" bson:"_id"`
	Owner           `json:"owner" bson:"owner"`
	PrimaryLanguage struct {
		Name string `json:"name" bson:"name"`
	} `json:"primaryLanguage" bson:"primary_language"`
	Stargazers Directory `json:"stargazers" bson:"stargazers"`
	Watchers   Directory `json:"watchers" bson:"watchers"`
}

type RepositoryResponse struct {
	Data struct {
		Search struct {
			Edges []struct {
				Cursor string     `json:"cursor"`
				Node   Repository `json:"node"`
			} `json:"edges"`
			PageInfo `json:"pageInfo"`
		} `json:"search"`
		RateLimit `json:"rateLimit"`
	} `json:"data"`
	Errors []Error `json:"errors"`
}

type RepositoryRank struct {
	NameWithOwner string `bson:"_id"`
	TotalCount    int    `bson:"total_count"`
}

type RepositoryModel struct {
	*Model
}

func NewRepositoryModel() *RepositoryModel {
	return &RepositoryModel{
		&Model{
			name: "repositories",
		},
	}
}
