package fk

import (
	"fmt"
	"testing"

	"github.com/fastschema/fastschema/db"
	u "github.com/fastschema/fastschema/pkg/utils"
	h "github.com/fastschema/fastschema/tests/integration/helpers"
)

type authorBook struct {
	author         db.Model
	book           db.Model
	firstLegacyID  uint64
	secondLegacyID uint64
	firstBookID    any
	secondBookID   any
}

func prepareAuthorBook(t *testing.T, client h.DBClient) authorBook {
	t.Helper()
	h.ClearDBData(client.C, "book_fk", "author_fk")

	authorModel := u.Must(client.C.Model("author_fk"))
	bookModel := u.Must(client.C.Model("book_fk"))

	const (
		firstLegacy  = 5001
		secondLegacy = 7002
	)

	u.Must(authorModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"name":"Author A","legacy_id":%d}`, firstLegacy),
	))
	u.Must(authorModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"name":"Author B","legacy_id":%d}`, secondLegacy),
	))

	firstBookID := u.Must(bookModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"title":"Book Both","author":{"legacy_id":%d}}`, firstLegacy),
	))
	secondBookID := u.Must(bookModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"title":"Book Legacy","author":{"legacy_id":%d}}`, secondLegacy),
	))

	return authorBook{
		author:         authorModel,
		book:           bookModel,
		firstLegacyID:  uint64(firstLegacy),
		secondLegacyID: uint64(secondLegacy),
		firstBookID:    firstBookID,
		secondBookID:   secondBookID,
	}
}

type citizenPassport struct {
	citizen            db.Model
	passport           db.Model
	firstLegacyID      uint64
	secondLegacyID     uint64
	unassignedLegacyID uint64
	firstPassportID    any
	secondPassportID   any
}

func prepareCitizenPassport(t *testing.T, client h.DBClient) citizenPassport {
	t.Helper()
	h.ClearDBData(client.C, "passport_fk", "citizen_fk")

	citizenModel := u.Must(client.C.Model("citizen_fk"))
	passportModel := u.Must(client.C.Model("passport_fk"))

	const (
		firstLegacy      = 9001
		secondLegacy     = 9002
		unassignedLegacy = 9003
	)

	u.Must(citizenModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"full_name":"Citizen A","legacy_id":%d}`, firstLegacy),
	))
	u.Must(citizenModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"full_name":"Citizen B","legacy_id":%d}`, secondLegacy),
	))
	u.Must(citizenModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"full_name":"Citizen C","legacy_id":%d}`, unassignedLegacy),
	))

	firstPassportID := u.Must(passportModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"number":"P-%d","holder":{"legacy_id":%d}}`, firstLegacy, firstLegacy),
	))
	secondPassportID := u.Must(passportModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"number":"P-%d","holder":{"legacy_id":%d}}`, secondLegacy, secondLegacy),
	))

	return citizenPassport{
		citizen:            citizenModel,
		passport:           passportModel,
		firstLegacyID:      uint64(firstLegacy),
		secondLegacyID:     uint64(secondLegacy),
		unassignedLegacyID: uint64(unassignedLegacy),
		firstPassportID:    firstPassportID,
		secondPassportID:   secondPassportID,
	}
}

type playlistTrack struct {
	playlist           db.Model
	track              db.Model
	trackOneCode       uint64
	trackTwoCode       uint64
	trackOneID         any
	trackTwoID         any
	mixPlaylistID      any
	mixPlaylistCode    uint64
	singlePlaylistID   any
	singlePlaylistCode uint64
}

func preparePlaylistTrack(t *testing.T, client h.DBClient) playlistTrack {
	t.Helper()
	h.ClearDBData(client.C, "playlist_track_fk", "playlist_fk", "track_fk")

	playlistModel := u.Must(client.C.Model("playlist_fk"))
	trackModel := u.Must(client.C.Model("track_fk"))

	const (
		trackOneCode       = 3001
		trackTwoCode       = 3002
		mixPlaylistCode    = 8001
		singlePlaylistCode = 8002
	)

	trackOneID := u.Must(trackModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"title":"Track 1","code":%d}`, trackOneCode),
	))
	trackTwoID := u.Must(trackModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"title":"Track 2","code":%d}`, trackTwoCode),
	))

	mixPlaylistID := u.Must(playlistModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"name":"Mix","code":%d,"tracks":[{"id":%d},{"id":%d}]}`, mixPlaylistCode, trackOneID, trackTwoID),
	))
	singlePlaylistID := u.Must(playlistModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(`{"name":"Slow","code":%d,"tracks":[{"id":%d}]}`, singlePlaylistCode, trackTwoID),
	))

	return playlistTrack{
		playlist:           playlistModel,
		track:              trackModel,
		trackOneCode:       trackOneCode,
		trackTwoCode:       trackTwoCode,
		trackOneID:         trackOneID,
		trackTwoID:         trackTwoID,
		mixPlaylistID:      mixPlaylistID,
		mixPlaylistCode:    mixPlaylistCode,
		singlePlaylistID:   singlePlaylistID,
		singlePlaylistCode: singlePlaylistCode,
	}
}
