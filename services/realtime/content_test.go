package realtimeservice_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	fhws "github.com/fasthttp/websocket"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	rs "github.com/fastschema/fastschema/services/realtime"
	"github.com/gofiber/contrib/websocket"
	"github.com/stretchr/testify/assert"
)

func createTestAppAndListen(t *testing.T) (*testApp, *rs.RealtimeService) {
	app, service := createTestApp(t)

	go func() {
		assert.NoError(t, app.restResolver.Server().Listen("localhost:55555"))
	}()

	readyCh := make(chan struct{})

	go func() {
		for {
			conn, err := net.Dial("tcp", "localhost:55555")
			if err != nil {
				continue
			}

			if conn != nil {
				readyCh <- struct{}{}
				conn.Close()
				break
			}
		}
	}()

	<-readyCh

	return app, service
}

func dial(urlStr string, requestHeader http.Header) (*fhws.Conn, *http.Response, error) {
	conn, resp, err := fhws.DefaultDialer.Dial(urlStr, requestHeader)
	time.Sleep(10 * time.Millisecond)

	return conn, resp, err
}

type eventSingleData struct {
	Event string         `json:"event"`
	Data  map[string]any `json:"data"`
}

type eventMultipleData struct {
	Event string           `json:"event"`
	Data  []map[string]any `json:"data"`
}

func testConcurrent(t *testing.T, url, topic string, times int, service *rs.RealtimeService) {
	var wg sync.WaitGroup
	for i := 0; i < times; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			conn, resp, err := dial(url, nil)

			defer resp.Body.Close()
			defer conn.Close()
			assert.NoError(t, err)

			_, ok := service.Topics().Load(topic)
			assert.True(t, ok)
		}(i)
	}
	wg.Wait()
}

func TestRealtimeContent(t *testing.T) {
	app, service := createTestAppAndListen(t)

	defer func() {
		assert.NoError(t, app.restResolver.Server().Shutdown())
	}()

	testConcurrent(t, "ws://localhost:55555/api/realtime/content?schema=blog", "content.blog", 10, service)
	testConcurrent(t, "ws://localhost:55555/api/realtime/content?schema=blog&event=create", "content.blog.create", 10, service)
	testConcurrent(t, "ws://localhost:55555/api/realtime/content?schema=blog&event=update&id=5", "content.blog.update.5", 10, service)

	// Connect to websocket
	_, resp1, err1 := dial("ws://localhost:55555/api/realtime/content", nil)
	resp1.Body.Close()
	assert.NoError(t, err1)
	assert.Equal(t, 101, resp1.StatusCode)
	assert.Equal(t, "websocket", resp1.Header.Get("Upgrade"))

	// Connect with invalid event
	conn2, resp2, err2 := dial("ws://localhost:55555/api/realtime/content?event=invalid", nil)
	resp2.Body.Close()
	assert.NoError(t, err2)
	_, _, err2 = conn2.ReadMessage()
	assert.Contains(t, err2.Error(), "realtime.content: invalid event")

	// Connect with invalid schema
	conn3, resp3, err3 := dial("ws://localhost:55555/api/realtime/content?schema=invalid", nil)
	resp3.Body.Close()
	assert.NoError(t, err3)
	_, _, err3 = conn3.ReadMessage()
	assert.Contains(t, err3.Error(), "realtime.content: schema not found")

	// Connect with invalid filter
	conn4, resp4, err4 := dial("ws://localhost:55555/api/realtime/content?schema=blog&filter=invalid", nil)
	resp4.Body.Close()
	assert.NoError(t, err4)
	_, _, err4 = conn4.ReadMessage()
	assert.Contains(t, err4.Error(), "realtime.content: invalid filter")

	// Success cases:
	// Connect successfully with all events
	conn5, resp5, err5 := dial("ws://localhost:55555/api/realtime/content?schema=blog", nil)
	assert.NoError(t, resp5.Body.Close())
	assert.NoError(t, err5)
	_, ok := service.Topics().Load("content.blog")
	assert.True(t, ok)
	assert.NoError(t, conn5.WriteMessage(fhws.CloseMessage, websocket.FormatCloseMessage(1000, "close")))

	// Connect successfully with event create
	conn6, resp6, err6 := dial("ws://localhost:55555/api/realtime/content?schema=blog&event=create", nil)
	assert.NoError(t, resp6.Body.Close())
	assert.NoError(t, err6)
	_, ok = service.Topics().Load("content.blog.create")
	assert.True(t, ok)
	assert.NoError(t, conn6.WriteMessage(fhws.CloseMessage, websocket.FormatCloseMessage(1000, "close")))

	// Connect successfully with event update on specific ID
	conn7, resp7, err7 := dial("ws://localhost:55555/api/realtime/content?schema=blog&event=update&id=5", nil)
	assert.NoError(t, resp7.Body.Close())
	assert.NoError(t, err7)
	_, ok = service.Topics().Load("content.blog.update.5")
	assert.True(t, ok)
	assert.NoError(t, conn7.WriteMessage(fhws.CloseMessage, websocket.FormatCloseMessage(1000, "close")))

	// Connect and send message
	conn8, resp8, err8 := dial("ws://localhost:55555/api/realtime/content?schema=blog", nil)
	assert.NoError(t, resp8.Body.Close())

	assert.NoError(t, err8)
	go func() {
		mt, msg, err := conn8.ReadMessage()
		assert.NoError(t, err)
		assert.Equal(t, fhws.TextMessage, mt)
		assert.Equal(t, []byte("test message"), msg)
	}()
	assert.NoError(t, conn8.WriteMessage(fhws.TextMessage, []byte("test message")))
	assert.NoError(t, conn8.WriteMessage(fhws.CloseMessage, websocket.FormatCloseMessage(1000, "close")))

	model := utils.Must(app.DB().Model("blog"))

	// Connect and subscribe to the create event
	conn9, resp9, err9 := dial("ws://localhost:55555/api/realtime/content?schema=blog&event=create&select=id,name", nil)
	assert.NoError(t, resp9.Body.Close())
	assert.NoError(t, err9)
	ctx := context.Background()
	time.Sleep(10 * time.Millisecond)
	go func() {
		createdID, err9 := model.Create(ctx, schema.NewEntity().Set("name", "test"))
		assert.NoError(t, err9)
		assert.Greater(t, createdID, uint64(0))
	}()

	data := eventSingleData{}
	assert.NoError(t, conn9.ReadJSON(&data))
	assert.Equal(t, "create", data.Event)
	assert.Equal(t, "test", data.Data["name"])
	assert.NoError(t, conn9.WriteMessage(fhws.CloseMessage, websocket.FormatCloseMessage(1000, "close")))

	// Connect and subscribe to the update event
	conn10, resp10, err10 := dial("ws://localhost:55555/api/realtime/content?schema=blog&event=update", nil)
	assert.NoError(t, err10)
	assert.NoError(t, resp10.Body.Close())
	time.Sleep(10 * time.Millisecond)
	go func() {
		createdID, err10 := model.Create(ctx, schema.NewEntity().Set("name", "test2"))
		assert.NoError(t, err10)

		_, err10 = model.Mutation().Where(db.EQ("id", createdID)).Update(ctx, schema.NewEntity().Set("name", "test2 updated"))
		assert.NoError(t, err10)
	}()

	data2 := eventMultipleData{}
	assert.NoError(t, conn10.ReadJSON(&data2))
	assert.Equal(t, "update", data2.Event)
	assert.Equal(t, "test2 updated", data2.Data[0]["name"])
	assert.NoError(t, conn10.WriteMessage(fhws.CloseMessage, websocket.FormatCloseMessage(1000, "close")))

	// Connect and subscribe to the update event with specific ID
	blogID := utils.Must(model.Create(ctx, schema.NewEntity().Set("name", "test3")))
	conn11, resp11, err11 := dial(fmt.Sprintf("ws://localhost:55555/api/realtime/content?schema=blog&event=update&id=%d", blogID), nil)
	assert.NoError(t, err11)
	assert.NoError(t, resp11.Body.Close())
	go func() {
		_, err11 = model.Mutation().Where(db.EQ("id", blogID)).Update(ctx, schema.NewEntity().Set("name", "test3 updated"))
		assert.NoError(t, err11)
	}()

	data3 := eventSingleData{}
	assert.NoError(t, conn11.ReadJSON(&data3))
	assert.Equal(t, "update", data3.Event)
	assert.Equal(t, "test3 updated", data3.Data["name"])
	assert.NoError(t, conn11.WriteMessage(fhws.CloseMessage, websocket.FormatCloseMessage(1000, "close")))

	// Connect and subscribe to the delete event
	conn12, resp12, err12 := dial("ws://localhost:55555/api/realtime/content?schema=blog&event=delete", nil)
	assert.NoError(t, err12)
	assert.NoError(t, resp12.Body.Close())
	go func() {
		_, err12 = model.Mutation().Where(db.EQ("id", blogID)).Delete(ctx)
		assert.NoError(t, err12)
	}()

	data4 := eventMultipleData{}
	assert.NoError(t, conn12.ReadJSON(&data4))
	assert.Equal(t, "delete", data4.Event)
	assert.Equal(t, "test3 updated", data4.Data[0]["name"])

	// Connect and subscribe to the create event with non-existing records based on filter
	conn13, resp13, err13 := dial(`ws://localhost:55555/api/realtime/content?schema=blog&event=create&filter={"name":"invalid"}`, nil)
	assert.NoError(t, err13)
	assert.NoError(t, resp13.Body.Close())
	go func() {
		_, err13 = model.Create(ctx, schema.NewEntity().Set("name", "test4"))
		assert.NoError(t, err13)
	}()
	time.Sleep(10 * time.Millisecond)
	assert.NoError(t, conn13.WriteMessage(fhws.CloseMessage, websocket.FormatCloseMessage(1000, "close")))

	// Connect and subscribe to the update event with non-existing records based on filter
	conn14, resp14, err14 := dial(`ws://localhost:55555/api/realtime/content?schema=blog&event=update&filter={"name":"invalid"}`, nil)
	assert.NoError(t, err14)
	assert.NoError(t, resp14.Body.Close())
	go func() {
		_, err14 = model.Create(ctx, schema.NewEntity().Set("name", "test5"))
		assert.NoError(t, err14)

		_, err14 = model.Mutation().Where(db.EQ("name", "test5")).Update(ctx, schema.NewEntity().Set("name", "test5 updated"))
		assert.NoError(t, err14)
	}()
	time.Sleep(10 * time.Millisecond)
	assert.NoError(t, conn14.WriteMessage(fhws.CloseMessage, websocket.FormatCloseMessage(1000, "close")))

	// Close with abnormal close message
	conn15, resp15, err15 := dial("ws://localhost:55555/api/realtime/content?schema=blog", nil)
	assert.NoError(t, err15)
	assert.NoError(t, resp15.Body.Close())
	assert.NoError(t, conn15.WriteMessage(fhws.CloseMessage, websocket.FormatCloseMessage(1006, "close")))

	// Close normally
	conn16, resp16, err16 := dial("ws://localhost:55555/api/realtime/content?schema=blog", nil)
	assert.NoError(t, err16)
	assert.NoError(t, resp16.Body.Close())
	assert.NoError(t, conn16.WriteMessage(fhws.CloseMessage, websocket.FormatCloseMessage(1000, "close")))
}
