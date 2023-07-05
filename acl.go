package core

import (
	"encoding/json"
	"io/ioutil"
)

type Acl interface {
	GetRoles() map[string]Role
	LoadRoles() error
}

func NewAcl() Acl {
	return &DefaultAcl{}
}

type DefaultAcl struct {
	RolesList map[string]Role
}

func (a *DefaultAcl) GetRoles() map[string]Role {
	return a.RolesList
}

func (a *DefaultAcl) LoadRoles() error {
	aclFileName := "acl.json"

	rolesString := []byte{}

	b, err := ioutil.ReadFile(aclFileName)
	if err != nil {
		rolesString = []byte(defaultRoles)
	} else {
		rolesString = b
	}

	return json.Unmarshal(rolesString, &a.RolesList)
}
