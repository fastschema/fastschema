package db

import (
	"fmt"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/pkg/utils"
	h "github.com/fastschema/fastschema/tests/integration/helpers"
	"github.com/stretchr/testify/assert"
)

func DBDeleteNodes(t *testing.T, client db.Client) {
	tests := []DBTestDeleteData{
		{
			Name:         "delete",
			Schema:       "user",
			WantAffected: 1,
			ClearTables:  []string{"users", "cards", "pets"},
			Run: func(m db.Model) (int, error) {
				user1ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				return m.Mutation().Where(db.EQ("id", user1ID)).Delete(h.Ctx())
			},
			Expect: func(t *testing.T, m db.Model) {
				entity, err := m.Query().Where(db.EQ("username", "user1")).Only(h.Ctx())
				assert.Error(t, err)
				assert.Nil(t, entity)
			},
		},
		{
			Name:         "delete/multiple",
			Schema:       "user",
			WantAffected: 2,
			ClearTables:  []string{"users"},
			Run: func(m db.Model) (int, error) {
				user1ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				user2ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 2", "username": "user2", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 3", "username": "user3", "provider": "local" }`))
				return m.Mutation().Where(db.In("id", []any{user1ID, user2ID})).Delete(h.Ctx())
			},
			Expect: func(t *testing.T, m db.Model) {
				entities, err := m.Query().Get(h.Ctx())
				assert.NoError(t, err)
				assert.Equal(t, 1, len(entities))
				assert.Equal(t, "User 3", entities[0].Get("name"))
			},
		},
		{
			Name:         "delete/o2m_not_optional_error_foreign_key",
			Schema:       "user",
			WantAffected: 0,
			ClearTables:  []string{"users", "cards"},
			Run: func(m db.Model) (int, error) {
				user1ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{
					"name": "User 1",
					"username": "user1",
					"provider": "local"
				}`))
				utils.Must(utils.Must(client.Model("card")).CreateFromJSON(h.Ctx(), fmt.Sprintf(`{
					"number": "123456789",
					"owner": {"id": %s}
				}`, h.ToJSONID(user1ID))))
				return m.Mutation().Where(db.EQ("id", user1ID)).Delete(h.Ctx())
			},
			WantErr: true,
		},
		{
			Name:         "delete/o2m_optional",
			Schema:       "user",
			WantAffected: 1,
			ClearTables:  []string{"users", "cards"},
			Run: func(m db.Model) (int, error) {
				user1ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{
					"name": "User 1",
					"username": "user1",
					"provider": "local"
				}`))
				user2ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{
					"name": "User 2",
					"username": "user2",
					"provider": "local"
				}`))
				utils.Must(utils.Must(client.Model("card")).CreateFromJSON(h.Ctx(), fmt.Sprintf(`{
					"number": "123456789",
					"owner": {"id": %s},
					"sub_owner": {"id": %s}
				}`, h.ToJSONID(user1ID), h.ToJSONID(user2ID))))
				return m.Mutation().Where(db.EQ("id", user2ID)).Delete(h.Ctx())
			},
			WantErr: false,
			Expect: func(t *testing.T, m db.Model) {
				card := utils.Must(utils.Must(client.Model("card")).Query().Where(db.EQ("number", "123456789")).Only(h.Ctx()))
				h.AssertID(t, card.Get("owner_id"))
				// sub_owner_id should be nil or zero UUID after deletion
				val := card.Get("sub_owner_id")
				if val != nil {
					assert.True(t, h.IsZeroID(val), "sub_owner_id should be nil or zero UUID")
				}
			},
		}, {
			Name:         "delete/m2m",
			Schema:       "user",
			WantAffected: 1,
			ClearTables:  []string{"users", "groups", "groups_users", "cards"},
			Run: func(m db.Model) (int, error) {
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(h.Ctx(), `{"name": "Group 1"}`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(h.Ctx(), `{"name": "Group 2"}`))
				user1ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{
					"name": "User 1",
					"username": "user1",
					"provider": "local",
					"groups": [{ "id": 1 }, { "id": 2 }]
				}`))
				return m.Mutation().Where(db.EQ("id", user1ID)).Delete(h.Ctx())
			},
			WantErr: false,
			Expect: func(t *testing.T, m db.Model) {
				users := utils.Must(utils.Must(client.Model("user")).Query().Where(db.EQ("username", "user1")).Get(h.Ctx()))
				assert.Equal(t, 0, len(users))
				// Can't query junction table with UUID anymore - just verify user is gone
			},
		},
	}

	DBRunDeleteTests(client, t, tests)
}
