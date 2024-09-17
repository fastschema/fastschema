package realtimeservice

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

func (rs *RealtimeService) Content(c fs.Context, _ any) (any, error) {
	client := c.WSClient()
	serializer, err := rs.createContentSerializer(c)
	if err != nil {
		return nil, client.Close(err.Error())
	}

	rs.AddClient(client, serializer.name, serializer)

	defer func() {
		if err := rs.RemoveClient(client); err != nil {
			c.Logger().Errorf("failed to remove client: %v, err: %v", client, err)
		}
	}()

	for {
		_, msg, err := client.Read()
		if err != nil {
			if client.IsCloseNormal(err) {
				break
			}

			c.Logger().Errorf("failed to read message: %v", err)
			break
		}

		if err = client.Write(msg, fs.WSMessageText); err != nil {
			c.Logger().Errorf("failed to write message: %v, err: %v", msg, err)
			break
		}
	}

	return nil, nil
}

type RealtimeCreateData struct {
	Schema *schema.Schema
	ID     uint64
	Data   *schema.Entity
}

type RealtimeUpdateData struct {
	Schema           *schema.Schema
	Predicates       []*db.Predicate
	UpdateData       *schema.Entity
	OriginalEntities []*schema.Entity
	Affected         int
}

type RealtimeDeleteData struct {
	Schema           *schema.Schema
	Predicates       []*db.Predicate
	OriginalEntities []*schema.Entity
	Affected         int
}

func (rs *RealtimeService) ContentCreateHook(
	ctx context.Context,
	schema *schema.Schema,
	dataCreate *schema.Entity,
	id uint64,
) error {
	go rs.Broadcast([]string{
		fmt.Sprintf("content.%s", schema.Name),
		fmt.Sprintf("content.%s.create", schema.Name),
		fmt.Sprintf("content.%s.create.%d", schema.Name, id),
	}, &RealtimeCreateData{
		Schema: schema,
		ID:     id,
		Data:   dataCreate,
	})

	return nil
}

func (rs *RealtimeService) ContentUpdateHook(
	ctx context.Context,
	schema *schema.Schema,
	predicates []*db.Predicate,
	updateData *schema.Entity,
	originalEntities []*schema.Entity,
	affected int,
) error {
	topics := []string{
		fmt.Sprintf("content.%s", schema.Name),
		fmt.Sprintf("content.%s.update", schema.Name),
	}

	if len(originalEntities) == 0 {
		return nil
	}

	for _, entity := range originalEntities {
		topics = append(
			topics,
			fmt.Sprintf("content.%s.update.%d", schema.Name, entity.ID()),
		)
	}

	go rs.Broadcast(topics, &RealtimeUpdateData{
		Schema:           schema,
		Predicates:       predicates,
		UpdateData:       updateData,
		OriginalEntities: originalEntities,
		Affected:         affected,
	})

	return nil
}

func (rs *RealtimeService) ContentDeleteHook(
	ctx context.Context,
	schema *schema.Schema,
	predicates []*db.Predicate,
	originalEntities []*schema.Entity,
	affected int,
) error {
	topics := []string{
		fmt.Sprintf("content.%s", schema.Name),
		fmt.Sprintf("content.%s.delete", schema.Name),
	}

	if len(originalEntities) == 0 {
		return nil
	}

	for _, entity := range originalEntities {
		topics = append(
			topics,
			fmt.Sprintf("content.%s.delete.%d", schema.Name, entity.ID()),
		)
	}

	go rs.Broadcast(topics, &RealtimeDeleteData{
		Schema:           schema,
		Predicates:       predicates,
		OriginalEntities: originalEntities,
		Affected:         affected,
	})

	return nil
}

type WSContentSerializer struct {
	db         func() db.Client
	schema     *schema.Schema
	event      WSContentEvent
	id         uint64
	name       string
	fields     string
	predicates []*db.Predicate
}

type WSContentSerializeData struct {
	Event WSContentEvent `json:"event"`
	Data  any            `json:"data"`
}

func (tc *WSContentSerializer) Serialize(data any) (msg []byte, err error) {
	query := db.Builder[*schema.Entity](tc.db(), tc.schema.Name)
	if len(tc.fields) > 0 {
		query.Select(strings.Split(tc.fields, ",")...)
	}

	realtimeCreate, ok := data.(*RealtimeCreateData)
	if ok {
		query = query.Where(db.EQ("id", realtimeCreate.ID)).Where(tc.predicates...)
		content, err := query.First(context.Background())
		if err != nil && !db.IsNotFound(err) {
			return nil, fmt.Errorf("realtime.content: %w", err)
		}

		if content == nil {
			return nil, nil
		}

		return json.Marshal(WSContentSerializeData{
			Event: WSContentEventCreate,
			Data:  content,
		})
	}

	realtimeUpdate, ok := data.(*RealtimeUpdateData)
	if ok {
		if len(realtimeUpdate.OriginalEntities) == 0 {
			return nil, nil
		}

		updateIDs := utils.Map(realtimeUpdate.OriginalEntities, func(e *schema.Entity) any {
			return e.ID()
		})

		query = query.Where(db.In("id", updateIDs)).Where(tc.predicates...)
		contents, err := query.Get(context.Background())
		if err != nil && !db.IsNotFound(err) {
			return nil, fmt.Errorf("realtime.content: %w", err)
		}

		if len(contents) == 0 {
			return nil, nil
		}

		sd := WSContentSerializeData{
			Event: WSContentEventUpdate,
			Data:  contents,
		}

		if tc.id > 0 {
			sd.Data = contents[0]
		}

		return json.Marshal(sd)
	}

	realtimeDelete, ok := data.(*RealtimeDeleteData)
	if ok {
		if len(realtimeDelete.OriginalEntities) == 0 {
			return nil, nil
		}

		sd := WSContentSerializeData{
			Event: WSContentEventDelete,
			Data:  realtimeDelete.OriginalEntities,
		}

		if tc.id > 0 {
			sd.Data = realtimeDelete.OriginalEntities[0]
		}

		return json.Marshal(sd)
	}

	return nil, nil
}

func (rs *RealtimeService) createContentSerializer(c fs.Context) (*WSContentSerializer, error) {
	schemaName := c.Arg("schema")
	event := c.Arg("event", "*")
	id := uint64(c.ArgInt("id"))
	fields := c.Arg("select")
	filter := c.Arg("filter")
	predicates := make([]*db.Predicate, 0)
	contentEvent := WSContentEventFromString(event)
	topicParts := []string{"content", schemaName}

	if !contentEvent.Valid() {
		return nil, fmt.Errorf("realtime.content: invalid event '%v'", event)
	}

	if event != WSContentEventAll.String() {
		topicParts = append(topicParts, event)
	}

	if id > 0 {
		topicParts = append(topicParts, fmt.Sprintf("%d", id))
	}

	schema, err := rs.DB().SchemaBuilder().Schema(schemaName)
	if err != nil {
		return nil, fmt.Errorf("realtime.content: schema not found: %v", schemaName)
	}

	if filter != "" {
		predicates, err = db.CreatePredicatesFromFilterObject(rs.DB().SchemaBuilder(), schema, filter)
		if err != nil {
			return nil, fmt.Errorf("realtime.content: invalid filter: %v", err)
		}
	}

	return &WSContentSerializer{
		db:         rs.DB,
		schema:     schema,
		event:      contentEvent,
		id:         id,
		name:       strings.Join(topicParts, "."),
		fields:     fields,
		predicates: predicates,
	}, nil
}
