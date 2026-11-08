package auth

import "github.com/artsgoz/artsgoz-backend/internal/user"

type FirebaseLoginRequest struct {
	IDToken  string `json:"id_token" binding:"required"`
	Username string `json:"username"`
}

type LoginResponse struct {
	Token string     `json:"token"`
	User  *user.User `json:"user"`
}
