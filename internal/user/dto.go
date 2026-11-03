package user

type UpdateProfileRequest struct {
	Email string `json:"email"`
}

type UpdateRoleRequest struct {
	Role string `json:"role" binding:"required"`
}
