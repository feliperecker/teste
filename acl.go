package core

import (
	"encoding/json"
	"io/ioutil"

	"go.uber.org/zap"
)

type Acl interface {
	GetRoles() map[string]Role
	LoadRoles() error
	Can(permission string, userRoles []string) bool
	SetRole(name string, role Role) error
	GetRole(name string) *Role
	SetRolePermission(name string, permission string, hasAccess bool) error

	SetDisabled(disabled bool) error
}

type NewAclOpts struct {
	Disabled bool
	Logger   *zap.Logger
}

func NewAcl(opts *NewAclOpts) Acl {
	return &DefaultAcl{
		Disabled: opts.Disabled,
		Logger:   opts.Logger,
	}
}

type DefaultAcl struct {
	Disabled  bool
	Logger    *zap.Logger
	RolesList map[string]Role
}

func (a *DefaultAcl) GetRoles() map[string]Role {
	return a.RolesList
}

func (a *DefaultAcl) LoadRoles() error {
	aclFileName := "acl.json"

	var rolesString []byte

	b, err := ioutil.ReadFile(aclFileName)
	if err != nil {
		rolesString = []byte(defaultRoles)
	} else {
		rolesString = b
	}

	return json.Unmarshal(rolesString, &a.RolesList)
}

func (a *DefaultAcl) Can(permission string, userRoles []string) bool {
	if a.Disabled {
		a.Logger.Warn("ACL is disabled: skipping permission check", zap.String("permission", permission), zap.Any("userRoles", userRoles))
		return true
	}
	// first check if user is administrator
	for i := range userRoles {
		if userRoles[i] == "administrator" {
			return true
		}
	}

	for j := range userRoles {
		R := a.RolesList[userRoles[j]]
		if R.Can(permission) {
			return true
		}
	}

	return false
}

func (a *DefaultAcl) SetRole(name string, role Role) error {
	a.RolesList[name] = role
	return nil
}

func (a *DefaultAcl) GetRole(name string) *Role {
	if v, ok := a.RolesList[name]; ok {
		return &v
	}

	return nil
}

func (a *DefaultAcl) SetRolePermission(name string, permission string, hasAccess bool) error {
	role := a.GetRole(name)
	if role == nil {
		return nil
	}

	if hasAccess {
		role.AddPermission(permission)
	} else {
		role.RemovePermission(permission)
	}

	return nil
}

func (a *DefaultAcl) SetDisabled(disabled bool) error {
	a.Disabled = disabled
	return nil
}
