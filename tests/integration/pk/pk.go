package pk_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	u "github.com/fastschema/fastschema/pkg/utils"
	h "github.com/fastschema/fastschema/tests/integration/helpers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type userSchemaStub struct{}

type systemSeries struct {
	_        any              `json:"-" fs:"namespace=pk_system_series;label_field=title"`
	ID       string           `json:"id" fs:"type=uuid;filterable;sortable"`
	Title    string           `json:"title"`
	Synopsis string           `json:"synopsis" fs:"optional"`
	Episodes []*systemEpisode `json:"episodes" fs.relation:"{'type':'o2m','schema':'system_episode','field':'series','owner':true}"`
	Topics   []*systemTopic   `json:"topics,omitempty" fs.relation:"{'type':'m2m','schema':'system_topic','field':'series','owner':true}"`
}

type systemEpisode struct {
	_        any               `json:"-" fs:"namespace=pk_system_episodes;label_field=title"`
	ID       uint64            `json:"id"`
	Title    string            `json:"title"`
	Duration int               `json:"duration"`
	Series   *systemSeries     `json:"series" fs.relation:"{'type':'o2m','schema':'system_series','field':'episodes'}"`
	Viewers  []*userSchemaStub `json:"viewers,omitempty" fs.relation:"{'type':'m2m','schema':'user','field':'favorite_episodes','owner':true}"`
}

type systemTopic struct {
	_      any             `json:"-" fs:"namespace=pk_system_topics;label_field=label"`
	ID     string          `json:"id" fs:"filterable;sortable"`
	Label  string          `json:"label"`
	Series []*systemSeries `json:"series,omitempty" fs.relation:"{'type':'m2m','schema':'system_series','field':'topics'}"`
}

type systemProfile struct {
	_           any             `json:"-" fs:"namespace=pk_system_profiles;label_field=display_name"`
	ID          string          `json:"id" fs:"filterable;sortable"`
	DisplayName string          `json:"display_name"`
	Bio         string          `json:"bio" fs:"optional"`
	Owner       *userSchemaStub `json:"owner" fs:"optional" fs.relation:"{'type':'o2o','schema':'user','field':'profile'}"`
}

var pkTables = []string{
	"liked_posts_likes",
	"followed_categories_followers",
	"posts_tags",
	"series_topics",
	"favorite_episodes_viewers",
	"pk_system_episodes",
	"pk_system_series",
	"pk_system_topics",
	"pk_system_profiles",
	"pk_comments",
	"pk_tags",
	"pk_posts",
	"pk_categories",
	"pk_users",
}

type pkGraph struct {
	categorySlug string
	userID       string
	postID       uint64
	tagID        uint64
	commentID    uint64
}

func seedBlogGraph(t *testing.T, client h.DBClient) *pkGraph {
	h.ClearDBData(client.C, pkTables...)

	categoryModel := u.Must(client.C.Model("category"))
	userModel := u.Must(client.C.Model("user"))
	postModel := u.Must(client.C.Model("post"))
	tagModel := u.Must(client.C.Model("tag"))
	commentModel := u.Must(client.C.Model("comment"))

	slug := fmt.Sprintf(
		"cat-%s",
		strings.ToLower(u.Must(uuid.NewV7()).String())[:8],
	)
	u.Must(categoryModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"id":"%s","title":"Tech","description":"Latest"}`, slug),
	))

	userID := u.Must(uuid.NewV7()).String()
	u.Must(userModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"id":"%s","name":"Alice","email":"alice@example.com"}`, userID),
	))

	postID := normalizeUint(t, u.Must(postModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"title":"PK 101","body":"content","category":{"id":"%s"},"author":{"id":"%s"}}`, slug, userID),
	)))

	tagID := uint64(9001)
	u.Must(tagModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"id":%d,"name":"go"}`, tagID),
	))
	u.Must(tagModel.
		Mutation().
		Where(db.EQ("id", tagID)).
		Update(h.Ctx(), entity.New().Set("posts", []*entity.Entity{entity.New(postID)})))

	commentID := normalizeUint(t, u.Must(commentModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"content":"first!","post":{"id":%d},"author":{"id":"%s"}}`, postID, userID),
	)))

	return &pkGraph{
		categorySlug: slug,
		userID:       userID,
		postID:       postID,
		tagID:        tagID,
		commentID:    commentID,
	}
}

func normalizeUint(t *testing.T, value any) uint64 {
	t.Helper()
	normalized, err := u.AnyToUint[uint64](value)
	require.NoError(t, err)
	return normalized
}

func coerceString(t *testing.T, value any) string {
	t.Helper()
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case uuid.UUID:
		return v.String()
	default:
		t.Fatalf("expected string-compatible column, got %T", value)
		return ""
	}
}

func coerceUint(t *testing.T, value any) uint64 {
	t.Helper()
	converted, err := u.AnyToUint[uint64](value)
	require.NoError(t, err)
	return converted
}

func coerceInt(t *testing.T, value any) int64 {
	t.Helper()
	converted, err := u.AnyToInt[int64](value)
	require.NoError(t, err)
	return converted
}
