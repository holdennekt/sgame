package domain

const SYSTEM = "SYSTEM"

type User struct {
	Id      string  `json:"id"`
	Name    string  `json:"name"`
	Avatar  *string `json:"avatar"`
	IsGuest bool    `json:"isGuest"`
}

type DbUser struct {
	User
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Host struct {
	User
	IsConnected bool `json:"isConnected" bson:"isConnected"`
}

type Player struct {
	User
	Score       int  `json:"score" bson:"score"`
	BetAmount   *int `json:"betAmount" bson:"betAmount"`
	IsConnected bool `json:"isConnected" bson:"isConnected"`
}
