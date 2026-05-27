package dto

type CreateUserRequest struct {
	Login    string `json:"login" bson:"login" binding:"min=4,max=20"`
	Password string `json:"password" bson:"password" binding:"min=8,max=40"`
}

type CreateUserResponse struct {
	Id string `json:"id" example:"507f1f77bcf86cd799439011"`
}

type UpdateUserRequest struct {
	Name     string  `json:"name" binding:"required,min=1,max=20"`
	Avatar   *string `json:"avatar"`
	Password string  `json:"password" binding:"omitempty,min=8,max=40"`
}

type AuthResponse struct {
	UserId string `json:"userId" example:"507f1f77bcf86cd799439011"`
}

type GuestLoginRequest struct {
	Name string `json:"name" binding:"required,min=1,max=50"`
}
