package models

type AdminPanelReauthInput struct {
	Password string `json:"password" validate:"required,min=1,max=72"`
}
