package model

type Csvs struct {
	Items []Csv `json:"csv"`
}

type Csv struct {
	Id   string `json:"-"`
	Name string `json:"name" bson:"name"`
	Path string `json:"path" bson:"path"`
	Date string `json:"date" bson:"date"`
}
