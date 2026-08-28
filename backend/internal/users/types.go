package users

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	AvatarURL string `json:"avatar_url"`
}

type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserResponse struct {
	Message string `json:"message"`
	User    User   `json:"user"`
}

type UsersResponse struct {
	Users []User `json:"users"`
}
