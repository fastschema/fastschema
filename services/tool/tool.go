package toolservice

import (
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
)

type AppLike interface {
	DB() db.Client
}

type ToolService struct {
	DB func() db.Client
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
	TotalFiles   int `json:"totalFiles"`
}

func (s *ToolService) Stats(c fs.Context, _ any) (_ *StatsData, err error) {
	totalSchemas := len(s.DB().SchemaBuilder().Schemas())
	totalUsers := 0
	totalRoles := 0
	totalFiles := 0

	if totalUsers, err = db.Builder[*fs.User](s.DB()).Count(c); err != nil {
		return nil, err
	}

	if totalRoles, err = db.Builder[*fs.Role](s.DB()).Count(c); err != nil {
		return nil, err
	}

	if totalFiles, err = db.Builder[*fs.File](s.DB()).Count(c); err != nil {
		return nil, err
	}

	return &StatsData{
		TotalSchemas: totalSchemas,
		TotalUsers:   totalUsers,
		TotalRoles:   totalRoles,
		TotalFiles:   totalFiles,
	}, nil
}
