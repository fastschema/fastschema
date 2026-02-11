package fk

import (
	"fmt"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	u "github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	h "github.com/fastschema/fastschema/tests/integration/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	schemaDir    = "../../../tests/integration/fk/data/schemas"
	migrationDir = "../../../tests/integration/fk/data/migrations"
	sqliteDSN    = "../../../tests/integration/fk/data/fk_test.db"
)

func TestMySQL(t *testing.T) {
	runFKTests(t, u.Map(h.MysqlConfigs, func(cfg h.DBConfig) h.DBClient {
		sb := u.Must(schema.NewBuilderFromDir(schemaDir))
		return h.NewMySQLClient(t, cfg, sb, migrationDir)
	}))
}

func TestPostgres(t *testing.T) {
	runFKTests(t, u.Map(h.PostgresConfigs, func(cfg h.DBConfig) h.DBClient {
		sb := u.Must(schema.NewBuilderFromDir(schemaDir))
		return h.NewPostgresClient(t, cfg, sb, migrationDir)
	}))
}

func TestSQLite(t *testing.T) {
	sb := u.Must(schema.NewBuilderFromDir(schemaDir))
	client := h.NewSQLiteClient(t, "sqlite", sqliteDSN, migrationDir, sb)
	runFKTests(t, []h.DBClient{client})
}

func runFKTests(t *testing.T, clients []h.DBClient) {
	for _, client := range clients {
		t.Run(client.Name, func(t *testing.T) {
			t.Run("O2M", func(t *testing.T) {
				runO2MCustomFKTests(t, client)
			})

			t.Run("O2O", func(t *testing.T) {
				runO2OCustomFKTests(t, client)
			})

			t.Run("M2M", func(t *testing.T) {
				runM2MCustomFKTests(t, client)
			})
		})
	}
}

func runO2MCustomFKTests(t *testing.T, client h.DBClient) {
	t.Run("Create", func(t *testing.T) {
		f := prepareAuthorBook(t, client)

		// Verify first book
		book := u.Must(f.book.Query(db.EQ("id", f.firstBookID)).
			Select("title", "author", "author_legacy_id").
			First(h.Ctx()))

		// The first book's author_legacy_id should match the first author's legacy_id
		assert.Equal(t, f.firstLegacyID, book.Get("author_legacy_id"))

		// Verify first author entity
		firstAuthor, ok := book.Get("author").(*entity.Entity)
		require.True(t, ok)
		assert.Equal(t, f.firstLegacyID, firstAuthor.Get("legacy_id"))

		// Verify second book
		other := u.Must(f.book.Query(db.EQ("id", f.secondBookID)).
			Select("title", "author", "author_legacy_id").
			First(h.Ctx()))

		// The second book's author_legacy_id should match the second author's legacy_id
		assert.Equal(t, f.secondLegacyID, other.Get("author_legacy_id"))

		// Verify second author entity
		secondAuthor, ok := other.Get("author").(*entity.Entity)
		require.True(t, ok)
		assert.Equal(t, f.secondLegacyID, secondAuthor.Get("legacy_id"))
	})

	t.Run("Update", func(t *testing.T) {
		f := prepareAuthorBook(t, client)

		// Update first book to point to second author
		u.Must(f.book.Mutation().
			Where(db.EQ("id", f.firstBookID)).
			Update(
				h.Ctx(),
				entity.New().Set("author", entity.New().Set("legacy_id", f.secondLegacyID)),
			))

		// Verify the update
		updated := u.Must(f.book.Query(db.EQ("id", f.firstBookID)).
			Select("author", "author_legacy_id").
			First(h.Ctx()))

		// The updated first book's author_legacy_id should now match the second author's legacy_id
		assert.Equal(t, f.secondLegacyID, updated.Get("author_legacy_id"))
		updatedAuthor, ok := updated.Get("author").(*entity.Entity)
		require.True(t, ok)

		// The updated first book's author entity should now be the second author
		assert.Equal(t, f.secondLegacyID, updatedAuthor.Get("legacy_id"))
	})

	t.Run("SourceColumnFilter", func(t *testing.T) {
		f := prepareAuthorBook(t, client)

		// Verify books by filtering on author_legacy_id
		results := u.Must(f.book.Query(db.EQ("author_legacy_id", f.secondLegacyID)).
			Select("title", "author_legacy_id").
			Get(h.Ctx()))
		require.Len(t, results, 1)

		// The second book's author_legacy_id should match the second author's legacy_id
		assert.Equal(t, f.secondLegacyID, results[0].Get("author_legacy_id"))
		assert.Equal(t, "Book Legacy", results[0].Get("title"))
	})

	t.Run("RelationFieldFilter", func(t *testing.T) {
		f := prepareAuthorBook(t, client)

		// Verify books by filtering on author_legacy_id
		results := u.Must(f.book.Query(db.EQ("legacy_id", f.secondLegacyID, "author")).
			Select("title", "author_legacy_id", "author").
			Get(h.Ctx()))
		require.Len(t, results, 1)
		assert.Equal(t, "Book Legacy", results[0].Get("title"))
		assert.Equal(t, f.secondLegacyID, results[0].Get("author_legacy_id"))

		// The second book's author_legacy_id should match the second author's legacy_id
		authorEntity, ok := results[0].Get("author").(*entity.Entity)
		require.True(t, ok)
		assert.Equal(t, f.secondLegacyID, authorEntity.Get("legacy_id"))
	})

	t.Run("RelationSelect", func(t *testing.T) {
		f := prepareAuthorBook(t, client)

		// Verify author with books
		authorWithBooks := u.Must(f.author.Query(db.EQ("legacy_id", f.firstLegacyID)).
			Select("name", "books").
			First(h.Ctx()))
		books, ok := authorWithBooks.Get("books").([]*entity.Entity)
		require.True(t, ok)
		require.Len(t, books, 1)

		// The first book's author_legacy_id should match the first author's legacy_id
		assert.Equal(t, f.firstLegacyID, books[0].Get("author_legacy_id"))
		assert.Equal(t, "Book Both", books[0].Get("title"))
	})

	t.Run("RelationFilterNonOwner", func(t *testing.T) {
		f := prepareAuthorBook(t, client)

		// Verify books by filtering on author's legacy_id
		books := u.Must(f.book.Query(db.EQ("legacy_id", f.firstLegacyID, "author")).
			Select("title", "author_legacy_id").
			Get(h.Ctx()))
		require.Len(t, books, 1)

		// The first book's author_legacy_id should match the first author's legacy_id
		assert.Equal(t, f.firstLegacyID, books[0].Get("author_legacy_id"))
		assert.Equal(t, "Book Both", books[0].Get("title"))
	})

	t.Run("RelationFilterOwner", func(t *testing.T) {
		f := prepareAuthorBook(t, client)
		// Verify authors by filtering on book title
		authors := u.Must(f.author.Query(db.EQ("title", "Book Both", "books")).
			Select("name").
			Get(h.Ctx()))
		require.Len(t, authors, 1)
		assert.Equal(t, "Author A", authors[0].Get("name"))
	})
}

func runO2OCustomFKTests(t *testing.T, client h.DBClient) {
	t.Run("Create", func(t *testing.T) {
		f := prepareCitizenPassport(t, client)

		// Verify first passport
		passport := u.Must(f.passport.Query(db.EQ("id", f.firstPassportID)).
			Select("number", "holder", "holder_legacy_id").
			First(h.Ctx()))

		// The first passport's holder_legacy_id should match the first citizen's legacy_id
		assert.Equal(t, f.firstLegacyID, passport.Get("holder_legacy_id"))
		holder, ok := passport.Get("holder").(*entity.Entity)
		require.True(t, ok)

		// Verify first holder entity
		assert.Equal(t, f.firstLegacyID, holder.Get("legacy_id"))
	})

	t.Run("Update", func(t *testing.T) {
		f := prepareCitizenPassport(t, client)

		// Update first passport to point to unassigned citizen
		u.Must(f.passport.Mutation().
			Where(db.EQ("id", f.firstPassportID)).
			Update(
				h.Ctx(),
				entity.New().Set("holder", entity.New().Set("legacy_id", f.unassignedLegacyID)),
			))

		// Verify the update
		updated := u.Must(f.passport.Query(db.EQ("id", f.firstPassportID)).
			Select("holder", "holder_legacy_id").
			First(h.Ctx()))

		// The updated first passport's holder_legacy_id should now match the unassigned citizen's legacy_id
		assert.Equal(t, f.unassignedLegacyID, updated.Get("holder_legacy_id"))
		newHolder, ok := updated.Get("holder").(*entity.Entity)
		require.True(t, ok)

		// The updated first passport's holder entity should now be the unassigned citizen
		assert.Equal(t, f.unassignedLegacyID, newHolder.Get("legacy_id"))
	})

	t.Run("SourceColumnFilter", func(t *testing.T) {
		f := prepareCitizenPassport(t, client)

		// Verify passports by filtering on holder_legacy_id
		results := u.Must(f.passport.Query(db.EQ("holder_legacy_id", f.secondLegacyID)).
			Select("number", "holder_legacy_id").
			Get(h.Ctx()))
		require.Len(t, results, 1)

		// The second passport's holder_legacy_id should match the second citizen's legacy_id
		assert.Equal(t, f.secondLegacyID, results[0].Get("holder_legacy_id"))
	})

	t.Run("RelationFieldFilter", func(t *testing.T) {
		f := prepareCitizenPassport(t, client)

		// Verify passports by filtering on holder_legacy_id
		results := u.Must(f.passport.Query(db.EQ("legacy_id", f.secondLegacyID, "holder")).
			Select("number", "holder_legacy_id", "holder").
			Get(h.Ctx()))
		require.Len(t, results, 1)
		assert.Equal(t, f.secondLegacyID, results[0].Get("holder_legacy_id"))

		// The second passport's holder_legacy_id should match the second citizen's legacy_id
		holderEntity, ok := results[0].Get("holder").(*entity.Entity)
		require.True(t, ok)
		assert.Equal(t, f.secondLegacyID, holderEntity.Get("legacy_id"))
	})

	t.Run("RelationSelect", func(t *testing.T) {
		f := prepareCitizenPassport(t, client)

		// Verify citizen with passport
		citizen := u.Must(f.citizen.Query(db.EQ("legacy_id", f.firstLegacyID)).
			Select("full_name", "passport", "passport.holder_legacy_id").
			First(h.Ctx()))

		// The first citizen's passport's holder_legacy_id should match the first citizen's legacy_id
		passport, ok := citizen.Get("passport").(*entity.Entity)
		require.True(t, ok)
		assert.Equal(t, f.firstLegacyID, passport.Get("holder_legacy_id"))
	})

	t.Run("RelationFilterNonOwner", func(t *testing.T) {
		f := prepareCitizenPassport(t, client)

		// Verify passports by filtering on citizen's legacy_id
		passports := u.Must(f.passport.Query(db.EQ("full_name", "Citizen A", "holder")).
			Select("number").
			Get(h.Ctx()))
		require.Len(t, passports, 1)

		// The first passport's holder_legacy_id should match the first citizen's legacy_id
		assert.Equal(t, fmt.Sprintf("P-%d", f.firstLegacyID), passports[0].Get("number"))
	})

	t.Run("RelationFilterOwner", func(t *testing.T) {
		f := prepareCitizenPassport(t, client)

		// Verify authors by filtering on book title
		citizens := u.Must(f.citizen.Query(db.EQ("number", fmt.Sprintf("P-%d", f.firstLegacyID), "passport")).
			Select("full_name").
			Get(h.Ctx()))
		require.Len(t, citizens, 1)

		// The first citizen's full_name should be "Citizen A"
		assert.Equal(t, "Citizen A", citizens[0].Get("full_name"))
	})
}

func runM2MCustomFKTests(t *testing.T, client h.DBClient) {
	t.Run("Create", func(t *testing.T) {
		f := preparePlaylistTrack(t, client)

		// Verify playlist with tracks
		playlist := u.Must(f.playlist.Query(db.EQ("id", f.mixPlaylistID)).
			Select("name", "tracks").
			First(h.Ctx()))

		// The mix playlist should have two tracks
		tracks, ok := playlist.Get("tracks").([]*entity.Entity)
		require.True(t, ok)
		require.Len(t, tracks, 2)

		// The mix playlist should contain both track codes
		codes := []any{tracks[0].Get("code"), tracks[1].Get("code")}
		assert.Contains(t, codes, f.trackOneCode)
		assert.Contains(t, codes, f.trackTwoCode)
	})

	t.Run("Update", func(t *testing.T) {
		f := preparePlaylistTrack(t, client)

		// Update mix playlist to only have the second track
		u.Must(f.playlist.Mutation().
			Where(db.EQ("id", f.mixPlaylistID)).
			Update(h.Ctx(), entity.New().Set("tracks", []*entity.Entity{
				entity.New(f.trackTwoID),
			})))

		// Verify the update
		playlist := u.Must(f.playlist.Query(db.EQ("id", f.mixPlaylistID)).
			Select("tracks").
			First(h.Ctx()))

		// The mix playlist should now have only one track
		tracks, ok := playlist.Get("tracks").([]*entity.Entity)
		require.True(t, ok)
		require.Len(t, tracks, 1)

		// The remaining track should be track two
		assert.Equal(t, f.trackTwoCode, tracks[0].Get("code"))
	})

	t.Run("RelationFilter", func(t *testing.T) {
		f := preparePlaylistTrack(t, client)

		// Verify playlists by filtering on track code
		playlists := u.Must(f.playlist.Query(db.EQ("code", f.trackOneCode, "tracks")).
			Select("code").
			Get(h.Ctx()))
		require.Len(t, playlists, 1)

		// The mix playlist should be returned when filtering by track one code
		assert.Equal(t, f.mixPlaylistCode, playlists[0].Get("code"))
	})

	t.Run("RelationSelect", func(t *testing.T) {
		f := preparePlaylistTrack(t, client)

		// Verify track with playlists
		track := u.Must(f.track.Query(db.EQ("id", f.trackTwoID)).
			Select("title", "playlists").
			First(h.Ctx()))

		// The track two should be in two playlists
		playlists, ok := track.Get("playlists").([]*entity.Entity)
		require.True(t, ok)
		require.Len(t, playlists, 2)

		// The playlists should contain both playlist codes
		codes := []any{playlists[0].Get("code"), playlists[1].Get("code")}
		assert.Contains(t, codes, f.mixPlaylistCode)
		assert.Contains(t, codes, f.singlePlaylistCode)
	})

	t.Run("RelationFilterOwner", func(t *testing.T) {
		f := preparePlaylistTrack(t, client)

		// Verify tracks by filtering on playlist code
		tracks := u.Must(f.track.Query(db.EQ("code", f.mixPlaylistCode, "playlists")).
			Select("code").
			Get(h.Ctx()))
		require.Len(t, tracks, 2)

		// The tracks should contain both track codes
		codes := []any{tracks[0].Get("code"), tracks[1].Get("code")}
		assert.Contains(t, codes, f.trackOneCode)
		assert.Contains(t, codes, f.trackTwoCode)
	})
}
