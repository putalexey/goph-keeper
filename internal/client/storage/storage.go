package storage

type Storager interface {
	Load() error
	Save() error
	SetToken(token string)
	GetToken() string
	SetLogin(login string)
	GetLogin() string
}
