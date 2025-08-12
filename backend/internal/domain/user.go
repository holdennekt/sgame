package domain

const SYSTEM = "SYSTEM"

type User struct {
	Id     string  `json:"id" bson:"id" binding:"required"`
	Name   string  `json:"name" bson:"name" binding:"min=1,max=20"`
	Avatar *string `json:"avatar" bson:"avatar" binding:"omitnil,url"`
}

type DbUserDTO struct {
	Login    string `json:"login" bson:"login" binding:"min=4,max=20"`
	Password string `json:"password" bson:"password" binding:"min=8,max=40"`
}

type DbUser struct {
	User      `bson:"inline"`
	DbUserDTO `bson:"inline"`
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
