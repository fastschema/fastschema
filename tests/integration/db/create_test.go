package db

import (
	"fmt"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func DBCreateNode(t *testing.T, client db.Client) {
	tests := []DBTestCreateData{
		{
			Name:        "fields",
			Schema:      "user",
			InputJSON:   `{ "name": "User 1", "provider": "local", "username": "user1", "age": 10 }`,
			ClearTables: []string{"users"},
			Expect: func(t *testing.T, m db.Model, e *entity.Entity) {
				entity, err := m.Query(db.EQ("id", e.ID())).Only(Ctx())
				assert.NoError(t, err)
				assert.NotNil(t, entity)
				assert.Equal(t, e.ID(), entity.ID())
				assert.Equal(t, "User 1", entity.Get("name"))
				assert.Equal(t, uint(10), entity.Get("age"))
			},
		},
		{
			Name:        "fields/json",
			Schema:      "user",
			ClearTables: []string{"users"},
			InputJSON: `{
				"name": "User 1",
				"username": "user1",
				"provider": "local",
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
			Expect: func(t *testing.T, m db.Model, e *entity.Entity) {
				entity, err := m.Query(db.EQ("id", e.ID())).Only(Ctx())
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

func DBCreateNodeEdges(t *testing.T, client db.Client) {
	tests := []DBTestCreateData{
		{
			Name:   "edges/o2o_two_types",
			Schema: "user",
			InputJSON: `{
				"name": "User 2",
				"username": "user2",
				"provider": "local",
				"car": {
					"id": 2
				},
				"sub_card": {
					"id": 1
				}
			}`,
			ClearTables: []string{"users", "cars", "cards"},
			Prepare: func(t *testing.T) {
				car1ID := utils.Must(utils.Must(client.Model("car")).CreateFromJSON(Ctx(), `{"name": "Car 1"}`))
				car1 := utils.Must(utils.Must(client.Model("car")).Query(db.EQ("id", car1ID)).Only(Ctx()))
				assert.Equal(t, uint64(1), car1.ID())
				utils.Must(utils.Must(client.Model("car")).CreateFromJSON(Ctx(), `{"name": "Car 2"}`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), fmt.Sprintf(`{"name": "User 1", "username": "user1", "provider": "local", "car": {"id": %d} }`, 1)))
				utils.Must(utils.Must(client.Model("card")).CreateFromJSON(Ctx(), fmt.Sprintf(`{"number": "1234567890", "owner": {"id": %d}}`, 1)))
			},
			Expect: func(t *testing.T, m db.Model, e *entity.Entity) {
				entity, err := m.Query(db.EQ("id", e.ID())).Only(Ctx())
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
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 2", "username": "user2", "provider": "local" }`))
			},
			Expect: func(t *testing.T, m db.Model, e *entity.Entity) {
				entity := utils.Must(m.Query(db.EQ("id", e.ID())).Only(Ctx()))
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
				"provider": "local",
				"spouse": {
					"id": 2
				}
			}`,
			ClearTables: []string{"users"},
			Prepare: func(t *testing.T) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 2", "username": "user2", "provider": "local" }`))
			},
			Expect: func(t *testing.T, m db.Model, e *entity.Entity) {
				user2 := utils.Must(m.Query(db.EQ("id", 2)).Only(Ctx()))
				user3 := utils.Must(m.Query(db.EQ("id", e.ID())).Only(Ctx()))
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
				utils.Must(utils.Must(client.Model("node")).CreateFromJSON(Ctx(), `{ "name": "Node 1" }`))
				utils.Must(utils.Must(client.Model("node")).CreateFromJSON(Ctx(), `{ "name": "Node 2" }`))
			},
			Expect: func(t *testing.T, m db.Model, e *entity.Entity) {
				node2 := utils.Must(m.Query(db.EQ("id", 2)).Only(Ctx()))
				node3 := utils.Must(m.Query(db.EQ("id", e.ID())).Only(Ctx()))
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
				utils.Must(utils.Must(client.Model("node")).CreateFromJSON(Ctx(), `{ "name": "Node 1" }`))
				utils.Must(utils.Must(client.Model("node")).CreateFromJSON(Ctx(), `{ "name": "Node 2" }`))
			},
			Expect: func(t *testing.T, m db.Model, e *entity.Entity) {
				node3 := utils.Must(m.Query(db.EQ("id", e.ID())).Only(Ctx()))
				node2 := utils.Must(m.Query(db.EQ("id", 2)).Only(Ctx()))
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
				"provider": "local",
				"sub_pets": [
					{
						"id": 1
					}
				]
			}`,
			ClearTables: []string{"users", "pets"},
			Prepare: func(t *testing.T) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(Ctx(), `{ "name": "Pet 1", "owner_id": 1 }`))
			},
			Expect: func(t *testing.T, m db.Model, e *entity.Entity) {
				pet1 := utils.Must(utils.Must(client.Model("pet")).Query(db.EQ("id", 1)).Only(Ctx()))
				user2 := utils.Must(m.Query(db.EQ("id", e.ID())).Only(Ctx()))
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
				"provider": "local",
				"sub_pets": [
					{ "id": 1 },
					{ "id": 2 }
				]
			}`,
			ClearTables: []string{"users", "pets"},
			Prepare: func(t *testing.T) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(Ctx(), `{ "name": "Pet 1", "owner_id": 1 }`))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(Ctx(), `{ "name": "Pet 2", "owner_id": 1 }`))
			},
			Expect: func(t *testing.T, m db.Model, e *entity.Entity) {
				user2 := utils.Must(m.Query(db.EQ("id", e.ID())).Only(Ctx()))
				pet1 := utils.Must(utils.Must(client.Model("pet")).Query(db.EQ("id", 1)).Only(Ctx()))
				pet2 := utils.Must(utils.Must(client.Model("pet")).Query(db.EQ("id", 2)).Only(Ctx()))

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
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 2", "username": "user2", "provider": "local" }`))
			},
			Expect: func(t *testing.T, m db.Model, e *entity.Entity) {
				pet1 := utils.Must(m.Query(db.EQ("id", e.ID())).Only(Ctx()))
				user2 := utils.Must(utils.Must(client.Model("user")).Query(db.EQ("id", 2)).Only(Ctx()))

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
				utils.Must(utils.Must(client.Model("node")).CreateFromJSON(Ctx(), `{ "name": "Node 1" }`))
			},
			Expect: func(t *testing.T, m db.Model, e *entity.Entity) {
				node1 := utils.Must(m.Query(db.EQ("id", 1)).Only(Ctx()))
				node2 := utils.Must(m.Query(db.EQ("id", e.ID())).Only(Ctx()))

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
				utils.Must(utils.Must(client.Model("node")).CreateFromJSON(Ctx(), `{ "name": "Node 1" }`))
				utils.Must(utils.Must(client.Model("node")).CreateFromJSON(Ctx(), `{ "name": "Node 2" }`))
				utils.Must(utils.Must(client.Model("node")).CreateFromJSON(Ctx(), `{ "name": "Node 3" }`))
			},
			Expect: func(t *testing.T, m db.Model, e *entity.Entity) {
				node4 := utils.Must(m.Query(db.EQ("id", e.ID())).Only(Ctx()))
				node2 := utils.Must(m.Query(db.EQ("id", 2)).Only(Ctx()))
				node3 := utils.Must(m.Query(db.EQ("id", 3)).Only(Ctx()))

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
				utils.Must(utils.Must(client.Model("node")).CreateFromJSON(Ctx(), `{ "name": "Node 1" }`))
				utils.Must(utils.Must(client.Model("node")).CreateFromJSON(Ctx(), `{ "name": "Node 2" }`))
				utils.Must(utils.Must(client.Model("node")).CreateFromJSON(Ctx(), `{ "name": "Node 3" }`))
				utils.Must(utils.Must(client.Model("node")).CreateFromJSON(Ctx(), `{ "name": "Node 4" }`))
			},
			Expect: func(t *testing.T, m db.Model, e *entity.Entity) {
				node1 := utils.Must(m.Query(db.EQ("id", 1)).Only(Ctx()))
				node3 := utils.Must(m.Query(db.EQ("id", 3)).Only(Ctx()))
				node4 := utils.Must(m.Query(db.EQ("id", 4)).Only(Ctx()))
				node5 := utils.Must(m.Query(db.EQ("id", e.ID())).Only(Ctx()))

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
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 2", "username": "user2", "provider": "local" }`))
			},
			Expect: func(t *testing.T, m db.Model, groupEntity *entity.Entity) {
				group1 := utils.Must(m.Query(db.EQ("id", groupEntity.ID())).Only(Ctx()))
				groupsUsers := utils.Must(utils.Must(client.Model("groups_users")).Query(db.EQ("groups", group1.ID())).Get(Ctx()))
				userIDs := utils.Map(groupsUsers, func(gu *entity.Entity) any {
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
				"provider": "local",
				"groups": [
					{ "id": 1 },
					{ "id": 2 }
				]
			}`,
			ClearTables: []string{"users", "groups", "groups_users"},
			Prepare: func(t *testing.T) {
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(Ctx(), `{ "name": "Group 1" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(Ctx(), `{ "name": "Group 2" }`))
			},
			Expect: func(t *testing.T, m db.Model, userEntity *entity.Entity) {
				user1 := utils.Must(m.Query(db.EQ("id", userEntity.ID())).Only(Ctx()))
				groupsUsers := utils.Must(utils.Must(client.Model("groups_users")).Query(db.EQ("users", user1.ID())).Get(Ctx()))
				groupIDs := utils.Map(groupsUsers, func(gui *entity.Entity) any {
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
				"provider": "local",
				"friends": [
					{ "id": 1 },
					{ "id": 2 }
				]
			}`,
			ClearTables: []string{"users", "friends_user"},
			Prepare: func(t *testing.T) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 2", "username": "user2", "provider": "local" }`))
			},
			Expect: func(t *testing.T, m db.Model, userEntity *entity.Entity) {
				user3 := utils.Must(m.Query(db.EQ("id", userEntity.ID())).Only(Ctx()))

				friendsUsers := utils.Must(utils.Must(client.Model("friends_user")).Query(db.EQ("user", user3.ID())).Get(Ctx()))

				friendIDs := utils.Map(friendsUsers, func(fu *entity.Entity) any {
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
				"provider": "local",
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
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 2", "username": "user2", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(Ctx(), `{ "name": "Group 1" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(Ctx(), `{ "name": "Group 2" }`))
			},
			Expect: func(t *testing.T, m db.Model, e *entity.Entity) {
				user3 := utils.Must(m.Query(db.EQ("id", e.ID())).Only(Ctx()))

				friendsUsers := utils.Must(utils.Must(client.Model("friends_user")).Query(db.EQ("user", user3.ID())).Get(Ctx()))

				friendIDs := utils.Map(friendsUsers, func(fu *entity.Entity) any {
					return fu.Get("friends")
				})

				groupsUsers := utils.Must(utils.Must(client.Model("groups_users")).Query(db.EQ("users", user3.ID())).Get(Ctx()))
				groupIDs := utils.Map(groupsUsers, func(gu *entity.Entity) any {
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
