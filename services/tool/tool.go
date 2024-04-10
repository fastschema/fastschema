package toolservice

import (
	"github.com/fastschema/fastschema/app"
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

func (s *ToolService) Stats(c app.Context, _ *any) (_ *StatsData, err error) {
	totalSchemas := len(s.DB().SchemaBuilder().Schemas())
	totalUsers := 0
	totalRoles := 0
	totalMedias := 0

	userModel, err := s.DB().Model("user")
	if err != nil {
		return nil, err
	}

	roleModel, err := s.DB().Model("role")
	if err != nil {
		return nil, err
	}

	mediaModel, err := s.DB().Model("media")
	if err != nil {
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
