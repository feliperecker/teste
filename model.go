package bolo

type Model interface {
	GetID() string
	LoadTeaser() (err error)
	LoadData() (err error)
	Save(app App) error
}
