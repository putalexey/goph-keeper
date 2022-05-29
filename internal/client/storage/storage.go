package storage

import "github.com/putalexey/goph-keeper/internal/common/models"

type Storager interface {
	Load() error
	Save() error
	SetToken(token string)
	GetToken() string
	SetLogin(login string)
	GetLogin() string
}

var SupportedTypes = []string{
	models.TypeText,
	models.TypeFile,
	models.TypeLogin,
	models.TypeCard,
}
