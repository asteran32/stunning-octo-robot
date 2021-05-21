package model

type User struct {
	Id          string
	LastName    string   `json:"lastname" bson:"lastname"`
	FirstName   string   `json:"firstname" bson:"firstname"`
	Email       string   `json:"email" bson:"email"`
	Password    string   `json:"-" bson:"password"`
	Roles       []string `json:"roles"`
	AccessToken string   `json:"accessToken"`
}
