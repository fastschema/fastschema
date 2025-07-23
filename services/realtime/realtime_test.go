package realtimeservice_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	rs "github.com/fastschema/fastschema/services/realtime"
	"github.com/stretchr/testify/assert"
)

type testApp struct {
	sb           *schema.Builder
	db           db.Client
	logger       *logger.MockLogger
	restResolver *restfulresolver.RestfulResolver
}

func (s testApp) DB() db.Client {
	return s.db
}

func (s testApp) Logger() logger.Logger {
	return s.logger
}

func createTestApp(t *testing.T) (*testApp, *rs.RealtimeService) {
	schemaDir := utils.Must(os.MkdirTemp("", "schema"))
	assert.NoError(t, utils.WriteFile(schemaDir+"/blog.json", `{
		"name": "blog",
		"namespace": "blogs",
		"label_field": "name",
		"fields": [
			{
				"type": "string",
				"name": "name",
				"label": "Name",
				"sortable": true
			}
		]
	}`))

	sb := utils.Must(schema.NewBuilderFromDir(schemaDir, fs.SystemSchemaTypes...))
	app := &testApp{
		sb:     sb,
		logger: logger.CreateMockLogger(true),
	}

	realtimeService := rs.New(app)
	resources := fs.NewResourcesManager()
	resources.Group("api").
		Group("realtime").
		Add(fs.NewResource("content", realtimeService.Content, &fs.Meta{WS: "/content"}))

	app.db = utils.Must(entdbadapter.NewTestClient(
		utils.Must(os.MkdirTemp("", "migrations")),
		app.sb,
		func() *db.Hooks {
			return &db.Hooks{
				PostDBCreate: []db.PostDBCreate{realtimeService.ContentCreateHook},
				PostDBUpdate: []db.PostDBUpdate{realtimeService.ContentUpdateHook},
				PostDBDelete: []db.PostDBDelete{realtimeService.ContentDeleteHook},
			}
		},
	))

	assert.NoError(t, resources.Init())
	app.restResolver = restfulresolver.NewRestfulResolver(&restfulresolver.ResolverConfig{
		ResourceManager: resources,
		Logger:          app.logger,
	})

	return app, realtimeService
}

type mockWSClient struct {
	id                string
	expectMethodError string
	message           chan []byte
}

func newMockWSClient(expectMethodErrors ...string) *mockWSClient {
	expectMethodErrors = append(expectMethodErrors, "")
	return &mockWSClient{
		expectMethodError: expectMethodErrors[0],
		message:           make(chan []byte),
		id:                utils.RandomString(10),
	}
}

func (m *mockWSClient) ID() string {
	return m.id
}

func (m *mockWSClient) Write(message []byte, messageTypes ...fs.WSMessageType) error {
	if m.expectMethodError == "write" {
		return fmt.Errorf("write error")
	}

	m.message <- message

	return nil
}

func (m *mockWSClient) Read() (messageType fs.WSMessageType, message []byte, err error) {
	if m.expectMethodError == "read" {
		return 0, nil, fmt.Errorf("read error")
	}
	return fs.WSMessageText, <-m.message, nil
}

func (m *mockWSClient) Close(msgs ...string) error {
	if m.expectMethodError == "close" {
		return fmt.Errorf("close error")
	}
	return nil
}

func (m *mockWSClient) IsCloseNormal(err error) bool {
	return (&restfulresolver.WSClient{}).IsCloseNormal(err)
}

func TestNew(t *testing.T) {
	_, service := createTestApp(t)
	assert.NotNil(t, service)
}

func TestAddRemoveClient(t *testing.T) {
	_, service := createTestApp(t)

	client1 := newMockWSClient()
	client2 := newMockWSClient()
	service.AddClient(client1, "testtopic", nil)
	service.AddClient(client2, "testtopic", nil)
	testTopic, ok := service.Topics().Load("testtopic")
	assert.True(t, ok)
	assert.Equal(t, 2, testTopic.Len())

	assert.NoError(t, service.RemoveClient(client1))
	testTopic, ok = service.Topics().Load("testtopic")
	assert.True(t, ok)
	assert.Equal(t, 1, testTopic.Len())

	assert.NoError(t, service.RemoveClient(client2, true))
	_, ok = service.Topics().Load("testtopic")
	assert.False(t, ok)
}

type testSerializer struct {
	errorMethod string
	data        []byte
}

func newTestSerializer(errorMethods ...string) *testSerializer {
	errorMethods = append(errorMethods, "")
	return &testSerializer{errorMethod: errorMethods[0]}
}

func (ts *testSerializer) Serialize(data any) ([]byte, error) {
	if ts.errorMethod == "serialize" {
		return nil, fmt.Errorf("serialize error")
	}
	return ts.data, nil
}

func TestBroadCast(t *testing.T) {
	app, service := createTestApp(t)

	// Broadcast non existing topic
	service.Broadcast([]string{"nonexisting"}, nil)

	// Broadcast on nil serializer
	client1 := newMockWSClient()
	service.AddClient(client1, "testtopic", nil)
	service.Broadcast([]string{"testtopic"}, nil)

	// Broadcast on closing client
	client2 := newMockWSClient()
	service.AddClient(client2, "testtopic", &testSerializer{})
	service.Broadcast([]string{"testtopic"}, nil)

	// Broadcast serializer return empty message
	client3 := newMockWSClient()
	serializer := newTestSerializer()
	serializer.data = nil
	service.AddClient(client3, "testtopic", serializer)
	service.Broadcast([]string{"testtopic"}, nil)

	// Broascast with serializer error
	client4 := newMockWSClient()
	go func() {
		msg := <-client4.message
		assert.Contains(t, string(msg), "failed to")
	}()

	serializer = newTestSerializer("serialize")
	service.AddClient(client4, "testtopic", serializer)
	service.Broadcast([]string{"testtopic"}, nil)
	time.Sleep(100 * time.Millisecond)
	assert.Contains(t, app.logger.Last().String(), "failed to")

	// Broadcast on client with write error
	client5 := newMockWSClient("write")
	go func() {
		msg := <-client5.message
		assert.Contains(t, string(msg), "failed to write message to client")
	}()
	serializer = &testSerializer{}
	serializer.data = []byte("test")
	service.AddClient(client5, "testtopic", serializer)
	service.Broadcast([]string{"testtopic"}, nil)
	time.Sleep(100 * time.Millisecond)
	assert.Contains(t, app.logger.Last().String(), "failed to write message to client")

	// Broadcast on client with normal write
	client6 := newMockWSClient()
	go func() {
		msg := <-client6.message
		assert.Equal(t, "test", string(msg))
	}()
	serializer = &testSerializer{}
	serializer.data = []byte("test")
	service.AddClient(client6, "testtopic", serializer)
	service.Broadcast([]string{"testtopic"}, nil)
}

func TestCreateResource(t *testing.T) {
	_, service := createTestApp(t)
	api := fs.NewResourcesManager().Group("api")
	service.CreateResource(api)
	assert.NotNil(t, api.Find("api.realtime.content"))
}
