package operator

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
)

func Unwind(field string) bson.D {
	return bson.D{
		{"$unwind", fmt.Sprintf("$%s", field)},
	}
}

func Match(field string, cond interface{}) bson.D {
	return bson.D{
		{"$match", bson.D{
			{field, cond},
		}},
	}
}

func Project(project interface{}) bson.D {
	return bson.D{
		{"$project", project},
	}
}

func First(field string) bson.D {
	return bson.D{
		{"$first", field},
	}
}

func Sum(field string) bson.D {
	return bson.D{
		{"$sum", fmt.Sprintf("$%s", field)},
	}
}

func Group(group interface{}) bson.D {
	return bson.D{
		{"$group", group},
	}
}

func Sort(field string, order int64) bson.D {
	return bson.D{
		{"$sort", bson.D{
			{field, order},
		}},
	}
}

func Count() bson.D {
	return bson.D{
		{"$count", "count"},
	}
}

func Skip(skip int64) bson.D {
	return bson.D{
		{"$skip", skip},
	}
}

func Limit(limit int64) bson.D {
	return bson.D{
		{"$limit", limit},
	}
}
