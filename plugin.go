package bolo

type Plugin interface {
	Init(app App) error

	GetName() string
	SetName(name string) error
}
