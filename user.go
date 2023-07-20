package bolo

type User interface {
	GetID() string
	SetID(id string) error
	GetDisplayName() string
	SetDisplayName(name string) error
	GetRoles() []string
	SetRoles(v []string) error
	AddRole(role string) error
	RemoveRole(role string) error
	GetEmail() string
	SetEmail(v string) error
	GetUsername() string
	SetUsername(v string) error
	GetFullName() string
	SetFullName(v string) error
	GetLanguage() string
	SetLanguage(v string) error
	IsActive() bool
	SetActive(blocked bool) error
	IsBlocked() bool
	SetBlocked(blocked bool) error
	// FillById
	FillById(ID string) error
}
