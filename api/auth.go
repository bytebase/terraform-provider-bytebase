package api

// Login is the API message for user login.
type Login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse is the API message for user login response.
type AuthResponse struct {
	Token string `json:"token"`
}
