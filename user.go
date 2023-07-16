package core

type User interface {
	GetID() string
	SetID(id string) error
	GetDisplayName() string
	SetDisplayName(name string) error
}
