package realtimeservice

import (
	"fmt"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
)

type AppLike interface {
	DB() db.Client
	Logger() logger.Logger
}

type WSSerializer interface {
	Serialize(data any) ([]byte, error)
}

type WSClientSerializers = fs.SyncMap[fs.WSClient, WSSerializer]

type RealtimeService struct {
	topics *fs.SyncMap[string, *WSClientSerializers]
	DB     func() db.Client
	Logger func() logger.Logger
}

func New(app AppLike) *RealtimeService {
	return &RealtimeService{
		topics: &fs.SyncMap[string, *WSClientSerializers]{},
		DB:     app.DB,
		Logger: app.Logger,
	}
}

func (rs *RealtimeService) CreateResource(api *fs.Resource) {
	api.
		Group("realtime").
		Add(fs.NewResource("content", rs.Content, &fs.Meta{WS: "/content"}))
}

func (rs *RealtimeService) Topics() *fs.SyncMap[string, *WSClientSerializers] {
	return rs.topics
}

func (rs *RealtimeService) AddClient(client fs.WSClient, topic string, serializer WSSerializer) {
	clientTopics, _ := rs.topics.LoadOrStore(topic, &WSClientSerializers{})
	clientTopics.Store(client, serializer)
	rs.topics.Store(topic, clientTopics)
}

func (rs *RealtimeService) RemoveClient(client fs.WSClient, callCloses ...bool) error {
	for _, topic := range rs.topics.Keys() {
		if clientSerializers, ok := rs.topics.Load(topic); ok {
			clientSerializers.Delete(client)

			if clientSerializers.Len() == 0 {
				rs.topics.Delete(topic)
			}
		}
	}

	if len(callCloses) > 0 && callCloses[0] {
		return client.Close()
	}

	return nil
}

func (rs *RealtimeService) Broadcast(topicNames []string, data any) {
	for _, topicName := range topicNames {
		clientsSerializers, ok := rs.topics.Load(topicName)
		if !ok {
			continue
		}

		for _, client := range clientsSerializers.Keys() {
			serializer, ok := clientsSerializers.Load(client)
			if !ok {
				rs.Logger().Errorf("failed to load serializer for client: %v", client)
			}

			if serializer == nil {
				continue
			}

			go func(client fs.WSClient, serializer WSSerializer) {
				msg, err := serializer.Serialize(data)
				if err != nil {
					writeErr := client.Write([]byte(fmt.Sprintf("failed to serialize message: %v", err)), fs.WSMessageText)
					rs.Logger().Errorf("failed to serialize message: %v, write error: %v", err, writeErr)
					return
				}

				if msg == nil {
					return
				}

				if err := client.Write(msg, fs.WSMessageText); err != nil {
					closeErr := rs.RemoveClient(client, true)
					rs.Logger().Errorf("failed to write message to client: %v, close error: %v", client, closeErr)
				}
			}(client, serializer)
		}
	}
}
