package roleservice

import (
	"github.com/fastschema/fastschema/app"
)

type RoleService struct {
	app app.App
}

func NewRoleService(fsApp app.App) *RoleService {
	return &RoleService{
		app: fsApp,
	}
}
