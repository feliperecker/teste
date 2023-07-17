package core_test

import (
	"reflect"
	"testing"

	"github.com/go-bolo/core"
	"github.com/stretchr/testify/assert"
)

func TestRole(t *testing.T) {
	r, _ := core.NewRole(&core.NewRoleOpts{Name: "editor"})
	assert.Equal(t, 0, len(r.Permissions))
	assert.False(t, r.Can("find_image"))

	r.AddPermission("upload_image")
	r.AddPermission("find_image")

	assert.Equal(t, 2, len(r.Permissions))
	assert.True(t, r.Can("find_image"))

	r.RemovePermission("find_image")

	assert.Equal(t, 1, len(r.Permissions))
	assert.False(t, r.Can("find_image"))
}

func TestNewRole(t *testing.T) {
	type args struct {
		opts *core.NewRoleOpts
	}
	tests := []struct {
		name    string
		args    args
		want    *core.Role
		wantErr bool
	}{
		{
			"success empty",
			args{opts: &core.NewRoleOpts{
				Name: "faxineira",
			}},
			&core.Role{
				Name: "faxineira",
			},
			false,
		},
		{
			"error no name",
			args{opts: &core.NewRoleOpts{}},
			nil,
			true,
		},
		{
			"success with permissions",
			args{opts: &core.NewRoleOpts{
				Name:        "porteiro",
				Permissions: []string{"block-user-access"},
			}},
			&core.Role{
				Name:        "porteiro",
				Permissions: []string{"block-user-access"},
			},
			false,
		},
		{
			"success with systemRole",
			args{opts: &core.NewRoleOpts{
				Name:         "editor",
				IsSystemRole: true,
			}},
			&core.Role{
				Name:         "editor",
				IsSystemRole: true,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := core.NewRole(tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRole_Can(t *testing.T) {
	editoRole := core.Role{
		Name:          "editor",
		Permissions:   []string{"update_image", "find_image", "create_content"},
		CanAddInUsers: true,
		IsSystemRole:  true,
	}

	type args struct {
		permission string
	}
	tests := []struct {
		name string
		r    *core.Role
		args args
		want bool
	}{
		{
			"can find_image",
			&editoRole,
			args{permission: "find_image"},
			true,
		},
		{
			"cant delete_content",
			&editoRole,
			args{permission: "delete_content"},
			false,
		},
		{
			"cant create_content",
			&editoRole,
			args{permission: "create_content"},
			true,
		},
		{
			"lixeiro cant jogar_lixo_na_rua",
			&core.Role{
				Name:          "lixeiro",
				Permissions:   []string{"pegar_lixo"},
				CanAddInUsers: true,
				IsSystemRole:  false,
			},
			args{permission: "jogar_lixo_na_rua"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Can(tt.args.permission); got != tt.want {
				t.Errorf("Role.Can() = %v, want %v", got, tt.want)
			}
		})
	}
}
