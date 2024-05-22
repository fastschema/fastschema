package db

import (
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func DBDeleteNodes(t *testing.T, client db.Client) {
	tests := []DBTestDeleteData{
		{
			Name:         "delete",
			Schema:       "user",
			Predicates:   []*db.Predicate{db.EQ("id", 1)},
			WantAffected: 1,
			ClearTables:  []string{"users", "cards", "pets"},
			Prepare: func(t *testing.T, m db.Model) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 1", "username": "user1" }`))
			},
			Expect: func(t *testing.T, m db.Model) {
				entity, err := m.Query().Where(db.EQ("id", 1)).Only(Ctx())
				assert.Error(t, err)
				assert.Nil(t, entity)
			},
		},
		{
			Name:         "delete/multiple",
			Schema:       "user",
			WantAffected: 2,
			Predicates:   []*db.Predicate{db.LT("id", 3)},
			ClearTables:  []string{"users"},
			Prepare: func(t *testing.T, m db.Model) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 1", "username": "user1" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 2", "username": "user2" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 3", "username": "user3" }`))
			},
			Expect: func(t *testing.T, m db.Model) {
				entities, err := m.Query().Get(Ctx())
				assert.NoError(t, err)
				assert.Equal(t, 1, len(entities))
				assert.Equal(t, "User 3", entities[0].Get("name"))
			},
		},
		{
			Name:         "delete/o2m_not_optional_error_foreign_key",
			Schema:       "user",
			WantAffected: 0,
			Predicates:   []*db.Predicate{db.EQ("id", 1)},
			ClearTables:  []string{"users", "cards"},
			Prepare: func(t *testing.T, m db.Model) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{
					"name": "User 1",
					"username": "user1"
				}`))
				utils.Must(utils.Must(client.Model("card")).CreateFromJSON(Ctx(), `{
					"number": "123456789",
					"owner": {
						"id": 1
					}
				}`))
			},
			WantErr: true,
		},
		{
			Name:         "delete/o2m_optional",
			Schema:       "user",
			WantAffected: 1,
			Predicates:   []*db.Predicate{db.EQ("id", 2)},
			ClearTables:  []string{"users", "cards"},
			Prepare: func(t *testing.T, m db.Model) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{
					"name": "User 1",
					"username": "user1"
				}`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{
					"name": "User 2",
					"username": "user2"
				}`))
				utils.Must(utils.Must(client.Model("card")).CreateFromJSON(Ctx(), `{
					"number": "123456789",
					"owner": {
						"id": 1
					},
					"sub_owner": {
						"id": 2
					}
				}`))
			},
			WantErr: false,
			Expect: func(t *testing.T, m db.Model) {
				card := utils.Must(utils.Must(client.Model("card")).Query().Where(db.EQ("id", 1)).Only(Ctx()))
				assert.Equal(t, uint64(1), card.Get("owner_id"))
				assert.Equal(t, nil, card.Get("sub_owner_id"))
			},
		}, {
			Name:         "delete/m2m",
			Schema:       "user",
			WantAffected: 1,
			Predicates:   []*db.Predicate{db.EQ("id", 1)},
			ClearTables:  []string{"users", "groups", "groups_users", "cards"},
			Prepare: func(t *testing.T, m db.Model) {
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(Ctx(), `{
					"name": "Group 1"
				}`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(Ctx(), `{
					"name": "Group 2"
				}`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{
					"name": "User 1",
					"username": "user1",
					"groups": [
						{ "id": 1 },
						{ "id": 2 }
					]
				}`))
			},
			WantErr: false,
			Expect: func(t *testing.T, m db.Model) {
				users := utils.Must(utils.Must(client.Model("user")).Query().Where(db.EQ("id", 1)).Get(Ctx()))
				assert.Equal(t, 0, len(users))
				groupsUsers := utils.Must(utils.Must(client.Model("groups_users")).Query().Where(db.EQ("users", 1)).Get(Ctx()))
				assert.Equal(t, 0, len(groupsUsers))
			},
		},
	}

	DBRunDeleteTests(client, t, tests)
}
