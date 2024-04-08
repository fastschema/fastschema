package db

import (
	"errors"
	"fmt"
	"testing"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

func DBCreateNode(t *testing.T, client app.DBClient) {
	tests := []DBTestCreateData{
		{
			Name:        "fields",
			Schema:      "user",
			InputJSON:   `{ "name": "User 1", "username": "user1", "age": 10 }`,
			ClearTables: []string{"users"},
			Expect: func(t *testing.T, m app.Model, e *schema.Entity) {
				entity, err := m.Query(app.EQ("id", e.ID())).Only()
				assert.NoError(t, err)
				assert.NotNil(t, entity)
				assert.Equal(t, e.ID(), entity.ID())
				assert.Equal(t, "User 1", entity.Get("name"))
				assert.Equal(t, uint(10), entity.Get("age"))
			},
		},
		{
			Name:        "fields/user-defined-id",
			Schema:      "user",
			ClearTables: []string{"users"},
			InputJSON:   `{ "name": "User 2", "username": "user2", "age": 20, "id": 2 }`,
			WantErr:     true,
			ExpectError: errors.New("cannot create entity with existing ID 2"),
			Run: func(model app.Model, entity *schema.Entity) (*schema.Entity, error) {
				createdEntityID, err := model.Create(entity)
				if err != nil {
					return nil, err
				}
				return model.Query(app.EQ("id", createdEntityID)).First()
			},
		},
		{
			Name:        "fields/json",
			Schema:      "user",
			ClearTables: []string{"users"},
			InputJSON: `{
				"name": "User 1",
				"username": "user1",
				"json": {
					"key1": "value1",
					"key2": "value2",
					"nested1": {
						"nested1_key_1": "nested1_value_1",
						"nested1_key_2": {
							"nested1_key_2_key_1": "nested1_key_2_value_1",
							"nested1_key_2_key_2": {
								"nested1_key_2_key_2_key_1": "nested1_key_2_key_2_value_1",
								"nested1_key_2_key_2_key_2": {
									"nested1_key_2_key_2_key_2_key_1": "nested1_key_2_key_2_key_2_value_1",
									"nested1_key_2_key_2_key_2_key_2": "nested1_key_2_key_2_key_2_value_2"
								}
							}
						}
					},
					"array1": [
						"array_value_1",
						"array_value_2"
					],
					"array2": [
						{
							"array2_key_1": "array2_value_1"
						},
						{
							"array2_key_2": "array2_value_2"
						}
					]
				}
			}`,
			Expect: func(t *testing.T, m app.Model, e *schema.Entity) {
				entity, err := m.Query(app.EQ("id", e.ID())).Only()
				assert.NoError(t, err)
				assert.Equal(t, e.ID(), entity.ID())
				assert.Equal(t, "User 1", entity.Get("name"))

				jsonFieldData := entity.Get("json")
				assert.Equal(t, "value1", utils.Pick(jsonFieldData, "key1"))
				assert.Equal(t, "array2_value_2", utils.Pick(jsonFieldData, "array2.1.array2_key_2"))
				assert.Equal(t, "nested1_key_2_key_2_key_2_value_2", utils.Pick(
					jsonFieldData,
					"nested1.nested1_key_2.nested1_key_2_key_2.nested1_key_2_key_2_key_2.nested1_key_2_key_2_key_2_key_2",
				))
			},
		},
	}

	DBRunCreateTests(client, t, tests)
}

func DBCreateNodeEdges(t *testing.T, client app.DBClient) {
	tests := []DBTestCreateData{
		{
			Name:   "edges/o2o_two_types",
			Schema: "user",
			InputJSON: `{
				"name": "User 2",
				"username": "user2",
				"car": {
					"id": 2
				},
				"sub_card": {
					"id": 1
				}
			}`,
			ClearTables: []string{"users", "cars", "cards"},
			Prepare: func(t *testing.T) {
				car1ID := utils.Must(utils.Must(client.Model("car")).CreateFromJSON(`{"name": "Car 1"}`))
				car1 := utils.Must(utils.Must(client.Model("car")).Query(app.EQ("id", car1ID)).Only())
				assert.Equal(t, uint64(1), car1.ID())
				utils.Must(utils.Must(client.Model("car")).CreateFromJSON(`{"name": "Car 2"}`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(fmt.Sprintf(`{"name": "User 1", "username": "user1", "car": {"id": %d} }`, 1)))
				utils.Must(utils.Must(client.Model("card")).CreateFromJSON(fmt.Sprintf(`{"number": "1234567890", "owner": {"id": %d}}`, 1)))
			},
			Expect: func(t *testing.T, m app.Model, e *schema.Entity) {
				entity, err := m.Query(app.EQ("id", e.ID())).Only()
				assert.NoError(t, err)
				assert.Equal(t, uint64(2), entity.ID())
				assert.Equal(t, "User 2", entity.Get("name"))
			},
		},
		{
			Name:   "edges/o2o_two_types/inverse",
			Schema: "card",
			InputJSON: `{
				"number": "0001",
				"owner": {
					"id": 2
					}
				}`,
			ClearTables: []string{"users", "cards"},
			Prepare: func(t *testing.T) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(`{ "name": "User 1", "username": "user1" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(`{ "name": "User 2", "username": "user2" }`))
			},
			Expect: func(t *testing.T, m app.Model, e *schema.Entity) {
				entity := utils.Must(m.Query(app.EQ("id", e.ID())).Only())
				assert.Equal(t, uint64(1), entity.ID())
				assert.Equal(t, "0001", entity.Get("number"))
				assert.Equal(t, uint64(2), entity.Get("owner_id"))
			},
		},
		{
			Name:   "edges/o2o_same_types/bidi",
			Schema: "user",
			InputJSON: `{
				"name": "User 3",
				"username": "user3",
				"spouse": {
					"id": 2
				}
			}`,
			ClearTables: []string{"users"},
			Prepare: func(t *testing.T) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(`{ "name": "User 1", "username": "user1" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(`{ "name": "User 2", "username": "user2", }`))
			},
			Expect: func(t *testing.T, m app.Model, e *schema.Entity) {
				user2 := utils.Must(m.Query(app.EQ("id", 2)).Only())
				user3 := utils.Must(m.Query(app.EQ("id", e.ID())).Only())
				assert.Equal(t, uint64(3), user3.ID())
				assert.Equal(t, user2.ID(), user3.Get("spouse_id"))
				assert.Equal(t, user3.ID(), user2.Get("spouse_id"))
			},
		},
		{
			Name:   "edges/o2o_same_types/recursive",
			Schema: "node",
			InputJSON: `{
				"name": "Node 3",
				"prev": {
					"id": 2
				}
			}`,
			ClearTables: []string{"nodes"},
			Prepare: func(t *testing.T) {
				utils.Must(utils.Must(client.Model("node")).CreateFromJSON(`{ "name": "Node 1" }`))
				utils.Must(utils.Must(client.Model("node")).CreateFromJSON(`{ "name": "Node 2" }`))
			},
			Expect: func(t *testing.T, m app.Model, e *schema.Entity) {
				node2 := utils.Must(m.Query(app.EQ("id", 2)).Only())
				node3 := utils.Must(m.Query(app.EQ("id", e.ID())).Only())
				assert.Equal(t, uint64(3), node3.ID())
				assert.Equal(t, node2.ID(), node3.Get("prev_id"))
				assert.Equal(t, "Node 3", node3.Get("name"))
			},
		},
		{
			Name:   "edges/o2o_same_types/recursive/inverse",
			Schema: "node",
			InputJSON: `{
				"name": "Node 3",
				"next": {
					"id": 2
				}
			}`,
			ClearTables: []string{"nodes"},
			Prepare: func(t *testing.T) {
				utils.Must(utils.Must(client.Model("node")).CreateFromJSON(`{ "name": "Node 1" }`))
				utils.Must(utils.Must(client.Model("node")).CreateFromJSON(`{ "name": "Node 2" }`))
			},
			Expect: func(t *testing.T, m app.Model, e *schema.Entity) {
				node3 := utils.Must(m.Query(app.EQ("id", e.ID())).Only())
				node2 := utils.Must(m.Query(app.EQ("id", 2)).Only())
				assert.Equal(t, uint64(3), node3.ID())
				assert.Equal(t, node3.ID(), node2.Get("prev_id"))
				assert.Equal(t, "Node 3", node3.Get("name"))
			},
		},
		{
			Name:   "edges/o2m_two_types",
			Schema: "user",
			InputJSON: `{
				"name": "User 2",
				"username": "user2",
				"sub_pets": [
					{
						"id": 1
					}
				]
			}`,
			ClearTables: []string{"users", "pets"},
			Prepare: func(t *testing.T) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(`{ "name": "User 1", "username": "user1" }`))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(`{ "name": "Pet 1", "owner_id": 1 }`))
			},
			Expect: func(t *testing.T, m app.Model, e *schema.Entity) {
				pet1 := utils.Must(utils.Must(client.Model("pet")).Query(app.EQ("id", 1)).Only())
				user2 := utils.Must(m.Query(app.EQ("id", e.ID())).Only())
				assert.Equal(t, uint64(2), user2.ID())
				assert.Equal(t, pet1.Get("sub_owner_id"), user2.ID())
				assert.Equal(t, "User 2", user2.Get("name"))
			},
		},
		{
			Name:   "edges/o2m_two_types/multiple",
			Schema: "user",
			InputJSON: `{
				"name": "User 2",
				"username": "user2",
				"sub_pets": [
					{ "id": 1 },
					{ "id": 2 }
				]
			}`,
			ClearTables: []string{"users", "pets"},
			Prepare: func(t *testing.T) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(`{ "name": "User 1", "username": "user1" }`))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(`{ "name": "Pet 1", "owner_id": 1 }`))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(`{ "name": "Pet 2", "owner_id": 1 }`))
			},
			Expect: func(t *testing.T, m app.Model, e *schema.Entity) {
				user2 := utils.Must(m.Query(app.EQ("id", e.ID())).Only())
				pet1 := utils.Must(utils.Must(client.Model("pet")).Query(app.EQ("id", 1)).Only())
				pet2 := utils.Must(utils.Must(client.Model("pet")).Query(app.EQ("id", 2)).Only())

				assert.Equal(t, uint64(2), user2.ID())
				assert.Equal(t, pet1.Get("sub_owner_id"), user2.ID())
				assert.Equal(t, pet2.Get("sub_owner_id"), user2.ID())
			},
		},
		{
			Name:   "edges/o2m_two_types/inverse",
			Schema: "pet",
			InputJSON: `{
				"name": "Pet 1",
				"owner": {
					"id": 2
				}
			}`,
			ClearTables: []string{"users", "pets"},
			Prepare: func(t *testing.T) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(`{ "name": "User 1", "username": "user1" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(`{ "name": "User 2", "username": "user2" }`))
			},
			Expect: func(t *testing.T, m app.Model, e *schema.Entity) {
				pet1 := utils.Must(m.Query(app.EQ("id", e.ID())).Only())
				user2 := utils.Must(utils.Must(client.Model("user")).Query(app.EQ("id", 2)).Only())

				assert.Equal(t, uint64(1), pet1.ID())
				assert.Equal(t, pet1.Get("owner_id"), user2.ID())
			},
		},
		{
			Name:   "edges/o2m_same_types",
			Schema: "node",
			InputJSON: `{
				"name": "Node 2",
				"parent": {
					"id": 1
				}
			}`,
			ClearTables: []string{"nodes"},
			Prepare: func(t *testing.T) {
				utils.Must(utils.Must(client.Model("node")).CreateFromJSON(`{ "name": "Node 1" }`))
			},
			Expect: func(t *testing.T, m app.Model, e *schema.Entity) {
				node1 := utils.Must(m.Query(app.EQ("id", 1)).Only())
				node2 := utils.Must(m.Query(app.EQ("id", e.ID())).Only())

				assert.Equal(t, uint64(2), node2.ID())
				assert.Equal(t, node2.Get("parent_id"), node1.ID())
			},
		},
		{
			Name:   "edges/o2m_same_types/inverse",
			Schema: "node",
			InputJSON: `{
				"name": "Node 4",
				"children": [
					{ 
						"id": 2
					},
					{
						"id": 3
					}
				]
			}`,
			ClearTables: []string{"nodes"},
			Prepare: func(t *testing.T) {
				utils.Must(utils.Must(client.Model("node")).CreateFromJSON(`{ "name": "Node 1" }`))
				utils.Must(utils.Must(client.Model("node")).CreateFromJSON(`{ "name": "Node 2" }`))
				utils.Must(utils.Must(client.Model("node")).CreateFromJSON(`{ "name": "Node 3" }`))
			},
			Expect: func(t *testing.T, m app.Model, e *schema.Entity) {
				node4 := utils.Must(m.Query(app.EQ("id", e.ID())).Only())
				node2 := utils.Must(m.Query(app.EQ("id", 2)).Only())
				node3 := utils.Must(m.Query(app.EQ("id", 3)).Only())

				assert.Equal(t, uint64(4), node4.ID())
				assert.Equal(t, "Node 4", node4.Get("name"))
				assert.Equal(t, node2.Get("parent_id"), node4.ID())
				assert.Equal(t, node3.Get("parent_id"), node4.ID())
			},
		},
		{
			Name:   "edges/o2m_same_types/both",
			Schema: "node",
			InputJSON: `{
				"name": "Node 5",
				"parent": {
					"id": 1
				},
				"children": [
					{
						"id": 3
					},
					{
						"id": 4
					}
				]
			}`,
			ClearTables: []string{"nodes"},
			Prepare: func(t *testing.T) {
				utils.Must(utils.Must(client.Model("node")).CreateFromJSON(`{ "name": "Node 1" }`))
				utils.Must(utils.Must(client.Model("node")).CreateFromJSON(`{ "name": "Node 2" }`))
				utils.Must(utils.Must(client.Model("node")).CreateFromJSON(`{ "name": "Node 3" }`))
				utils.Must(utils.Must(client.Model("node")).CreateFromJSON(`{ "name": "Node 4" }`))
			},
			Expect: func(t *testing.T, m app.Model, e *schema.Entity) {
				node1 := utils.Must(m.Query(app.EQ("id", 1)).Only())
				node3 := utils.Must(m.Query(app.EQ("id", 3)).Only())
				node4 := utils.Must(m.Query(app.EQ("id", 4)).Only())
				node5 := utils.Must(m.Query(app.EQ("id", e.ID())).Only())

				assert.Equal(t, uint64(5), node5.ID())
				assert.Equal(t, "Node 5", node5.Get("name"))
				assert.Equal(t, node5.Get("parent_id"), node1.ID())
				assert.Equal(t, node3.Get("parent_id"), node5.ID())
				assert.Equal(t, node4.Get("parent_id"), node5.ID())
			},
		},
		{
			Name:   "edges/m2m",
			Schema: "group",
			InputJSON: `{
				"name": "Group 1",
				"users": [
					{ "id": 1 },
					{ "id": 2 }
				]
			}`,
			ClearTables: []string{"users", "groups", "groups_users"},
			Prepare: func(t *testing.T) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(`{ "name": "User 1", "username": "user1" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(`{ "name": "User 2", "username": "user2" }`))
			},
			Expect: func(t *testing.T, m app.Model, groupEntity *schema.Entity) {
				group1 := utils.Must(m.Query(app.EQ("id", groupEntity.ID())).Only())
				groupsUsers := utils.Must(utils.Must(client.Model("groups_users")).Query(app.EQ("groups", group1.ID())).Get())
				userIDs := utils.Map(groupsUsers, func(gu *schema.Entity) any {
					return gu.Get("users")
				})

				assert.Equal(t, uint64(1), group1.ID())
				assert.Equal(t, "Group 1", group1.Get("name"))
				assert.Equal(t, userIDs, []any{uint64(1), uint64(2)})
			},
		},
		{
			Name:   "edges/m2m/inverse",
			Schema: "user",
			InputJSON: `{
				"name": "User 1",
				"username": "user1",
				"groups": [
					{ "id": 1 },
					{ "id": 2 }
				]
			}`,
			ClearTables: []string{"users", "groups", "groups_users"},
			Prepare: func(t *testing.T) {
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(`{ "name": "Group 1" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(`{ "name": "Group 2" }`))
			},
			Expect: func(t *testing.T, m app.Model, userEntity *schema.Entity) {
				user1 := utils.Must(m.Query(app.EQ("id", userEntity.ID())).Only())
				groupsUsers := utils.Must(utils.Must(client.Model("groups_users")).Query(app.EQ("users", user1.ID())).Get())
				groupIDs := utils.Map(groupsUsers, func(gui *schema.Entity) any {
					return gui.Get("groups")
				})

				assert.Equal(t, "User 1", user1.Get("name"))
				assert.Equal(t, uint64(1), user1.ID())
				assert.Equal(t, groupIDs, []any{uint64(1), uint64(2)})
			},
		},
		{
			Name:   "edges/m2m/bidi",
			Schema: "user",
			InputJSON: `{
				"name": "User 3",
				"username": "user3",
				"friends": [
					{ "id": 1 },
					{ "id": 2 }
				]
			}`,
			ClearTables: []string{"users", "friends_user"},
			Prepare: func(t *testing.T) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(`{ "name": "User 1", "username": "user1" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(`{ "name": "User 2", "username": "user2" }`))
			},
			Expect: func(t *testing.T, m app.Model, userEntity *schema.Entity) {
				user3 := utils.Must(m.Query(app.EQ("id", userEntity.ID())).Only())

				friendsUsers := utils.Must(utils.Must(client.Model("friends_user")).Query(app.EQ("user", user3.ID())).Get())

				friendIDs := utils.Map(friendsUsers, func(fu *schema.Entity) any {
					return fu.Get("friends")
				})

				assert.Equal(t, "User 3", user3.Get("name"))
				assert.Equal(t, friendIDs, []any{uint64(1), uint64(2)})
				assert.Equal(t, uint64(3), user3.ID())
			},
		},
		{
			Name:   "edges/m2m/bidi/batch",
			Schema: "user",
			InputJSON: `{
				"name": "User 3",
				"username": "user3",
				"friends": [
					{ "id": 1 },
					{ "id": 2 }
				],
				"groups": [
					{ "id": 1 },
					{ "id": 2 }
				]
			}`,
			ClearTables: []string{"users", "groups", "groups_users", "friends_user"},
			Prepare: func(t *testing.T) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(`{ "name": "User 1", "username": "user1" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(`{ "name": "User 2", "username": "user2" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(`{ "name": "Group 1" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(`{ "name": "Group 2" }`))
			},
			Expect: func(t *testing.T, m app.Model, e *schema.Entity) {
				user3 := utils.Must(m.Query(app.EQ("id", e.ID())).Only())

				friendsUsers := utils.Must(utils.Must(client.Model("friends_user")).Query(app.EQ("user", user3.ID())).Get())

				friendIDs := utils.Map(friendsUsers, func(fu *schema.Entity) any {
					return fu.Get("friends")
				})

				groupsUsers := utils.Must(utils.Must(client.Model("groups_users")).Query(app.EQ("users", user3.ID())).Get())
				groupIDs := utils.Map(groupsUsers, func(gu *schema.Entity) any {
					return gu.Get("groups")
				})

				assert.Equal(t, uint64(3), user3.ID())
				assert.Equal(t, "User 3", user3.Get("name"))
				assert.Equal(t, friendIDs, []any{uint64(1), uint64(2)})
				assert.Equal(t, groupIDs, []any{uint64(1), uint64(2)})
			},
		},
	}

	DBRunCreateTests(client, t, tests)
}
