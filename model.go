package bolo

type Model interface {
	GetID() string
	LoadTeaserData() (err error)
	LoadData() (err error)
	Save(app App) error
}
