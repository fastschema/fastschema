package db

import (
	"fmt"
	"testing"

	"entgo.io/ent/dialect"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/pkg/utils"
	h "github.com/fastschema/fastschema/tests/integration/helpers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func DBUpdateNodes(t *testing.T, client db.Client) {
	tests := []DBTestUpdateData{
		{
			Name:         "fields",
			Schema:       "user",
			InputJSON:    `{ "age": 20 }`,
			WantAffected: 2,
			ClearTables:  []string{"users"},
			Prepare: func(t *testing.T, m db.Model) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 2", "username": "user2", "provider": "local" }`))
			},
			Expect: func(t *testing.T, m db.Model) {
				entities, err := m.Query().Order("username").Get(h.Ctx())
				assert.NoError(t, err)
				assert.NotNil(t, entities)
				h.AssertID(t, entities[0].ID())
				assert.Equal(t, uint(20), entities[0].Get("age"))
				h.AssertID(t, entities[1].ID())
				assert.Equal(t, uint(20), entities[1].Get("age"))
			},
		},
		{
			Name:         "fields/clear/nested_clear_block_o2m",
			Schema:       "user",
			InputJSON:    `{}`,
			WantAffected: 1,
			ClearTables:  []string{"users", "pets"},
			Run: func(m db.Model, _ *entity.Entity) (int, error) {
				user1ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "Pet 1", "owner": {"id": %s}}`, h.ToJSONID(user1ID))))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "Pet 2", "owner": {"id": %s}}`, h.ToJSONID(user1ID))))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{
					"name": "User 2",
					"username": "user2",
					"provider": "local",
					"sub_pets": [ { "id": 1 }, { "id": 2 } ]
				}`))
				inputJSON := `{
					"sub_pets": { "$clear": [ { "id": 1 }, { "id": 2 } ] }
				}`
				e := utils.Must(entity.NewEntityFromJSON(inputJSON))
				return m.Mutation().Where(db.EQ("username", "user2")).Update(h.Ctx(), e)
			},
			Expect: func(t *testing.T, m db.Model) {
				pets := utils.Must(utils.Must(client.Model("pet")).Query().Order("id").Get(h.Ctx()))
				require.Len(t, pets, 2)
				for _, p := range pets {
					// sub_owner_id can be nil or a zero UUID when cleared
					val := p.Get("sub_owner_id")
					if val != nil {
						assert.Equal(t, uuid.UUID{}, val, "sub_owner_id should be nil or zero UUID")
					}
				}
			},
		},
		{
			Name:         "fields/add/nested_add_block_o2m",
			Schema:       "user",
			InputJSON:    `{}`,
			WantAffected: 1,
			ClearTables:  []string{"users", "pets"},
			Run: func(m db.Model, _ *entity.Entity) (int, error) {
				user1ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "Pet 1", "owner": {"id": %s}}`, h.ToJSONID(user1ID))))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "Pet 2", "owner": {"id": %s}}`, h.ToJSONID(user1ID))))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 2", "username": "user2", "provider": "local" }`))
				inputJSON := `{
					"sub_pets": { "$add": [ { "id": 1 }, { "id": 2 } ] }
				}`
				e := utils.Must(entity.NewEntityFromJSON(inputJSON))
				return m.Mutation().Where(db.EQ("username", "user2")).Update(h.Ctx(), e)
			},
			Expect: func(t *testing.T, m db.Model) {
				pets := utils.Must(utils.Must(client.Model("pet")).Query().Order("id").Get(h.Ctx()))
				require.Len(t, pets, 2)
				for _, p := range pets {
					h.AssertID(t, p.Get("sub_owner_id"))
				}
			},
		},
		{
			Name:         "fields/add_clear/nested_add_clear_block_o2m",
			Schema:       "user",
			InputJSON:    `{}`,
			WantAffected: 1,
			ClearTables:  []string{"users", "pets"},
			Run: func(m db.Model, _ *entity.Entity) (int, error) {
				user1ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				user2ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 2", "username": "user2", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "Pet 1", "owner": {"id": %s}}`, h.ToJSONID(user2ID))))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "Pet 2", "owner": {"id": %s}}`, h.ToJSONID(user1ID))))
				inputJSON := `{
					"sub_pets": { "$clear": true, "$add": [ { "id": 2 } ] }
				}`
				e := utils.Must(entity.NewEntityFromJSON(inputJSON))
				return m.Mutation().Where(db.EQ("username", "user2")).Update(h.Ctx(), e)
			},
			Expect: func(t *testing.T, m db.Model) {
				pets := utils.Must(utils.Must(client.Model("pet")).Query().Order("id").Get(h.Ctx()))
				require.Len(t, pets, 2)
				// sub_owner_id can be nil or a zero UUID when cleared
				val := pets[0].Get("sub_owner_id")
				if val != nil {
					assert.Equal(t, uuid.UUID{}, val, "sub_owner_id should be nil or zero UUID")
				}
				h.AssertID(t, pets[1].Get("sub_owner_id"))
			},
		},
		{
			Name:         "predicates",
			Schema:       "user",
			InputJSON:    `{}`,
			WantAffected: 2,
			ClearTables:  []string{"users"},
			Run: func(m db.Model, _ *entity.Entity) (int, error) {
				user1ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 2", "username": "user2", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 3", "username": "user3", "provider": "local" }`))
				e := utils.Must(entity.NewEntityFromJSON(`{ "age": 20 }`))
				return m.Mutation().Where(db.NEQ("id", user1ID)).Update(h.Ctx(), e)
			},
			Expect: func(t *testing.T, m db.Model) {
				entities, err := m.Query(db.NEQ("username", "user1")).Order("username").Get(h.Ctx())
				assert.NoError(t, err)
				assert.NotNil(t, entities)
				assert.Len(t, entities, 2)
				h.AssertID(t, entities[0].ID())
				assert.Equal(t, uint(20), entities[0].Get("age"))
				h.AssertID(t, entities[1].ID())
				assert.Equal(t, uint(20), entities[1].Get("age"))
			},
		},
		{
			Name:         "fields/set_modifier/expr",
			Schema:       "user",
			InputJSON:    `{}`,
			ClearTables:  []string{"users"},
			WantAffected: 1,
			Run: func(m db.Model, _ *entity.Entity) (int, error) {
				user1ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local", "bio": "My BIO" }`))
				bioColumn := utils.If(client.Dialect() == dialect.Postgres, "bio", "`bio`")
				inputJSON := fmt.Sprintf(`{
					"name": "User 1 name",
					"username": "user1",
					"provider": "local",
					"$expr": {
						"bio": "LOWER(%s)"
					}
				}`, bioColumn)
				e := utils.Must(entity.NewEntityFromJSON(inputJSON))
				return m.Mutation().Where(db.EQ("id", user1ID)).Update(h.Ctx(), e)
			},
			Expect: func(t *testing.T, m db.Model) {
				entity := utils.Must(m.Query(db.EQ("username", "user1")).Only(h.Ctx()))
				assert.NotNil(t, entity)
				assert.Equal(t, "my bio", entity.Get("bio"))
			},
		},
		{
			Name:         "fields/add",
			Schema:       "user",
			InputJSON:    `{}`,
			ClearTables:  []string{"users"},
			WantAffected: 1,
			Run: func(m db.Model, _ *entity.Entity) (int, error) {
				user1ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local", "age": "5" }`))
				inputJSON := `{
					"$add": {
						"age": 3
					}
				}`
				e := utils.Must(entity.NewEntityFromJSON(inputJSON))
				return m.Mutation().Where(db.EQ("id", user1ID)).Update(h.Ctx(), e)
			},
			Expect: func(t *testing.T, m db.Model) {
				entity := utils.Must(m.Query(db.EQ("username", "user1")).Only(h.Ctx()))
				assert.NotNil(t, entity)
				assert.Equal(t, uint(8), entity.Get("age"))
			},
		},
		{
			Name:         "fields/add_o2m_m2m",
			Schema:       "user",
			InputJSON:    `{}`,
			ClearTables:  []string{"users", "groups", "pets", "sub_groups_sub_users", "groups_users"},
			WantAffected: 1,
			Run: func(m db.Model, _ *entity.Entity) (int, error) {
				user1ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(h.Ctx(), `{ "name": "Group 1" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(h.Ctx(), `{ "name": "Group 2" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(h.Ctx(), `{ "name": "Group 3" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(h.Ctx(), `{ "name": "Group 4" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(h.Ctx(), `{ "name": "Group 5" }`))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "Pet 1", "owner": {"id": %s}}`, h.ToJSONID(user1ID))))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "Pet 2", "owner": {"id": %s}}`, h.ToJSONID(user1ID))))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "Pet 3", "owner": {"id": %s}}`, h.ToJSONID(user1ID))))
				inputJSON := `{
					"$add": {
						"sub_pets": [ { "id": 2 }, { "id": 3 } ],
						"sub_groups": [ { "id": 4 }, { "id": 5 } ]
					}
				}`
				e := utils.Must(entity.NewEntityFromJSON(inputJSON))
				return m.Mutation().Where(db.EQ("id", user1ID)).Update(h.Ctx(), e)
			},
			Expect: func(t *testing.T, m db.Model) {
				pet2 := utils.Must(utils.Must(client.Model("pet")).Query(db.EQ("name", "Pet 2")).Only(h.Ctx()))
				pet3 := utils.Must(utils.Must(client.Model("pet")).Query(db.EQ("name", "Pet 3")).Only(h.Ctx()))
				h.AssertID(t, pet2.Get("sub_owner_id"))
				h.AssertID(t, pet3.Get("sub_owner_id"))

				user1 := utils.Must(m.Query(db.EQ("username", "user1")).Only(h.Ctx()))
				subGroupsUsers := utils.Must(utils.Must(client.Model("sub_groups_sub_users")).Query(db.EQ("sub_users", user1.ID())).Get(h.Ctx()))
				subGroupsIDs := utils.Map(subGroupsUsers, func(e *entity.Entity) uint64 {
					return e.Get("sub_groups").(uint64)
				})
				assert.Equal(t, []uint64{4, 5}, subGroupsIDs)
			},
		},
		{
			Name:         "fields/clear",
			Schema:       "user",
			InputJSON:    `{}`,
			ClearTables:  []string{"users"},
			WantAffected: 1,
			Run: func(m db.Model, _ *entity.Entity) (int, error) {
				user1ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local", "bio": "My BIO" }`))
				inputJSON := `{
					"deleted": true,
					"$clear": {
						"bio": true
					}
				}`
				e := utils.Must(entity.NewEntityFromJSON(inputJSON))
				return m.Mutation().Where(db.EQ("id", user1ID)).Update(h.Ctx(), e)
			},
			Expect: func(t *testing.T, m db.Model) {
				user := utils.Must(m.Query(db.EQ("username", "user1")).Only(h.Ctx()))
				assert.NotNil(t, user)
				assert.Equal(t, true, user.Get("deleted"))
				assert.Equal(t, nil, user.Get("bio"))
			},
		},
		{
			Name:         "fields/clear/o2o_o2m_m2m_all",
			Schema:       "user",
			InputJSON:    `{}`,
			WantAffected: 1,
			ClearTables:  []string{"users", "groups", "pets", "cars"},
			Run: func(m db.Model, _ *entity.Entity) (int, error) {
				user1ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("car")).CreateFromJSON(h.Ctx(), `{ "name": "Car 1" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(h.Ctx(), `{ "name": "Group 1" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(h.Ctx(), `{ "name": "Group 2" }`))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "Pet 1", "owner": {"id": %s}}`, h.ToJSONID(user1ID))))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "Pet 2", "owner": {"id": %s}}`, h.ToJSONID(user1ID))))
				user2ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{
					"name": "User 2",
					"username": "user2",
					"provider": "local",
					"bio": "My BIO",
					"car": { "id": 1 },
					"sub_pets": [ { "id": 1 }, { "id": 2 } ],
					"sub_groups": [ { "id": 1 }, { "id": 2 } ]
				}`))
				inputJSON := `{
					"$clear": {
						"bio": true,
						"car": true,
						"sub_pets": true,
						"sub_groups": true
					}
				}`
				e := utils.Must(entity.NewEntityFromJSON(inputJSON))
				return m.Mutation().Where(db.EQ("id", user2ID)).Update(h.Ctx(), e)
			},
			Expect: func(t *testing.T, m db.Model) {
				user := utils.Must(m.Query(db.EQ("username", "user2")).Only(h.Ctx()))
				assert.NotNil(t, user)
				assert.Equal(t, nil, user.Get("bio"))
				assert.Equal(t, nil, user.Get("car_id"))

				subGroupsUsers := utils.Must(utils.Must(client.Model("sub_groups_sub_users")).Query(db.EQ("sub_users", user.ID())).Get(h.Ctx()))
				assert.Equal(t, 0, len(subGroupsUsers))

				subPets := utils.Must(utils.Must(client.Model("pet")).Query(db.EQ("sub_owner_id", user.ID())).Get(h.Ctx()))
				assert.Equal(t, 0, len(subPets))
			},
		},
		{
			Name:         "fields/clear/o2o_o2m_m2m_part",
			Schema:       "user",
			InputJSON:    `{}`,
			WantAffected: 1,
			ClearTables:  []string{"users", "groups", "pets", "cars"},
			Run: func(m db.Model, _ *entity.Entity) (int, error) {
				user1ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("car")).CreateFromJSON(h.Ctx(), `{ "name": "Car 1" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(h.Ctx(), `{ "name": "Group 1" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(h.Ctx(), `{ "name": "Group 2" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(h.Ctx(), `{ "name": "Group 3" }`))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "Pet 1", "owner": {"id": %s}}`, h.ToJSONID(user1ID))))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "Pet 2", "owner": {"id": %s}}`, h.ToJSONID(user1ID))))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(h.Ctx(), fmt.Sprintf(`{"name": "Pet 3", "owner": {"id": %s}}`, h.ToJSONID(user1ID))))
				user2ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{
					"name": "User 2",
					"username": "user2",
					"provider": "local",
					"sub_pets": [ { "id": 1 }, { "id": 2 }, { "id": 3 } ],
					"sub_groups": [ { "id": 1 }, { "id": 2 }, { "id": 3 } ]
				}`))
				inputJSON := `{
					"$clear": {
						"bio": true,
						"car": true,
						"sub_pets": [ { "id": 1 }, { "id": 2 } ],
						"sub_groups": [ { "id": 1 }, { "id": 2 }]
					}
				}`
				e := utils.Must(entity.NewEntityFromJSON(inputJSON))
				return m.Mutation().Where(db.EQ("id", user2ID)).Update(h.Ctx(), e)
			},
			Expect: func(t *testing.T, m db.Model) {
				user := utils.Must(m.Query(db.EQ("username", "user2")).Only(h.Ctx()))
				assert.NotNil(t, user)

				subGroupsUsers := utils.Must(utils.Must(client.Model("sub_groups_sub_users")).Query(db.EQ("sub_users", user.ID())).Get(h.Ctx()))
				subGroupsUsersIds := utils.Map(subGroupsUsers, func(e *entity.Entity) uint64 {
					return e.Get("sub_groups").(uint64)
				})
				assert.Equal(t, 1, len(subGroupsUsersIds))
				assert.Equal(t, []uint64{3}, subGroupsUsersIds)

				subPets := utils.Must(utils.Must(client.Model("pet")).Query(db.EQ("sub_owner_id", user.ID())).Get(h.Ctx()))
				subPetsIds := utils.Map(subPets, func(e *entity.Entity) uint64 {
					return e.Get("id").(uint64)
				})
				assert.Equal(t, 1, len(subPetsIds))
				assert.Equal(t, []uint64{3}, subPetsIds)
			},
		},
		{
			Name:         "fields/set/block",
			Schema:       "user",
			InputJSON:    `{}`,
			WantAffected: 1,
			ClearTables:  []string{"users", "pets", "cards"},
			Run: func(m db.Model, _ *entity.Entity) (int, error) {
				user1ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				user2ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 2", "username": "user2", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("card")).CreateFromJSON(h.Ctx(), fmt.Sprintf(`{
					"number": "00001",
					"owner": {
						"id": %s
					}
				}`, h.ToJSONID(user1ID))))
				utils.Must(utils.Must(client.Model("card")).CreateFromJSON(h.Ctx(), fmt.Sprintf(`{
					"number": "00002",
					"owner": {
						"id": %s
					}
				}`, h.ToJSONID(user2ID))))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(h.Ctx(), fmt.Sprintf(`{
					"name": "Pet 1",
					"owner": {
						"id": %s
					}
				}`, h.ToJSONID(user1ID))))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(h.Ctx(), fmt.Sprintf(`{
					"name": "Pet 2",
					"owner": {
						"id": %s
					}
				}`, h.ToJSONID(user1ID))))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(h.Ctx(), `{ "name": "Group 1" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(h.Ctx(), `{ "name": "Group 2" }`))

				user3ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{
					"name": "User 3",
					"username": "user3",
					"provider": "local",
					"sub_card": { "id": 1 },
					"sub_pets": [ { "id": 1 } ],
					"sub_groups": [ { "id": 1 } ]
				}`))
				inputJSON := `{
					"name": "User 3 updated",
					"username": "user3",
					"provider": "local",
					"$set": {
						"bio": "Hello World",
						"sub_card": { "id": 2 },
						"sub_pets": [ { "id": 2 } ],
						"sub_groups": [ { "id": 2 } ]
					}
				}`
				e := utils.Must(entity.NewEntityFromJSON(inputJSON))
				return m.Mutation().Where(db.EQ("id", user3ID)).Update(h.Ctx(), e)
			},
			Expect: func(t *testing.T, m db.Model) {
				user3 := utils.Must(m.Query(db.EQ("username", "user3")).Only(h.Ctx()))
				assert.NotNil(t, user3)

				assert.Equal(t, "User 3 updated", user3.Get("name").(string))
				assert.Equal(t, "Hello World", user3.Get("bio").(string))

				subCards := utils.Must(utils.Must(client.Model("pet")).Query(db.EQ("sub_owner_id", user3.ID())).Get(h.Ctx()))
				subCardsIds := utils.Map(subCards, func(e *entity.Entity) uint64 {
					return e.Get("id").(uint64)
				})
				assert.Equal(t, 1, len(subCardsIds))
				assert.Equal(t, []uint64{2}, subCardsIds)

				subPets := utils.Must(utils.Must(client.Model("pet")).Query(db.EQ("sub_owner_id", user3.ID())).Get(h.Ctx()))
				subPetsIds := utils.Map(subPets, func(e *entity.Entity) uint64 {
					return e.Get("id").(uint64)
				})
				assert.Equal(t, 1, len(subPetsIds))
				assert.Equal(t, []uint64{2}, subPetsIds)

				subGroupsUsers := utils.Must(utils.Must(client.Model("sub_groups_sub_users")).Query(db.EQ("sub_users", user3.ID())).Get(h.Ctx()))
				subGroupsUsersIds := utils.Map(subGroupsUsers, func(e *entity.Entity) uint64 {
					return e.Get("sub_groups").(uint64)
				})
				assert.Equal(t, 1, len(subGroupsUsersIds))
				assert.Equal(t, []uint64{2}, subGroupsUsersIds)
			},
		},
		{
			Name:         "edges/o2o_non_inverse_and_m2o",
			Schema:       "user",
			InputJSON:    `{}`,
			WantAffected: 1,
			ClearTables:  []string{"users", "cars", "workplaces", "rooms", "users"},
			Run: func(m db.Model, _ *entity.Entity) (int, error) {
				user1ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("car")).CreateFromJSON(h.Ctx(), `{ "name": "Car 1" }`))
				utils.Must(utils.Must(client.Model("workplace")).CreateFromJSON(h.Ctx(), `{ "name": "Workplace 1" }`))
				utils.Must(utils.Must(client.Model("room")).CreateFromJSON(h.Ctx(), `{ "name": "Room 1" }`))
				user2ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{
					"name": "User 2",
					"username": "user2",
					"provider": "local",
					"car": {
						"id": 1
					},
					"workplace": {
						"id": 1
					}
				}`))
				inputJSON := fmt.Sprintf(`{
					"$clear": {
						"car": true,
						"workplace": true
					},
					"$add": {
						"room": { "id": 1 },
						"parent": { "id": %s }
					}
				}`, h.ToJSONID(user1ID))
				e := utils.Must(entity.NewEntityFromJSON(inputJSON))
				return m.Mutation().Where(db.EQ("id", user2ID)).Update(h.Ctx(), e)
			},
			Expect: func(t *testing.T, m db.Model) {
				user2 := utils.Must(m.Query(db.EQ("username", "user2")).Only(h.Ctx()))
				assert.NotNil(t, user2)

				assert.Equal(t, nil, user2.Get("car_id"))
				assert.Equal(t, nil, user2.Get("workplace_id"))
				assert.Equal(t, uint64(1), user2.Get("room_id"))
				h.AssertID(t, user2.Get("parent_id"))
			},
		},
		{
			Name:         "edges/o2o_bidi",
			Schema:       "user",
			InputJSON:    `{}`,
			WantAffected: 1,
			ClearTables:  []string{"users"},
			Run: func(m db.Model, _ *entity.Entity) (int, error) {
				user1ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				user2ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 2", "username": "user2", "provider": "local" }`))
				user3ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 3", "username": "user3", "provider": "local" }`))
				user4ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), fmt.Sprintf(`{
					"name": "User 4",
					"username": "user4",
					"provider": "local",
					"partner": {
						"id": %s
					},
					"spouse": {
						"id": %s
					}
				}`, h.ToJSONID(user1ID), h.ToJSONID(user2ID))))
				inputJSON := fmt.Sprintf(`{
					"$clear": {
						"partner": true,
						"spouse": { "id": %s }
					},
					"$add": {
						"spouse": { "id": %s }
					}
				}`, h.ToJSONID(user2ID), h.ToJSONID(user3ID))
				e := utils.Must(entity.NewEntityFromJSON(inputJSON))
				return m.Mutation().Where(db.EQ("id", user4ID)).Update(h.Ctx(), e)
			},
			Expect: func(t *testing.T, m db.Model) {
				user4 := utils.Must(m.Query(db.EQ("username", "user4")).Only(h.Ctx()))
				assert.NotNil(t, user4)

				// partner_id can be nil or a zero UUID when cleared
				val := user4.Get("partner_id")
				if val != nil {
					assert.Equal(t, uuid.UUID{}, val, "partner_id should be nil or zero UUID")
				}
				h.AssertID(t, user4.Get("spouse_id"))
			},
		},
		{
			Name:         "edges/clear_add_m2m",
			Schema:       "user",
			InputJSON:    `{}`,
			WantAffected: 1,
			ClearTables:  []string{"users", "groups", "groups_users", "followers_following", "blockers_blocking", "friends_user"},
			Run: func(m db.Model, _ *entity.Entity) (int, error) {
				utils.Must(utils.Must(client.Model("comment")).CreateFromJSON(h.Ctx(), `{ "content": "Comment 1" }`))
				utils.Must(utils.Must(client.Model("comment")).CreateFromJSON(h.Ctx(), `{ "content": "Comment 2" }`))

				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(h.Ctx(), `{ "name": "Group 1" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(h.Ctx(), `{ "name": "Group 2" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(h.Ctx(), `{ "name": "Group 3" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(h.Ctx(), `{ "name": "Group 4" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(h.Ctx(), `{ "name": "Group 5" }`))

				user1ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				user2ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 2", "username": "user2", "provider": "local" }`))
				user3ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 3", "username": "user3", "provider": "local" }`))
				user4ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 4", "username": "user4", "provider": "local" }`))
				user5ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 5", "username": "user5", "provider": "local" }`))
				user6ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 6", "username": "user6", "provider": "local" }`))
				user7ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 7", "username": "user7", "provider": "local" }`))
				user8ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 8", "username": "user8", "provider": "local" }`))

				user9ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), fmt.Sprintf(`{
					"name": "User 9",
					"username": "user9",
					"provider": "local",
					"blocking": [{ "id": %s }, { "id": %s }],
					"following": [{ "id": %s }, { "id": %s }],
					"friends": [{ "id": %s }, { "id": %s }],
					"groups": [ { "id": 1 }, { "id": 2 }, { "id": 3 } ],
					"comments": [{ "id": 1 }, { "id": 2 }]
				}`, h.ToJSONID(user1ID), h.ToJSONID(user2ID), h.ToJSONID(user3ID), h.ToJSONID(user4ID), h.ToJSONID(user5ID), h.ToJSONID(user6ID))))

				inputJSON := fmt.Sprintf(`{
					"$clear": {
						"blocking": true,
						"following": [{ "id": %s }],
						"friends": { "id": %s },
						"groups": [ { "id": 1 }, { "id": 2 } ],
						"comments": true
					},
					"$add": {
						"friends": [ { "id": %s }, { "id": %s } ],
						"groups": [ { "id": 4 }, { "id": 5 } ]
					}
				}`, h.ToJSONID(user3ID), h.ToJSONID(user5ID), h.ToJSONID(user7ID), h.ToJSONID(user8ID))
				e := utils.Must(entity.NewEntityFromJSON(inputJSON))
				return m.Mutation().Where(db.EQ("id", user9ID)).Update(h.Ctx(), e)
			},
			Expect: func(t *testing.T, m db.Model) {
				user9 := utils.Must(m.Query(db.EQ("username", "user9")).Only(h.Ctx()))
				assert.NotNil(t, user9)

				blockingUsers := utils.Must(utils.Must(client.Model("blockers_blocking")).Query(db.EQ("blockers", user9.ID())).Get(h.Ctx()))
				assert.Equal(t, 0, len(blockingUsers))

				followingUsers := utils.Must(utils.Must(client.Model("followers_following")).Query(db.EQ("followers", user9.ID())).Get(h.Ctx()))
				assert.Equal(t, 1, len(followingUsers))
				h.AssertID(t, followingUsers[0].Get("following"))

				friends := utils.Must(utils.Must(client.Model("friends_user")).Query(db.EQ("user", user9.ID())).Get(h.Ctx()))
				assert.Equal(t, 3, len(friends))

				subGroupsUsers := utils.Must(utils.Must(client.Model("groups_users")).Query(db.EQ("users", user9.ID())).Get(h.Ctx()))
				subGroupsUsersIds := utils.Map(subGroupsUsers, func(e *entity.Entity) uint64 {
					return e.Get("groups").(uint64)
				})
				assert.Equal(t, []uint64{3, 4, 5}, subGroupsUsersIds)
			},
		},
		{
			Name:         "fields/add_set_clear",
			Schema:       "user",
			InputJSON:    `{}`,
			ClearTables:  []string{"users"},
			WantAffected: 1,
			Run: func(m db.Model, _ *entity.Entity) (int, error) {
				user1ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local", "age": 10, "bio": "Bio 1" }`))
				inputJSON := `{
					"name": "User 1 updated",
					"username": "user1",
					"provider": "local",
					"deleted": true,
					"$add": {
						"age": 1
					},
					"$clear": {
						"bio": true
					}
				}`
				e := utils.Must(entity.NewEntityFromJSON(inputJSON))
				return m.Mutation().Where(db.EQ("id", user1ID)).Update(h.Ctx(), e)
			},
			Expect: func(t *testing.T, m db.Model) {
				user1 := utils.Must(m.Query(db.EQ("username", "user1")).Only(h.Ctx()))
				assert.NotNil(t, user1)

				assert.Equal(t, "User 1 updated", user1.Get("name"))
				assert.Equal(t, true, user1.Get("deleted"))
				assert.Equal(t, uint(11), user1.Get("age"))
				assert.Equal(t, nil, user1.Get("bio"))
			},
		},
		{
			Name:         "fields/ensure_exists",
			Schema:       "user",
			InputJSON:    `{}`,
			WantAffected: 1,
			ClearTables:  []string{"users"},
			Run: func(m db.Model, _ *entity.Entity) (int, error) {
				user1ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(h.Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local", "age": 10, "bio": "Bio 1" }`))
				inputJSON := `{
					"$add": {
						"age": 1
					},
					"$clear": {
						"bio": true
					},
					"deleted": true
				}`
				e := utils.Must(entity.NewEntityFromJSON(inputJSON))
				return m.Mutation().Where(db.EQ("id", user1ID), db.IsFalse("deleted")).Update(h.Ctx(), e)
			},
			Expect: func(t *testing.T, m db.Model) {
				user1 := utils.Must(m.Query(db.EQ("username", "user1")).Only(h.Ctx()))
				assert.NotNil(t, user1)

				assert.Equal(t, true, user1.Get("deleted"))
				assert.Equal(t, uint(11), user1.Get("age"))
				assert.Equal(t, nil, user1.Get("bio"))
			},
		},
	}

	DBRunUpdateTests(client, t, tests)
}
