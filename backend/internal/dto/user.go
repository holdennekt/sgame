package dto

type CreateUserRequest struct {
	Login    string `json:"login" bson:"login" binding:"min=4,max=20"`
	Password string `json:"password" bson:"password" binding:"min=8,max=40"`
}

type CreateUserResponse struct {
	Id string `json:"id" example:"507f1f77bcf86cd799439011"`
}

type UpdateUserRequest struct {
	Password string  `json:"password,omitempty" binding:"min=8,max=40"`
	Name     string  `json:"name,omitempty" binding:"min=1,max=20"`
	Avatar   *string `json:"avatar,omitempty"`
}

type AuthResponse struct {
	UserId string `json:"userId" example:"507f1f77bcf86cd799439011"`
}
