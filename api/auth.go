package api

// Login is the API message for user login.
type Login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse is the API message for user login response.
type AuthResponse struct {
	UserID   int    `json:"userId"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Token    string `json:"token"`
}
