package db

import (
	"testing"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/testutils"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func DBDeleteNodes(t *testing.T, client app.DBClient) {
	tests := []testutils.DBTestDeleteData{
		{
			Name:         "delete",
			Schema:       "user",
			Predicates:   []*app.Predicate{app.EQ("id", 1)},
			WantAffected: 1,
			ClearTables:  []string{"users", "cards", "pets"},
			Prepare: func(t *testing.T, m app.Model) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(`{ "name": "User 1", "username": "user1" }`))
			},
			Expect: func(t *testing.T, m app.Model) {
				entity, err := m.Query().Where(app.EQ("id", 1)).Only()
				assert.Error(t, err)
				assert.Nil(t, entity)
			},
		},
		{
			Name:         "delete/multiple",
			Schema:       "user",
			WantAffected: 2,
			Predicates:   []*app.Predicate{app.LT("id", 3)},
			ClearTables:  []string{"users"},
			Prepare: func(t *testing.T, m app.Model) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(`{ "name": "User 1", "username": "user1" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(`{ "name": "User 2", "username": "user2" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(`{ "name": "User 3", "username": "user3" }`))
			},
			Expect: func(t *testing.T, m app.Model) {
				entities, err := m.Query().Get()
				assert.NoError(t, err)
				assert.Equal(t, 1, len(entities))
				assert.Equal(t, "User 3", entities[0].Get("name"))
			},
		},
		{
			Name:         "delete/o2m_not_optional_error_foreign_key",
			Schema:       "user",
			WantAffected: 0,
			Predicates:   []*app.Predicate{app.EQ("id", 1)},
			ClearTables:  []string{"users", "cards"},
			Prepare: func(t *testing.T, m app.Model) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(`{
					"name": "User 1",
					"username": "user1"
				}`))
				utils.Must(utils.Must(client.Model("card")).CreateFromJSON(`{
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
			Predicates:   []*app.Predicate{app.EQ("id", 2)},
			ClearTables:  []string{"users", "cards"},
			Prepare: func(t *testing.T, m app.Model) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(`{
					"name": "User 1",
					"username": "user1"
				}`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(`{
					"name": "User 2",
					"username": "user2"
				}`))
				utils.Must(utils.Must(client.Model("card")).CreateFromJSON(`{
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
			Expect: func(t *testing.T, m app.Model) {
				card := utils.Must(utils.Must(client.Model("card")).Query().Where(app.EQ("id", 1)).Only())
				assert.Equal(t, uint64(1), card.Get("owner_id"))
				assert.Equal(t, nil, card.Get("sub_owner_id"))
			},
		}, {
			Name:         "delete/m2m",
			Schema:       "user",
			WantAffected: 1,
			Predicates:   []*app.Predicate{app.EQ("id", 1)},
			ClearTables:  []string{"users", "groups", "groups_users", "cards"},
			Prepare: func(t *testing.T, m app.Model) {
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(`{
					"name": "Group 1"
				}`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(`{
					"name": "Group 2"
				}`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(`{
					"name": "User 1",
					"username": "user1",
					"groups": [
						{ "id": 1 },
						{ "id": 2 }
					]
				}`))
			},
			WantErr: false,
			Expect: func(t *testing.T, m app.Model) {
				users := utils.Must(utils.Must(client.Model("user")).Query().Where(app.EQ("id", 1)).Get())
				assert.Equal(t, 0, len(users))
				groupsUsers := utils.Must(utils.Must(client.Model("groups_users")).Query().Where(app.EQ("users", 1)).Get())
				assert.Equal(t, 0, len(groupsUsers))
			},
		},
	}

	testutils.DBRunDeleteTests(client, t, tests)
}
