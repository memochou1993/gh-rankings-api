package collection

type Users struct {
	Data struct {
		Search struct {
			UserCount int `json:"userCount"`
			Edges     []struct {
				Cursor string `json:"cursor"`
				Node   struct {
					Login string `json:"login"`
				} `json:"node"`
			} `json:"edges"`
			PageInfo struct {
				EndCursor   string `json:"endCursor"`
				HasNextPage bool   `json:"hasNextPage"`
				StartCursor string `json:"startCursor"`
			} `json:"pageInfo"`
		} `json:"search"`
	} `json:"data"`
}
