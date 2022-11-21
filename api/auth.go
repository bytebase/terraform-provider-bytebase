package api

// AuthResponse is the API message for user login.
type AuthResponse struct {
	UserID   int    `json:"userId"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Token    string `json:"token"`
}
