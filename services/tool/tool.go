package toolservice

import (
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/utils"
)

type AppLike interface {
	DB() app.DBClient
}

type ToolService struct {
	DB func() app.DBClient
}

func New(app AppLike) *ToolService {
	return &ToolService{
		DB: app.DB,
	}
}

type StatsData struct {
	TotalSchemas int `json:"totalSchemas"`
	TotalUsers   int `json:"totalUsers"`
	TotalRoles   int `json:"totalRoles"`
	TotalMedias  int `json:"totalMedias"`
}

func (s *ToolService) Stats(c app.Context, _ any) (_ *StatsData, err error) {
	totalSchemas := len(s.DB().SchemaBuilder().Schemas())
	totalUsers := 0
	totalRoles := 0
	totalMedias := 0

	userModel, userModelErr := s.DB().Model("user")
	roleModel, roleModelErr := s.DB().Model("role")
	mediaModel, modelModelErr := s.DB().Model("media")

	if err = utils.MergeErrorMessages(userModelErr, roleModelErr, modelModelErr); err != nil {
		return nil, err
	}

	if totalUsers, err = userModel.Query().Count(nil); err != nil {
		return nil, err
	}

	if totalRoles, err = roleModel.Query().Count(nil); err != nil {
		return nil, err
	}

	if totalMedias, err = mediaModel.Query().Count(nil); err != nil {
		return nil, err
	}

	return &StatsData{
		TotalSchemas: totalSchemas,
		TotalUsers:   totalUsers,
		TotalRoles:   totalRoles,
		TotalMedias:  totalMedias,
	}, nil
}
