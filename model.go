package core

type Model interface {
	FindOne(app *App) (any, err error)
	Query(app *App) (any, err error)
}

type Record interface {
	LoadTeaser() (err error)
	LoadData() (err error)
}
