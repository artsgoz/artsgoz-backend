package auth

type LoginRequest struct {
	IDToken string `json:"id_token"`
}

type LoginResponse struct {
	UID                string `json:"uid"`
	Role               string `json:"role,omitempty"`
	OnboardingRequired bool   `json:"onboarding_required"`
}
