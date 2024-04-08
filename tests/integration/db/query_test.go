package db

import (
	"testing"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

func DBQueryNode(t *testing.T, client app.DBClient) {
	tests := []DBTestQueryData{
		{
			Name:        "Query_with_no_filter",
			Schema:      "user",
			ClearTables: []string{"users"},
			Prepare: func(t *testing.T, client app.DBClient, m app.Model) []*schema.Entity {
				user1ID, err1 := utils.Must(m.Mutation()).Create(schema.NewEntity().Set("name", "John Doe").Set("username", "john"))
				assert.NoError(t, err1)
				user2ID, err2 := utils.Must(m.Mutation()).Create(schema.NewEntity().Set("name", "Jane Doe").Set("username", "jane"))
				assert.NoError(t, err2)
				return []*schema.Entity{
					utils.Must(m.Query(app.EQ("id", user1ID)).First()),
					utils.Must(m.Query(app.EQ("id", user2ID)).First()),
				}
			},
		},
		{
			Name:   "Query_with_filter",
			Schema: "user",
			Filter: `{
				"age": {
					"$gt": 5
				}
			}`,
			ClearTables: []string{"users"},
			Prepare: func(t *testing.T, client app.DBClient, m app.Model) []*schema.Entity {
				user1ID, err1 := utils.Must(m.Mutation()).Create(schema.NewEntity().Set("name", "John Doe").Set("age", 10).Set("username", "john"))
				assert.NoError(t, err1)
				_, err2 := utils.Must(m.Mutation()).Create(schema.NewEntity().Set("name", "Jane Doe").Set("age", 2).Set("username", "jane"))
				assert.NoError(t, err2)
				return []*schema.Entity{utils.Must(m.Query(app.EQ("id", user1ID)).First())}
			},
		},
		{
			Name:        "Query_with_limit_offset_and_order",
			Schema:      "user",
			Limit:       3,
			Offset:      2,
			Order:       []string{"-name", "id"},
			ClearTables: []string{"users"},
			Prepare: func(t *testing.T, client app.DBClient, m app.Model) []*schema.Entity {
				user1ID := utils.Must(m.CreateFromJSON(`{"name": "User 1", "username": "user1"}`))
				assert.True(t, user1ID > 0)
				user22ID := utils.Must(m.CreateFromJSON(`{"name": "User 2 2", "username": "user22"}`))
				assert.True(t, user22ID > 0)
				user2ID := utils.Must(m.CreateFromJSON(`{"name": "User 2", "username": "user2"}`))
				assert.True(t, user2ID > 0)
				user3ID := utils.Must(m.CreateFromJSON(`{"name": "User 3", "username": "user3"}`))
				assert.True(t, user3ID > 0)
				user4ID := utils.Must(m.CreateFromJSON(`{"name": "User 4", "username": "user4"}`))
				assert.True(t, user4ID > 0)
				return []*schema.Entity{
					utils.Must(m.Query(app.EQ("id", user22ID)).First()),
					utils.Must(m.Query(app.EQ("id", user2ID)).First()),
					utils.Must(m.Query(app.EQ("id", user1ID)).First()),
				}
			},
		},
		{
			Name:   "Query_with_columns",
			Schema: "user",
			Columns: []string{
				"id",
				"name",
			},
			ClearTables: []string{"users"},
			Prepare: func(t *testing.T, client app.DBClient, m app.Model) []*schema.Entity {
				user1ID := utils.Must(m.CreateFromJSON(`{"name": "User 1", "age": 10, "username": "user1"}`))
				assert.True(t, user1ID > 0)
				return []*schema.Entity{schema.NewEntity(user1ID).Set("name", "User 1")}
			},
		},
		{
			Name:   "Query_with_relation_filter",
			Schema: "car",
			Filter: `{
				"owner.groups.name": "Group 2"
			}`,
			ClearTables: []string{"groups_users", "users", "groups", "cars"},
			Prepare: func(t *testing.T, client app.DBClient, m app.Model) []*schema.Entity {
				group1ID := utils.Must(utils.Must(client.Model("group")).CreateFromJSON(`{"name": "Group 1"}`))
				assert.True(t, group1ID > 0)
				group2ID := utils.Must(utils.Must(client.Model("group")).CreateFromJSON(`{"name": "Group 2"}`))
				assert.True(t, group2ID > 0)

				user1ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(`{"name": "User 1", "groups": [{"id": 1}], "username": "user1"}`))
				assert.True(t, user1ID > 0)
				user2ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(`{"name": "User 2", "groups": [{"id": 2}], "username": "user2"}`))
				assert.True(t, user2ID > 0)

				car1ID := utils.Must(m.CreateFromJSON(`{"name": "Car 1", "owner": {"id": 1}}`))
				assert.True(t, car1ID > 0)
				car2ID := utils.Must(m.CreateFromJSON(`{"name": "Car 2", "owner": {"id": 2}}`))
				assert.True(t, car2ID > 0)

				car2 := utils.Must(m.Query(app.EQ("id", car2ID)).First())
				return []*schema.Entity{car2}
			},
		},
		{
			Name:   "Query_with_edges_O2M_two_types",
			Schema: "user",
			Columns: []string{
				"id",
				"name",
				"pets",
			},
			ClearTables: []string{"users", "pets"},
			Prepare: func(t *testing.T, client app.DBClient, m app.Model) []*schema.Entity {
				user1ID := utils.Must(m.CreateFromJSON(`{"name": "User 1", "username": "user1"}`))
				assert.True(t, user1ID > 0)
				pet1ID := utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(`{"name": "Pet 1", "owner": {"id": 1}}`))
				assert.True(t, pet1ID > 0)
				pet2ID := utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(`{"name": "Pet 2", "owner": {"id": 1}}`))
				assert.True(t, pet2ID > 0)
				return []*schema.Entity{schema.NewEntity(user1ID).Set("name", "User 1").Set("pets", []*schema.Entity{
					utils.Must(utils.Must(client.Model("pet")).Query(app.EQ("id", pet1ID)).First()),
					utils.Must(utils.Must(client.Model("pet")).Query(app.EQ("id", pet2ID)).First()),
				})}
			},
		},
		{
			Name:   "Query_with_edges_O2M_two_types_reverse",
			Schema: "pet",
			Columns: []string{
				"id",
				"name",
				"owner",
			},
			ClearTables: []string{"users", "pets"},
			Prepare: func(t *testing.T, client app.DBClient, m app.Model) []*schema.Entity {
				user1ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(`{"name": "User 1", "username": "user1"}`))
				assert.True(t, user1ID > 0)
				pet1ID := utils.Must(m.CreateFromJSON(`{"name": "Pet 1", "owner": {"id": 1}}`))
				assert.True(t, pet1ID > 0)
				pet2ID := utils.Must(m.CreateFromJSON(`{"name": "Pet 2", "owner": {"id": 1}}`))
				assert.True(t, pet2ID > 0)
				user1 := utils.Must(utils.Must(client.Model("user")).Query(app.EQ("id", user1ID)).First())
				return []*schema.Entity{
					schema.NewEntity(pet1ID).Set("name", "Pet 1").Set("owner_id", uint64(1)).Set("owner", user1),
					schema.NewEntity(pet2ID).Set("name", "Pet 2").Set("owner_id", uint64(1)).Set("owner", user1),
				}
			},
		},
		{
			Name:        "Query_with_edges_O2M_same_type",
			Schema:      "node",
			Columns:     []string{"name", "children"},
			ClearTables: []string{"nodes"},
			Filter: `{
				"id": 1,
			}`,
			Prepare: func(t *testing.T, client app.DBClient, m app.Model) []*schema.Entity {
				node1ID := utils.Must(m.CreateFromJSON(`{"name": "Node 1"}`))
				assert.True(t, node1ID > 0)
				node2ID := utils.Must(m.CreateFromJSON(`{"name": "Node 2", "parent": {"id": 1}}`))
				assert.True(t, node2ID > 0)
				node3ID := utils.Must(m.CreateFromJSON(`{"name": "Node 3", "parent": {"id": 1}}`))
				assert.True(t, node3ID > 0)
				return []*schema.Entity{
					schema.NewEntity(node1ID).Set("name", "Node 1").Set("children", []*schema.Entity{
						utils.Must(m.Query(app.EQ("id", node2ID)).First()),
						utils.Must(m.Query(app.EQ("id", node3ID)).First()),
					}),
				}
			},
		},
		{
			Name:    "Query_with_edges_O2M_same_type_reverse",
			Schema:  "node",
			Columns: []string{"name", "parent"},
			Filter: `{
				"id": {
					"$in": [3, 4]
				}
			}`,
			ClearTables: []string{"nodes"},
			Prepare: func(t *testing.T, client app.DBClient, m app.Model) []*schema.Entity {
				node1ID := utils.Must(m.CreateFromJSON(`{"name": "Node 1"}`))
				assert.True(t, node1ID > 0)
				node2ID := utils.Must(m.CreateFromJSON(`{"name": "Node 2"}`))
				assert.True(t, node2ID > 0)
				node3ID := utils.Must(m.CreateFromJSON(`{"name": "Node 3", "parent": {"id": 1}}`))
				assert.True(t, node3ID > 0)
				node4ID := utils.Must(m.CreateFromJSON(`{"name": "Node 4", "parent": {"id": 2}}`))
				assert.True(t, node4ID > 0)
				return []*schema.Entity{
					schema.NewEntity(node3ID).
						Set("name", "Node 3").
						Set("parent_id", uint64(1)).
						Set("parent", utils.Must(m.Query(app.EQ("id", node1ID)).First())),
					schema.NewEntity(node4ID).
						Set("name", "Node 4").
						Set("parent_id", uint64(2)).
						Set("parent", utils.Must(m.Query(app.EQ("id", node2ID)).First())),
				}
			},
		},
		{
			Name:    "Query_with_edges_O2O_two_types",
			Schema:  "user",
			Columns: []string{"name", "card", "sub_card"},
			Filter: `{
				"id": {
					"$in": [1, 2]
				}
			}`,
			ClearTables: []string{"users", "cards"},
			Prepare: func(t *testing.T, client app.DBClient, m app.Model) []*schema.Entity {
				user1ID := utils.Must(m.CreateFromJSON(`{"name": "User 1", "username": "user1"}`))
				user2ID := utils.Must(m.CreateFromJSON(`{"name": "User 2", "username": "user2"}`))
				assert.True(t, user1ID > 0)
				card1ID := utils.Must(utils.Must(client.Model("card")).CreateFromJSON(`{"number": "1234", "owner": {"id": 1}, "sub_owner": {"id": 2}}`))
				assert.True(t, card1ID > 0)
				card1 := utils.Must(utils.Must(client.Model("card")).Query(app.EQ("id", card1ID)).First())
				return []*schema.Entity{
					schema.NewEntity(user1ID).Set("name", "User 1").Set("card", card1),
					schema.NewEntity(user2ID).Set("name", "User 2").Set("sub_card", card1),
				}
			},
		},
		{
			Name:    "Query_with_edges_O2O_two_types_reverse",
			Schema:  "card",
			Columns: []string{"number", "owner", "sub_owner"},
			Filter: `{
				"id": {
					"$in": [1, 2]
				}
			}`,
			ClearTables: []string{"users", "cards"},
			Prepare: func(t *testing.T, client app.DBClient, m app.Model) []*schema.Entity {
				user1ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(`{"name": "User 1", "username": "user1"}`))
				user2ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(`{"name": "User 2", "username": "user2"}`))
				assert.True(t, user1ID > 0)

				card1ID := utils.Must(m.CreateFromJSON(`{"number": "1234", "owner": {"id": 1}, "sub_owner": {"id": 2}}`))
				assert.True(t, card1ID > 0)
				card2ID := utils.Must(m.CreateFromJSON(`{"number": "5678", "owner": {"id": 2}}`))

				user1 := utils.Must(utils.Must(client.Model("user")).Query(app.EQ("id", user1ID)).First())
				user2 := utils.Must(utils.Must(client.Model("user")).Query(app.EQ("id", user2ID)).First())

				return []*schema.Entity{
					schema.NewEntity(card1ID).
						Set("number", "1234").
						Set("owner_id", uint64(1)).
						Set("sub_owner_id", uint64(2)).
						Set("owner", user1).
						Set("sub_owner_id", uint64(2)).
						Set("sub_owner", user2),
					schema.NewEntity(card2ID).
						Set("number", "5678").
						Set("owner_id", uint64(2)).
						Set("owner", user2),
				}
			},
		},
		{
			Name:        "Query_with_edges_O2O_same_type",
			Schema:      "node",
			Columns:     []string{"name", "next", "prev_id"},
			ClearTables: []string{"nodes"},
			Prepare: func(t *testing.T, client app.DBClient, m app.Model) []*schema.Entity {
				node1ID := utils.Must(m.CreateFromJSON(`{"name": "Node 1"}`))
				assert.True(t, node1ID > 0)
				node2ID := utils.Must(m.CreateFromJSON(`{"name": "Node 2"}`))
				assert.True(t, node2ID > 0)
				node3ID := utils.Must(m.CreateFromJSON(`{"name": "Node 3"}`))
				assert.True(t, node3ID > 0)
				node4ID := utils.Must(m.CreateFromJSON(`{"name": "Node 4"}`))
				assert.True(t, node4ID > 0)

				a1 := utils.Must(utils.Must(m.Mutation()).Where(app.EQ("id", node3ID)).Update(schema.NewEntity().Set("prev_id", 1)))
				assert.Equal(t, 1, a1)

				a2 := utils.Must(utils.Must(m.Mutation()).Where(app.EQ("id", node4ID)).Update(schema.NewEntity().Set("prev_id", 2)))
				assert.Equal(t, 1, a2)

				return []*schema.Entity{
					schema.NewEntity(node1ID).
						Set("name", "Node 1").
						Set("next", utils.Must(m.Query(app.EQ("id", node3ID)).First())),
					schema.NewEntity(node2ID).
						Set("name", "Node 2").
						Set("next", utils.Must(m.Query(app.EQ("id", node4ID)).First())),
					schema.NewEntity(node3ID).
						Set("name", "Node 3").
						Set("prev_id", uint64(1)),
					schema.NewEntity(node4ID).
						Set("name", "Node 4").
						Set("prev_id", uint64(2)),
				}
			},
		},
		{
			Name:        "Query_with_edges_O2O_same_type_reverse",
			Schema:      "node",
			Columns:     []string{"name", "prev", "prev_id"},
			ClearTables: []string{"nodes"},
			Prepare: func(t *testing.T, client app.DBClient, m app.Model) []*schema.Entity {
				node1ID := utils.Must(m.CreateFromJSON(`{"name": "Node 1"}`))
				assert.True(t, node1ID > 0)
				node2ID := utils.Must(m.CreateFromJSON(`{"name": "Node 2"}`))
				assert.True(t, node2ID > 0)
				node3ID := utils.Must(m.CreateFromJSON(`{"name": "Node 3", "prev_id": 1}`))
				assert.True(t, node3ID > 0)
				node4ID := utils.Must(m.CreateFromJSON(`{"name": "Node 4", "prev_id": 2}`))
				assert.True(t, node4ID > 0)

				return []*schema.Entity{
					schema.NewEntity(node1ID).Set("name", "Node 1"),
					schema.NewEntity(node2ID).Set("name", "Node 2"),
					schema.NewEntity(node3ID).
						Set("name", "Node 3").
						Set("prev_id", uint64(1)).
						Set("prev", utils.Must(m.Query(app.EQ("id", 1)).First())),
					schema.NewEntity(node4ID).
						Set("name", "Node 4").
						Set("prev_id", uint64(2)).
						Set("prev", utils.Must(m.Query(app.EQ("id", 2)).First())),
				}
			},
		},
		{
			Name:        "Query_with_edges_O2O_bidi",
			Schema:      "user",
			Columns:     []string{"name", "spouse"},
			ClearTables: []string{"users"},
			Prepare: func(t *testing.T, client app.DBClient, m app.Model) []*schema.Entity {
				user1ID := utils.Must(m.CreateFromJSON(`{"name": "User 1", "username": "user1"}`))
				assert.True(t, user1ID > 0)
				user2ID := utils.Must(m.CreateFromJSON(`{"name": "User 2", "username": "user2"}`))
				assert.True(t, user2ID > 0)

				a1 := utils.Must(utils.Must(m.Mutation()).Where(app.EQ("id", user1ID)).Update(schema.NewEntity().Set("spouse_id", user2ID)))
				assert.Equal(t, 1, a1)

				a2 := utils.Must(utils.Must(m.Mutation()).Where(app.EQ("id", user2ID)).Update(schema.NewEntity().Set("spouse_id", user1ID)))
				assert.Equal(t, 1, a2)

				return []*schema.Entity{
					schema.NewEntity(user1ID).Set("name", "User 1").Set("spouse_id", uint64(2)).Set("spouse", utils.Must(m.Query(app.EQ("id", user2ID)).First())),
					schema.NewEntity(user2ID).Set("name", "User 2").Set("spouse_id", uint64(1)).Set("spouse", utils.Must(m.Query(app.EQ("id", user1ID)).First())),
				}
			},
		},
		{
			Name:        "Query_with_edges_M2M_two_types",
			Schema:      "group",
			Columns:     []string{"name", "users"},
			ClearTables: []string{"groups", "users", "groups_users"},
			Prepare: func(t *testing.T, client app.DBClient, m app.Model) []*schema.Entity {
				userModel := utils.Must(client.Model("user"))
				group1ID := utils.Must(m.CreateFromJSON(`{"name": "Group 1"}`))
				assert.True(t, group1ID > 0)
				group2ID := utils.Must(m.CreateFromJSON(`{"name": "Group 2"}`))
				assert.True(t, group2ID > 0)

				group3ID := utils.Must(m.CreateFromJSON(`{"name": "Group 3"}`))
				assert.True(t, group3ID > 0)

				user1ID := utils.Must(userModel.CreateFromJSON(`{"name": "User 1", "username": "user1", "groups": [{"id": 1}]}`))
				assert.True(t, user1ID > 0)
				user2ID := utils.Must(userModel.CreateFromJSON(`{"name": "User 2", "username": "user2", "groups": [{"id": 1}, {"id": 2}]}`))
				assert.True(t, user2ID > 0)

				user3ID := utils.Must(userModel.CreateFromJSON(`{"name": "User 3", "username": "user3"}`))
				assert.True(t, user3ID > 0)

				user1 := utils.Must(userModel.Query(app.EQ("id", user1ID)).First())
				user2 := utils.Must(userModel.Query(app.EQ("id", user2ID)).First())

				return []*schema.Entity{
					schema.NewEntity(group1ID).Set("name", "Group 1").Set("users", []*schema.Entity{
						user1,
						user2,
					}),
					schema.NewEntity(group2ID).Set("name", "Group 2").Set("users", []*schema.Entity{
						user2,
					}),
					schema.NewEntity(group3ID).Set("name", "Group 3").Set("users", []*schema.Entity{}),
				}
			},
		},
		{
			Name:        "Query_with_edges_M2M_two_types_reverse",
			Schema:      "user",
			Columns:     []string{"name", "groups"},
			ClearTables: []string{"groups", "users", "groups_users"},
			Prepare: func(t *testing.T, client app.DBClient, m app.Model) []*schema.Entity {
				groupModel := utils.Must(client.Model("group"))
				user1ID := utils.Must(m.CreateFromJSON(`{"name": "User 1", "username": "user1"}`))
				assert.True(t, user1ID > 0)
				user2ID := utils.Must(m.CreateFromJSON(`{"name": "User 2", "username": "user2"}`))
				assert.True(t, user2ID > 0)

				user3ID := utils.Must(m.CreateFromJSON(`{"name": "User 3", "username": "user3"}`))
				assert.True(t, user3ID > 0)

				group1ID := utils.Must(groupModel.CreateFromJSON(`{"name": "Group 1", "users": [{"id": 1}]}`))
				assert.True(t, group1ID > 0)

				group2ID := utils.Must(groupModel.CreateFromJSON(`{"name": "Group 2", "users": [{"id": 1}, {"id": 2}]}`))
				assert.True(t, group2ID > 0)

				group3ID := utils.Must(groupModel.CreateFromJSON(`{"name": "Group 3"}`))
				assert.True(t, group3ID > 0)

				group1 := utils.Must(groupModel.Query(app.EQ("id", group1ID)).First())
				group2 := utils.Must(groupModel.Query(app.EQ("id", group2ID)).First())

				return []*schema.Entity{
					schema.NewEntity(user1ID).Set("name", "User 1").Set("groups", []*schema.Entity{
						group1,
						group2,
					}),
					schema.NewEntity(user2ID).Set("name", "User 2").Set("groups", []*schema.Entity{
						group2,
					}),
					schema.NewEntity(user3ID).Set("name", "User 3").Set("groups", []*schema.Entity{}),
				}
			},
		},
		{
			Name:        "Query_with_edges_M2M_same_type",
			Schema:      "user",
			Columns:     []string{"name", "following"},
			ClearTables: []string{"users", "followers_following"},
			Order:       []string{"id"},
			Prepare: func(t *testing.T, client app.DBClient, m app.Model) []*schema.Entity {
				user1ID := utils.Must(m.CreateFromJSON(`{"name": "User 1", "username": "user1"}`))
				assert.True(t, user1ID > 0)

				user2ID := utils.Must(m.CreateFromJSON(`{"name": "User 2", "username": "user2"}`))
				assert.True(t, user2ID > 0)

				user3ID := utils.Must(m.CreateFromJSON(`{"name": "User 3", "username": "user3"}`))
				assert.True(t, user3ID > 0)

				user4ID := utils.Must(m.CreateFromJSON(`{"name": "User 4", "username": "user4"}`))
				assert.True(t, user4ID > 0)

				user5ID := utils.Must(m.CreateFromJSON(`{"name": "User 5", "username": "user5"}`))
				assert.True(t, user5ID > 0)

				_, err := utils.Must(m.Mutation()).Where(app.EQ("id", user1ID)).Update(schema.NewEntity(user1ID).Set("following", []*schema.Entity{
					schema.NewEntity(user2ID),
					schema.NewEntity(user3ID),
				}))
				assert.NoError(t, err)

				_, err = utils.Must(m.Mutation()).Where(app.EQ("id", user2ID)).Update(schema.NewEntity(user2ID).Set("following", []*schema.Entity{
					schema.NewEntity(user3ID),
					schema.NewEntity(user4ID),
				}))
				assert.NoError(t, err)

				_, err = utils.Must(m.Mutation()).Where(app.EQ("id", user3ID)).Update(schema.NewEntity(user3ID).Set("following", []*schema.Entity{
					schema.NewEntity(user4ID),
				}))
				assert.NoError(t, err)

				user2 := utils.Must(m.Query(app.EQ("id", user2ID)).First())
				user3 := utils.Must(m.Query(app.EQ("id", user3ID)).First())
				user4 := utils.Must(m.Query(app.EQ("id", user4ID)).First())

				return []*schema.Entity{
					schema.NewEntity(user1ID).Set("name", "User 1").Set("following", []*schema.Entity{
						user2,
						user3,
					}),
					schema.NewEntity(user2ID).Set("name", "User 2").Set("following", []*schema.Entity{
						user3,
						user4,
					}),
					schema.NewEntity(user3ID).Set("name", "User 3").Set("following", []*schema.Entity{
						user4,
					}),
					schema.NewEntity(user4ID).Set("name", "User 4").Set("following", []*schema.Entity{}),
					schema.NewEntity(user5ID).Set("name", "User 5").Set("following", []*schema.Entity{}),
				}
			},
		},
		{
			Name:        "Query_with_edges_M2M_same_type_reverse",
			Schema:      "user",
			Columns:     []string{"name", "followers"},
			ClearTables: []string{"users", "followers_following"},
			Order:       []string{"id"},
			Prepare: func(t *testing.T, client app.DBClient, m app.Model) []*schema.Entity {
				user1ID := utils.Must(m.CreateFromJSON(`{"name": "User 1", "username": "user1"}`))
				assert.True(t, user1ID > 0)

				user2ID := utils.Must(m.CreateFromJSON(`{"name": "User 2", "username": "user2", "following": [{"id": 1}]}`))
				assert.True(t, user2ID > 0)

				user3ID := utils.Must(m.CreateFromJSON(`{"name": "User 3", "username": "user3", "following": [{"id": 1}, {"id": 2}]}`))
				assert.True(t, user3ID > 0)

				user4ID := utils.Must(m.CreateFromJSON(`{"name": "User 4", "username": "user4", "following": [{"id": 1}, {"id": 2}, {"id": 3}]}`))
				assert.True(t, user4ID > 0)

				user5ID := utils.Must(m.CreateFromJSON(`{"name": "User 5", "username": "user5"}`))
				assert.True(t, user5ID > 0)

				user2 := utils.Must(m.Query(app.EQ("id", user2ID)).First())
				user3 := utils.Must(m.Query(app.EQ("id", user3ID)).First())
				user4 := utils.Must(m.Query(app.EQ("id", user4ID)).First())

				return []*schema.Entity{
					schema.NewEntity(user1ID).Set("name", "User 1").Set("followers", []*schema.Entity{
						user2,
						user3,
						user4,
					}),
					schema.NewEntity(user2ID).Set("name", "User 2").Set("followers", []*schema.Entity{
						user3,
						user4,
					}),
					schema.NewEntity(user3ID).Set("name", "User 3").Set("followers", []*schema.Entity{
						user4,
					}),
					schema.NewEntity(user4ID).Set("name", "User 4").Set("followers", []*schema.Entity{}),
					schema.NewEntity(user5ID).Set("name", "User 5").Set("followers", []*schema.Entity{}),
				}
			},
		},
		{
			Name:        "Query_with_edges_M2M_bidi",
			Schema:      "user",
			Columns:     []string{"name", "friends"},
			ClearTables: []string{"users", "friends_user"},
			Order:       []string{"id"},
			Prepare: func(t *testing.T, client app.DBClient, m app.Model) []*schema.Entity {
				user1ID := utils.Must(m.CreateFromJSON(`{"name": "User 1", "username": "user1"}`))
				assert.True(t, user1ID > 0)

				user2ID := utils.Must(m.CreateFromJSON(`{"name": "User 2", "username": "user2", "friends": [{"id": 1}]}`))
				assert.True(t, user2ID > 0)

				user3ID := utils.Must(m.CreateFromJSON(`{"name": "User 3", "username": "user3", "friends": [{"id": 1}, {"id": 2}]}`))
				assert.True(t, user3ID > 0)

				user4ID := utils.Must(m.CreateFromJSON(`{"name": "User 4", "username": "user4", "friends": [{"id": 1}]}`))
				assert.True(t, user4ID > 0)

				user5ID := utils.Must(m.CreateFromJSON(`{"name": "User 5", "username": "user5", "friends": [{"id": 4}]}`))
				assert.True(t, user5ID > 0)

				user6ID := utils.Must(m.CreateFromJSON(`{"name": "User 6", "username": "user6"}`))
				assert.True(t, user6ID > 0)

				user1 := utils.Must(m.Query(app.EQ("id", user1ID)).First())
				user2 := utils.Must(m.Query(app.EQ("id", user2ID)).First())
				user3 := utils.Must(m.Query(app.EQ("id", user3ID)).First())
				user4 := utils.Must(m.Query(app.EQ("id", user4ID)).First())
				user5 := utils.Must(m.Query(app.EQ("id", user5ID)).First())

				return []*schema.Entity{
					schema.NewEntity(user1ID).Set("name", "User 1").Set("friends", []*schema.Entity{
						user2,
						user3,
						user4,
					}),
					schema.NewEntity(user2ID).Set("name", "User 2").Set("friends", []*schema.Entity{
						user1,
						user3,
					}),
					schema.NewEntity(user3ID).Set("name", "User 3").Set("friends", []*schema.Entity{
						user1,
						user2,
					}),
					schema.NewEntity(user4ID).Set("name", "User 4").Set("friends", []*schema.Entity{
						user1,
						user5,
					}),
					schema.NewEntity(user5ID).Set("name", "User 5").Set("friends", []*schema.Entity{
						user4,
					}),
					schema.NewEntity(user6ID).Set("name", "User 6").Set("friends", []*schema.Entity{}),
				}
			},
		},
		{
			Name:    "Query_with_edges_O2O_fields",
			Schema:  "user",
			Columns: []string{"name", "card.number", "sub_card.number"},
			Filter: `{
				"id": {
					"$in": [1, 2]
				}
			}`,
			ClearTables: []string{"users", "cards"},
			Prepare: func(t *testing.T, client app.DBClient, m app.Model) []*schema.Entity {
				user1ID := utils.Must(m.CreateFromJSON(`{"name": "User 1", "username": "user1"}`))
				user2ID := utils.Must(m.CreateFromJSON(`{"name": "User 2", "username": "user2"}`))
				assert.True(t, user1ID > 0)
				card1ID := utils.Must(utils.Must(client.Model("card")).CreateFromJSON(`{"number": "1234", "owner": {"id": 1}, "sub_owner": {"id": 2}}`))
				assert.True(t, card1ID > 0)
				card1 := schema.NewEntity(card1ID).Set("number", "1234").Set("owner_id", uint64(1))
				card1Sub := schema.NewEntity(card1ID).Set("number", "1234").Set("sub_owner_id", uint64(2))
				return []*schema.Entity{
					schema.NewEntity(user1ID).Set("name", "User 1").Set("card", card1),
					schema.NewEntity(user2ID).Set("name", "User 2").Set("sub_card", card1Sub),
				}
			},
		},
		{
			Name:    "Query_with_edges_O2O_reverse_fields",
			Schema:  "card",
			Columns: []string{"number", "owner.name", "owner.age", "sub_owner.name", "sub_owner.age"},
			Filter: `{
				"id": {
					"$in": [1, 2]
				}
			}`,
			ClearTables: []string{"users", "cards"},
			Prepare: func(t *testing.T, client app.DBClient, m app.Model) []*schema.Entity {
				user1ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(`{"name": "User 1", "username": "user1", "age": 5}`))
				user2ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(`{"name": "User 2", "username": "user2", "age": 8}`))
				assert.True(t, user1ID > 0)

				card1ID := utils.Must(m.CreateFromJSON(`{"number": "1234", "owner": {"id": 1}, "sub_owner": {"id": 2}}`))
				assert.True(t, card1ID > 0)
				card2ID := utils.Must(m.CreateFromJSON(`{"number": "5678", "owner": {"id": 2}}`))

				user1 := schema.NewEntity(user1ID).Set("name", "User 1").Set("age", uint(5))
				user2 := schema.NewEntity(user2ID).Set("name", "User 2").Set("age", uint(8))

				return []*schema.Entity{
					schema.NewEntity(card1ID).
						Set("number", "1234").
						Set("owner_id", uint64(1)).
						Set("sub_owner_id", uint64(2)).
						Set("owner", user1).
						Set("sub_owner_id", uint64(2)).
						Set("sub_owner", user2),
					schema.NewEntity(card2ID).Set("number", "5678").Set("owner_id", uint64(2)).Set("owner", user2),
				}
			},
		},
		{
			Name:   "Query_with_edges_O2M_fields",
			Schema: "user",
			Columns: []string{
				"name",
				"pets.name",
				"pets.created_at",
			},
			ClearTables: []string{"users", "pets"},
			Order:       []string{"id"},
			Prepare: func(t *testing.T, client app.DBClient, m app.Model) []*schema.Entity {
				user1ID := utils.Must(m.CreateFromJSON(`{"name": "User 1", "username": "user1"}`))
				assert.True(t, user1ID > 0)
				pet1ID := utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(`{"name": "Pet 1", "owner": {"id": 1}}`))
				assert.True(t, pet1ID > 0)
				pet2ID := utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(`{"name": "Pet 2", "owner": {"id": 1}}`))
				assert.True(t, pet2ID > 0)

				pet1 := utils.Must(utils.Must(client.Model("pet")).Query(app.EQ("id", pet1ID)).First())
				pet2 := utils.Must(utils.Must(client.Model("pet")).Query(app.EQ("id", pet2ID)).First())

				return []*schema.Entity{schema.NewEntity(user1ID).
					Set("name", "User 1").
					Set("pets", []*schema.Entity{
						schema.NewEntity(pet1ID).
							Set("name", "Pet 1").
							Set("created_at", pet1.Get("created_at")).
							Set("owner_id", uint64(1)),
						schema.NewEntity(pet2ID).
							Set("name", "Pet 2").
							Set("created_at", pet2.Get("created_at")).
							Set("owner_id", uint64(1)),
					}),
				}
			},
		},
		{
			Name:   "Query_with_edges_O2M_reverse_fields",
			Schema: "pet",
			Columns: []string{
				"id",
				"name",
				"owner.name",
				"owner.age",
			},
			ClearTables: []string{"users", "pets"},
			Prepare: func(t *testing.T, client app.DBClient, m app.Model) []*schema.Entity {
				user1ID := utils.Must(utils.Must(client.Model("user")).CreateFromJSON(`{"name": "User 1", "username": "user1", "age": 8}`))
				assert.True(t, user1ID > 0)
				pet1ID := utils.Must(m.CreateFromJSON(`{"name": "Pet 1", "owner": {"id": 1}}`))
				assert.True(t, pet1ID > 0)
				pet2ID := utils.Must(m.CreateFromJSON(`{"name": "Pet 2", "owner": {"id": 1}}`))
				assert.True(t, pet2ID > 0)
				user1 := schema.NewEntity(user1ID).
					Set("name", "User 1").
					Set("age", uint(8))
				return []*schema.Entity{
					schema.NewEntity(pet1ID).
						Set("name", "Pet 1").
						Set("owner_id", uint64(1)).
						Set("owner", user1),
					schema.NewEntity(pet2ID).
						Set("name", "Pet 2").
						Set("owner_id", uint64(1)).
						Set("owner", user1),
				}
			},
		},
		{
			Name:        "Query_with_edges_M2M_fields",
			Schema:      "group",
			Columns:     []string{"name", "users.name", "users.age"},
			ClearTables: []string{"groups", "users", "groups_users"},
			Order:       []string{"id"},
			Prepare: func(t *testing.T, client app.DBClient, m app.Model) []*schema.Entity {
				userModel := utils.Must(client.Model("user"))
				group1ID := utils.Must(m.CreateFromJSON(`{"name": "Group 1"}`))
				assert.True(t, group1ID > 0)
				group2ID := utils.Must(m.CreateFromJSON(`{"name": "Group 2"}`))
				assert.True(t, group2ID > 0)

				group3ID := utils.Must(m.CreateFromJSON(`{"name": "Group 3"}`))
				assert.True(t, group3ID > 0)

				user1ID := utils.Must(userModel.CreateFromJSON(`{"name": "User 1", "username": "user1", "age": 8, "groups": [{"id": 1}]}`))
				assert.True(t, user1ID > 0)
				user2ID := utils.Must(userModel.CreateFromJSON(`{"name": "User 2", "username": "user2", "age": 5, "groups": [{"id": 1}, {"id": 2}]}`))
				assert.True(t, user2ID > 0)

				user3ID := utils.Must(userModel.CreateFromJSON(`{"name": "User 3", "username": "user3"}`))
				assert.True(t, user3ID > 0)

				return []*schema.Entity{
					schema.NewEntity(group1ID).
						Set("name", "Group 1").
						Set("users", []*schema.Entity{
							schema.NewEntity(user1ID).
								Set("name", "User 1").
								Set("age", uint(8)),
							schema.NewEntity(user2ID).
								Set("name", "User 2").
								Set("age", uint(5)),
						}),
					schema.NewEntity(group2ID).
						Set("name", "Group 2").
						Set("users", []*schema.Entity{
							schema.NewEntity(user2ID).
								Set("name", "User 2").
								Set("age", uint(5)),
						}),
					schema.NewEntity(group3ID).
						Set("name", "Group 3").
						Set("users", []*schema.Entity{}),
				}
			},
		},
	}

	DBRunQueryTests(client, t, tests)
}

func DBCountNode(t *testing.T, client app.DBClient) {
	tests := []DBTestCountData{
		{
			Name:        "Count_with_no_filter",
			Schema:      "user",
			ClearTables: []string{"users"},
			Prepare: func(t *testing.T, client app.DBClient, m app.Model) int {
				user1ID := utils.Must(m.CreateFromJSON(`{"name": "User 1", "username": "user1"}`))
				assert.True(t, user1ID > 0)

				user2ID := utils.Must(m.CreateFromJSON(`{"name": "User 2", "username": "user2"}`))
				assert.True(t, user2ID > 0)

				return 2
			},
		},
		{
			Name:        "Count_with_filter",
			Schema:      "user",
			ClearTables: []string{"users"},
			Filter: `{
				"id": {
					"$gt": 1
				}
			}`,
			Prepare: func(t *testing.T, client app.DBClient, m app.Model) int {
				user1ID := utils.Must(m.CreateFromJSON(`{"name": "User 1", "username": "user1"}`))
				assert.True(t, user1ID > 0)

				user2ID := utils.Must(m.CreateFromJSON(`{"name": "User 2", "username": "user2"}`))
				assert.True(t, user2ID > 0)

				return 1
			},
		},
		{
			Name:   "Count_with_columns",
			Schema: "user",
			Filter: `{
					"id": {
						"$gt": 1
					}
				}`,
			Column:      "status",
			ClearTables: []string{"users"},
			Prepare: func(t *testing.T, client app.DBClient, m app.Model) int {
				user1ID := utils.Must(m.CreateFromJSON(`{"name": "User 1", "username": "user1", "status": "offline"}`))
				assert.True(t, user1ID > 0)

				user2ID := utils.Must(m.CreateFromJSON(`{"name": "User 2", "username": "user2", "status": "online"}`))
				assert.True(t, user2ID > 0)

				user3ID := utils.Must(m.CreateFromJSON(`{"name": "User 3", "username": "user3", "status": "offline"}`))
				assert.True(t, user3ID > 0)

				user4ID := utils.Must(m.CreateFromJSON(`{"name": "User 4", "username": "user4", "status": "online"}`))
				assert.True(t, user4ID > 0)

				return 3
			},
		},
		{
			Name:   "Count_with_column_and_unique",
			Schema: "user",
			Filter: `{
					"id": {
						"$gt": 1
					}
				}`,
			Unique:      true,
			Column:      "status",
			ClearTables: []string{"users"},
			Prepare: func(t *testing.T, client app.DBClient, m app.Model) int {
				user1ID := utils.Must(m.CreateFromJSON(`{"name": "User 1", "username": "user1", "status": "offline"}`))
				assert.True(t, user1ID > 0)

				user2ID := utils.Must(m.CreateFromJSON(`{"name": "User 2", "username": "user2", "status": "online"}`))
				assert.True(t, user2ID > 0)

				user3ID := utils.Must(m.CreateFromJSON(`{"name": "User 3", "username": "user3", "status": "offline"}`))
				assert.True(t, user3ID > 0)

				user4ID := utils.Must(m.CreateFromJSON(`{"name": "User 4", "username": "user4", "status": "online"}`))
				assert.True(t, user4ID > 0)

				return 2
			},
		},
	}

	DBRunCountTests(client, t, tests)
}
