package toolservice

import (
	"github.com/fastschema/fastschema/app"
	"golang.org/x/sync/errgroup"
)

type ToolService struct {
	app app.App
}

func NewToolService(
	app app.App,
) *ToolService {
	return &ToolService{
		app: app,
	}
}

type StatsData struct {
	TotalSchemas int `json:"totalSchemas"`
	TotalUsers   int `json:"totalUsers"`
	TotalRoles   int `json:"totalRoles"`
	TotalMedias  int `json:"totalMedias"`
}

func (s *ToolService) Stats(c app.Context, _ *any) (_ *StatsData, err error) {
	var errGroup errgroup.Group
	totalSchemas := len(s.app.SchemaBuilder().Schemas())
	totalUsers := 0
	totalRoles := 0
	totalMedias := 0

	errGroup.Go(func() error {
		userModel, err := s.app.DB().Model("user")
		if err != nil {
			return err
		}

		totalUsers, err = userModel.Query().Count(nil)
		return err
	})

	errGroup.Go(func() error {
		roleModel, err := s.app.DB().Model("role")
		if err != nil {
			return err
		}

		totalRoles, err = roleModel.Query().Count(nil)
		return err
	})

	errGroup.Go(func() error {
		mediaModel, err := s.app.DB().Model("media")
		if err != nil {
			return err
		}

		totalMedias, err = mediaModel.Query().Count(nil)
		return err
	})

	if err := errGroup.Wait(); err != nil {
		return nil, err
	}

	return &StatsData{
		TotalSchemas: totalSchemas,
		TotalUsers:   totalUsers,
		TotalRoles:   totalRoles,
		TotalMedias:  totalMedias,
	}, nil
}
