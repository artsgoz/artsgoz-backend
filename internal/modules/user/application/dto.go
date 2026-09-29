package application

type LoginRequest struct {
	IDToken string `json:"id_token"`
}

type RegisterRequest struct {
	IDToken string `json:"id_token"`
	Role    string `json:"role,omitempty"`
}

type LoginResponse struct {
	UID  string `json:"uid"`
	Role string `json:"role"`
}
