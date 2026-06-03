package domain

const SYSTEM = "SYSTEM"

type User struct {
	Id      string  `json:"id" bson:"id"`
	Name    string  `json:"name" bson:"name"`
	Avatar  *string `json:"avatar" bson:"avatar"`
	IsGuest bool    `json:"isGuest" bson:"isGuest"`
}

type DbUser struct {
	User     `bson:"inline"`
	Login    string `json:"login" bson:"login"`
	Password string `json:"password" bson:"password"`
}

type Host struct {
	User        `bson:"inline"`
	IsConnected bool `json:"isConnected" bson:"isConnected"`
}

type Player struct {
	User        `bson:"inline"`
	Score       int  `json:"score" bson:"score"`
	BetAmount   *int `json:"betAmount" bson:"betAmount"`
	IsConnected bool `json:"isConnected" bson:"isConnected"`
}
