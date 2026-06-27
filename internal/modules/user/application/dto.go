package application

type LoginRequest struct {
	IDToken string `json:"id_token"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	UID  string `json:"uid"`
	Role string `json:"role"`
}
