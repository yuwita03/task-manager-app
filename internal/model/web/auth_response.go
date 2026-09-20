package web

type AuthResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

type UserResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}