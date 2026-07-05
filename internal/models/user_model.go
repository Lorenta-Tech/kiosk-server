package models

import "time"

type User struct {
	ID        string
	GoogleID  string
	Email     string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type GoogleAuthRequest struct {
	IDToken string `json:"id_token" validate:"required"`
}

type GoogleClaims struct {
	Sub   string
	Email string
	Name  string
}

// Response
type AuthResponse struct {
	Token string   `json:"token"`
	User  UserInfo `json:"user"`
}

type UserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}
