package mediaservice_test

import (
	"testing"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/rclonefs"
	rr "github.com/fastschema/fastschema/pkg/restresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	ms "github.com/fastschema/fastschema/services/media"
	"github.com/stretchr/testify/assert"
)

type testApp struct {
	sb    *schema.Builder
	db    app.DBClient
	disks []app.Disk
}

func (s testApp) DB() app.DBClient {
	return s.db
}

func (s testApp) Disk(names ...string) app.Disk {
	return s.disks[0]
}

func createMediaService(t *testing.T) (*ms.MediaService, *rr.Server) {
	sb := utils.Must(schema.NewBuilderFromDir(t.TempDir()))
	disks := utils.Must(rclonefs.NewFromConfig([]*app.DiskConfig{{
		Name:    "local_test",
		Driver:  "local",
		Root:    t.TempDir(),
		BaseURL: "http://localhost:3000/files",
	}}, t.TempDir()))

	testApp := &testApp{sb: sb, disks: disks}
	mediaService := ms.New(testApp)
	testApp.db = utils.Must(entdbadapter.NewTestClient(t.TempDir(), sb, &app.Hooks{
		AfterDBContentList: []app.AfterDBContentListHook{mediaService.MediaListHook},
	}))
	resources := app.NewResourcesManager()
	resources.Middlewares = append(resources.Middlewares, func(c app.Context) error {
		c.Value("user", &app.User{ID: 1})
		return c.Next()
	})
	resources.Group("media").
		Add(app.NewResource("upload", mediaService.Upload, app.Meta{app.POST: "/upload"})).
		Add(app.NewResource("delete", mediaService.Delete, app.Meta{app.DELETE: ""}))
	assert.NoError(t, resources.Init())
	restResolver := rr.NewRestResolver(resources).Init(app.CreateMockLogger(true))

	return mediaService, restResolver.Server()
}

func TestNewMediaService(t *testing.T) {
	service, server := createMediaService(t)
	assert.NotNil(t, service)
	assert.NotNil(t, server)
}
