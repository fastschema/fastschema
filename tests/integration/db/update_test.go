package db

import (
	"fmt"
	"testing"

	"entgo.io/ent/dialect"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
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
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 2", "username": "user2", "provider": "local" }`))
			},
			Expect: func(t *testing.T, m db.Model) {
				entities, err := m.Query().Get(Ctx())
				assert.NoError(t, err)
				assert.NotNil(t, entities)
				assert.Equal(t, uint64(1), entities[0].ID())
				assert.Equal(t, uint(20), entities[0].Get("age"))
				assert.Equal(t, uint64(2), entities[1].ID())
				assert.Equal(t, uint(20), entities[1].Get("age"))
			},
		},
		{
			Name:         "predicates",
			Schema:       "user",
			InputJSON:    `{ "age": 20 }`,
			WantAffected: 2,
			ClearTables:  []string{"users"},
			Predicates:   []*db.Predicate{db.GT("id", 1)},
			Prepare: func(t *testing.T, m db.Model) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 2", "username": "user2", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 3", "username": "user3", "provider": "local" }`))
			},
			Expect: func(t *testing.T, m db.Model) {
				entities, err := m.Query(db.GT("id", 1)).Get(Ctx())
				assert.NoError(t, err)
				assert.NotNil(t, entities)
				assert.Equal(t, uint64(2), entities[0].ID())
				assert.Equal(t, uint(20), entities[0].Get("age"))
				assert.Equal(t, uint64(3), entities[1].ID())
				assert.Equal(t, uint(20), entities[1].Get("age"))
			},
		},
		{
			Name:        "fields/set_modifier/expr",
			Schema:      "user",
			ClearTables: []string{"users"},
			InputJSON: fmt.Sprintf(`{
				"name": "User 1 name",
				"username": "user1",
				"provider": "local",
				"$expr": {
					"bio": "LOWER(%s)"
				}
			}`, utils.If(client.Dialect() == dialect.Postgres, "bio", "`bio`")),
			WantAffected: 1,
			Prepare: func(t *testing.T, m db.Model) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local", "bio": "My BIO" }`))
			},
			Predicates: []*db.Predicate{db.EQ("id", 1)},
			Expect: func(t *testing.T, m db.Model) {
				entity := utils.Must(m.Query(db.EQ("id", 1)).Only(Ctx()))
				assert.NotNil(t, entity)
				assert.Equal(t, "my bio", entity.Get("bio"))
			},
		},
		{
			Name:   "fields/add",
			Schema: "user",
			InputJSON: `{
				"$add": {
					"age": 3
				}
			}`,
			ClearTables:  []string{"users"},
			Predicates:   []*db.Predicate{db.EQ("id", 1)},
			WantAffected: 1,
			Prepare: func(t *testing.T, m db.Model) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local", "age": "5" }`))
			},
			Expect: func(t *testing.T, m db.Model) {
				entity := utils.Must(m.Query(db.EQ("id", 1)).Only(Ctx()))
				assert.NotNil(t, entity)
				assert.Equal(t, uint(8), entity.Get("age"))
			},
		},
		{
			Name:   "fields/add_o2m_m2m",
			Schema: "user",
			InputJSON: `{
				"$add": {
					"sub_pets": [ { "id": 2 }, { "id": 3 } ],
					"sub_groups": [ { "id": 4 }, { "id": 5 } ]
				}
			}`,
			ClearTables:  []string{"users", "groups", "pets", "sub_groups_sub_users", "groups_users"},
			Predicates:   []*db.Predicate{db.EQ("id", 1)},
			WantAffected: 1,
			Prepare: func(t *testing.T, m db.Model) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(Ctx(), `{ "name": "Group 1" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(Ctx(), `{ "name": "Group 2" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(Ctx(), `{ "name": "Group 3" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(Ctx(), `{ "name": "Group 4" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(Ctx(), `{ "name": "Group 5" }`))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(Ctx(), `{
					"name": "Pet 1",
					"owner": {
						"id": 1
					}
				}`))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(Ctx(), `{
					"name": "Pet 2",
					"owner": {
						"id": 1
					}
				}`))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(Ctx(), `{
					"name": "Pet 3",
					"owner": {
						"id": 1
					}
				}`))
			},
			Expect: func(t *testing.T, m db.Model) {
				pet2 := utils.Must(utils.Must(client.Model("pet")).Query(db.EQ("id", 2)).Only(Ctx()))
				pet3 := utils.Must(utils.Must(client.Model("pet")).Query(db.EQ("id", 3)).Only(Ctx()))
				assert.Equal(t, uint64(1), pet2.Get("sub_owner_id"))
				assert.Equal(t, uint64(1), pet3.Get("sub_owner_id"))

				subGroupsUsers := utils.Must(utils.Must(client.Model("sub_groups_sub_users")).Query(db.EQ("sub_users", 1)).Get(Ctx()))
				subGroupsIDs := utils.Map(subGroupsUsers, func(e *entity.Entity) uint64 {
					return e.Get("sub_groups").(uint64)
				})
				assert.Equal(t, []uint64{4, 5}, subGroupsIDs)
			},
		},
		{
			Name:   "fields/clear",
			Schema: "user",
			InputJSON: `{
				"deleted": true,
				"$clear": {
					"bio": true
				}
			}`,
			ClearTables:  []string{"users"},
			Predicates:   []*db.Predicate{db.EQ("id", 1)},
			WantAffected: 1,
			Prepare: func(t *testing.T, m db.Model) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local", "bio": "My BIO" }`))
			},
			Expect: func(t *testing.T, m db.Model) {
				user := utils.Must(m.Query(db.EQ("id", 1)).Only(Ctx()))
				assert.NotNil(t, user)
				assert.Equal(t, true, user.Get("deleted"))
				assert.Equal(t, nil, user.Get("bio"))
			},
		},
		{
			Name:         "fields/clear/o2o_o2m_m2m_all",
			Schema:       "user",
			Predicates:   []*db.Predicate{db.EQ("id", 2)},
			WantAffected: 1,
			ClearTables:  []string{"users", "groups", "pets", "cars"},
			Prepare: func(t *testing.T, m db.Model) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("car")).CreateFromJSON(Ctx(), `{ "name": "Car 1" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(Ctx(), `{ "name": "Group 1" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(Ctx(), `{ "name": "Group 2" }`))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(Ctx(), `{
					"name": "Pet 1",
					"owner": {
						"id": 1
					}
				}`))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(Ctx(), `{
					"name": "Pet 2",
					"owner": {
						"id": 1
					}
				}`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{
					"name": "User 2",
					"username": "user2",
					"provider": "local",
					"bio": "My BIO",
					"car": {
						"id": 1
					},
					"sub_pets": [ { "id": 1 }, { "id": 2 } ],
					"sub_groups": [ { "id": 1 }, { "id": 2 } ]
				}`))
			},
			InputJSON: `{
				"$clear": {
					"bio": true,
					"car": true,
					"sub_pets": true,
					"sub_groups": true
				}
			}`,
			Expect: func(t *testing.T, m db.Model) {
				user := utils.Must(m.Query(db.EQ("id", 2)).Only(Ctx()))
				assert.NotNil(t, user)
				assert.Equal(t, nil, user.Get("bio"))
				assert.Equal(t, nil, user.Get("car_id"))

				subGroupsUsers := utils.Must(utils.Must(client.Model("sub_groups_sub_users")).Query(db.EQ("sub_users", 2)).Get(Ctx()))
				assert.Equal(t, 0, len(subGroupsUsers))

				subPets := utils.Must(utils.Must(client.Model("pet")).Query(db.EQ("owner_id", 2)).Get(Ctx()))
				assert.Equal(t, 0, len(subPets))
			},
		},
		{
			Name:         "fields/clear/o2o_o2m_m2m_part",
			Schema:       "user",
			Predicates:   []*db.Predicate{db.EQ("id", 2)},
			WantAffected: 1,
			ClearTables:  []string{"users", "groups", "pets", "cars"},
			Prepare: func(t *testing.T, m db.Model) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("car")).CreateFromJSON(Ctx(), `{ "name": "Car 1" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(Ctx(), `{ "name": "Group 1" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(Ctx(), `{ "name": "Group 2" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(Ctx(), `{ "name": "Group 3" }`))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(Ctx(), `{
					"name": "Pet 1",
					"owner": {
						"id": 1
					}
				}`))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(Ctx(), `{
					"name": "Pet 2",
					"owner": {
						"id": 1
					}
				}`))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(Ctx(), `{
					"name": "Pet 3",
					"owner": {
						"id": 1
					}
				}`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{
					"name": "User 2",
					"username": "user2",
					"provider": "local",
					"sub_pets": [ { "id": 1 }, { "id": 2 }, { "id": 3 } ],
					"sub_groups": [ { "id": 1 }, { "id": 2 }, { "id": 3 } ]
				}`))
			},
			InputJSON: `{
				"$clear": {
					"bio": true,
					"car": true,
					"sub_pets": [ { "id": 1 }, { "id": 2 } ],
					"sub_groups": [ { "id": 1 }, { "id": 2 }]
				}
			}`,
			Expect: func(t *testing.T, m db.Model) {
				user := utils.Must(m.Query(db.EQ("id", 2)).Only(Ctx()))
				assert.NotNil(t, user)

				subGroupsUsers := utils.Must(utils.Must(client.Model("sub_groups_sub_users")).Query(db.EQ("sub_users", 2)).Get(Ctx()))
				subGroupsUsersIds := utils.Map(subGroupsUsers, func(e *entity.Entity) uint64 {
					return e.Get("sub_groups").(uint64)
				})
				assert.Equal(t, 1, len(subGroupsUsersIds))
				assert.Equal(t, []uint64{3}, subGroupsUsersIds)

				subPets := utils.Must(utils.Must(client.Model("pet")).Query(db.EQ("sub_owner_id", 2)).Get(Ctx()))
				subPetsIds := utils.Map(subPets, func(e *entity.Entity) uint64 {
					return e.Get("id").(uint64)
				})
				assert.Equal(t, 1, len(subPetsIds))
				assert.Equal(t, []uint64{3}, subPetsIds)
			},
		},
		{
			Name:   "fields/set/block",
			Schema: "user",
			InputJSON: `{
				"name": "User 3 updated",
				"username": "user3",
				"provider": "local",
				"$set": {
					"bio": "Hello World",
					"sub_card": { "id": 2 },
					"sub_pets": [ { "id": 2 } ],
					"sub_groups": [ { "id": 2 } ]
				}
			}`,
			Predicates:   []*db.Predicate{db.EQ("id", 3)},
			WantAffected: 1,
			ClearTables:  []string{"users", "pets", "cards"},
			Prepare: func(t *testing.T, m db.Model) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 2", "username": "user2", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("card")).CreateFromJSON(Ctx(), `{
					"number": "00001",
					"owner": {
						"id": 1
					}
				}`))
				utils.Must(utils.Must(client.Model("card")).CreateFromJSON(Ctx(), `{
					"number": "00002",
					"owner": {
						"id": 2
					}
				}`))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(Ctx(), `{
					"name": "Pet 1",
					"owner": {
						"id": 1
					}
				}`))
				utils.Must(utils.Must(client.Model("pet")).CreateFromJSON(Ctx(), `{
					"name": "Pet 2",
					"owner": {
						"id": 1
					}
				}`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(Ctx(), `{ "name": "Group 1" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(Ctx(), `{ "name": "Group 2" }`))

				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{
					"name": "User 3",
					"username": "user3",
					"provider": "local",
					"sub_card": { "id": 1 },
					"sub_pets": [ { "id": 1 } ],
					"sub_groups": [ { "id": 1 } ]
				}`))
			},
			Expect: func(t *testing.T, m db.Model) {
				user3 := utils.Must(m.Query(db.EQ("id", 3)).Only(Ctx()))
				assert.NotNil(t, user3)

				assert.Equal(t, "User 3 updated", user3.Get("name").(string))
				assert.Equal(t, "Hello World", user3.Get("bio").(string))

				subCards := utils.Must(utils.Must(client.Model("pet")).Query(db.EQ("sub_owner_id", 3)).Get(Ctx()))
				subCardsIds := utils.Map(subCards, func(e *entity.Entity) uint64 {
					return e.Get("id").(uint64)
				})
				assert.Equal(t, 1, len(subCardsIds))
				assert.Equal(t, []uint64{2}, subCardsIds)

				subPets := utils.Must(utils.Must(client.Model("pet")).Query(db.EQ("sub_owner_id", 3)).Get(Ctx()))
				subPetsIds := utils.Map(subPets, func(e *entity.Entity) uint64 {
					return e.Get("id").(uint64)
				})
				assert.Equal(t, 1, len(subPetsIds))
				assert.Equal(t, []uint64{2}, subPetsIds)

				subGroupsUsers := utils.Must(utils.Must(client.Model("sub_groups_sub_users")).Query(db.EQ("sub_users", 3)).Get(Ctx()))
				subGroupsUsersIds := utils.Map(subGroupsUsers, func(e *entity.Entity) uint64 {
					return e.Get("sub_groups").(uint64)
				})
				assert.Equal(t, 1, len(subGroupsUsersIds))
				assert.Equal(t, []uint64{2}, subGroupsUsersIds)
			},
		},
		{
			Name:   "edges/o2o_non_inverse_and_m2o",
			Schema: "user",
			InputJSON: `{
				"$clear": {
					"car": true,
					"workplace": true
				},
				"$add": {
					"room": { "id": 1 },
					"parent": { "id": 1 }
				}
			}`,
			Predicates:   []*db.Predicate{db.EQ("id", 2)},
			WantAffected: 1,
			ClearTables:  []string{"users", "cars", "workplaces", "rooms", "users"},
			Prepare: func(t *testing.T, m db.Model) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("car")).CreateFromJSON(Ctx(), `{ "name": "Car 1" }`))
				utils.Must(utils.Must(client.Model("workplace")).CreateFromJSON(Ctx(), `{ "name": "Workplace 1" }`))
				utils.Must(utils.Must(client.Model("room")).CreateFromJSON(Ctx(), `{ "name": "Room 1" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{
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
			},
			Expect: func(t *testing.T, m db.Model) {
				user2 := utils.Must(m.Query(db.EQ("id", 2)).Only(Ctx()))
				assert.NotNil(t, user2)

				assert.Equal(t, nil, user2.Get("car_id"))
				assert.Equal(t, nil, user2.Get("workplace_id"))
				assert.Equal(t, uint64(1), user2.Get("room_id"))
				assert.Equal(t, uint64(1), user2.Get("parent_id"))
			},
		},
		{
			Name:   "edges/o2o_bidi",
			Schema: "user",
			InputJSON: `{
				"$clear": {
					"partner": true,
					"spouse": { "id": 2 }
				},
				"$add": {
					"spouse": { "id": 3 }
				}
			}`,
			Predicates:   []*db.Predicate{db.EQ("id", 4)},
			WantAffected: 1,
			ClearTables:  []string{"users"},
			Prepare: func(t *testing.T, m db.Model) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 2", "username": "user2", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 3", "username": "user3", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{
					"name": "User 4",
					"username": "user4",
					"provider": "local",
					"partner": {
						"id": 1
					},
					"spouse": {
						"id": 2
					}
				}`))
			},
			Expect: func(t *testing.T, m db.Model) {
				user4 := utils.Must(m.Query(db.EQ("id", 4)).Only(Ctx()))
				assert.NotNil(t, user4)

				assert.Equal(t, nil, user4.Get("partner_id"))
				assert.Equal(t, uint64(3), user4.Get("spouse_id"))
			},
		},
		{
			Name:         "edges/clear_add_m2m",
			Schema:       "user",
			Predicates:   []*db.Predicate{db.EQ("id", 9)},
			WantAffected: 1,
			ClearTables:  []string{"users", "groups", "groups_users", "followers_following", "blockers_blocking", "friends_user"},
			Prepare: func(t *testing.T, m db.Model) {
				utils.Must(utils.Must(client.Model("comment")).CreateFromJSON(Ctx(), `{ "content": "Comment 1" }`))
				utils.Must(utils.Must(client.Model("comment")).CreateFromJSON(Ctx(), `{ "content": "Comment 2" }`))

				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(Ctx(), `{ "name": "Group 1" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(Ctx(), `{ "name": "Group 2" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(Ctx(), `{ "name": "Group 3" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(Ctx(), `{ "name": "Group 4" }`))
				utils.Must(utils.Must(client.Model("group")).CreateFromJSON(Ctx(), `{ "name": "Group 5" }`))

				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 2", "username": "user2", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 3", "username": "user3", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 4", "username": "user4", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 5", "username": "user5", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 6", "username": "user6", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 7", "username": "user7", "provider": "local" }`))
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 8", "username": "user8", "provider": "local" }`))

				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{
					"name": "User 9",
					"username": "user9",
					"provider": "local",
					"blocking": [{ "id": 1 }, { "id": 2 }],
					"following": [{ "id": 3 }, { "id": 4 }],
					"friends": [{ "id": 5 }, { "id": 6 }],
					"groups": [ { "id": 1 }, { "id": 2 }, { "id": 3 } ],
					"comments": [{ "id": 1 }, { "id": 2 }]
				}`))
			},
			InputJSON: `{
				"$clear": {
					"blocking": true,
					"following": [{ "id": 3 }],
					"friends": { "id": 5 },
					"groups": [ { "id": 1 }, { "id": 2 } ],
					"comments": true
				},
				"$add": {
					"friends": [ { "id": 7 }, { "id": 8 } ],
					"groups": [ { "id": 4 }, { "id": 5 } ]
				}
			}`,
			Expect: func(t *testing.T, m db.Model) {
				user9 := utils.Must(m.Query(db.EQ("id", 9)).Only(Ctx()))
				assert.NotNil(t, user9)

				blockingUsers := utils.Must(utils.Must(client.Model("blockers_blocking")).Query(db.EQ("blockers", 9)).Get(Ctx()))
				assert.Equal(t, 0, len(blockingUsers))

				followingUsers := utils.Must(utils.Must(client.Model("followers_following")).Query(db.EQ("followers", 9)).Get(Ctx()))
				assert.Equal(t, 1, len(followingUsers))
				assert.Equal(t, uint64(4), followingUsers[0].Get("following"))

				friends := utils.Must(utils.Must(client.Model("friends_user")).Query(db.EQ("user", 9)).Get(Ctx()))
				friendsIds := utils.Map(friends, func(e *entity.Entity) uint64 {
					return e.Get("friends").(uint64)
				})
				assert.Equal(t, []uint64{6, 7, 8}, friendsIds)

				subGroupsUsers := utils.Must(utils.Must(client.Model("groups_users")).Query(db.EQ("users", 9)).Get(Ctx()))
				subGroupsUsersIds := utils.Map(subGroupsUsers, func(e *entity.Entity) uint64 {
					return e.Get("groups").(uint64)
				})
				assert.Equal(t, []uint64{3, 4, 5}, subGroupsUsersIds)
			},
		},
		{
			Name:   "fields/add_set_clear",
			Schema: "user",
			InputJSON: `{
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
			}`,
			ClearTables:  []string{"users"},
			Predicates:   []*db.Predicate{db.EQ("id", 1)},
			WantAffected: 1,
			Prepare: func(t *testing.T, m db.Model) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local", "age": 10, "bio": "Bio 1" }`))
			},
			Expect: func(t *testing.T, m db.Model) {
				user1 := utils.Must(m.Query(db.EQ("id", 1)).Only(Ctx()))
				assert.NotNil(t, user1)

				assert.Equal(t, "User 1 updated", user1.Get("name"))
				assert.Equal(t, true, user1.Get("deleted"))
				assert.Equal(t, uint(11), user1.Get("age"))
				assert.Equal(t, nil, user1.Get("bio"))
			},
		},
		{
			Name:   "fields/ensure_exists",
			Schema: "user",
			InputJSON: `{
				"$add": {
					"age": 1
				},
				"$clear": {
					"bio": true
				},
				"deleted": true
			}`,
			Predicates: []*db.Predicate{
				db.EQ("id", 1),
				db.IsFalse("deleted"),
			},
			WantAffected: 1,
			ClearTables:  []string{"users"},
			Prepare: func(t *testing.T, m db.Model) {
				utils.Must(utils.Must(client.Model("user")).CreateFromJSON(Ctx(), `{ "name": "User 1", "username": "user1", "provider": "local", "age": 10, "bio": "Bio 1" }`))
			},
			Expect: func(t *testing.T, m db.Model) {
				user1 := utils.Must(m.Query(db.EQ("id", 1)).Only(Ctx()))
				assert.NotNil(t, user1)

				assert.Equal(t, true, user1.Get("deleted"))
				assert.Equal(t, uint(11), user1.Get("age"))
				assert.Equal(t, nil, user1.Get("bio"))
			},
		},
	}

	DBRunUpdateTests(client, t, tests)
}
