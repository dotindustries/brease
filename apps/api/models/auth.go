package models

type TokenPair struct {
	AccessToken  string `json:"accessToken" validate:"required"`
	RefreshToken string `json:"refreshToken" validate:"required"`
}

func (*TokenPair) TypeName() string { return "TokenPair" }
