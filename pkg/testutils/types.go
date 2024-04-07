package testutils

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/schema"
)

type MockTestCreateData struct {
	Name        string
	Schema      string
	InputJSON   string
	Run         func(model app.Model, entity *schema.Entity) error
	Expect      func(sqlmock.Sqlmock)
	ExpectError string
	Transaction bool
}

type MockTestUpdateData struct {
	Name         string
	Schema       string
	InputJSON    string
	Run          func(model app.Model, entity *schema.Entity) (int, error)
	Expect       func(sqlmock.Sqlmock)
	Predicates   []*app.Predicate
	WantErr      bool
	WantAffected int
	Transaction  bool
}

type MockTestDeleteData struct {
	Name         string
	Schema       string
	Run          func(model app.Model) (int, error)
	Expect       func(sqlmock.Sqlmock)
	Predicates   []*app.Predicate
	WantErr      bool
	WantAffected int
	Transaction  bool
}

type MockTestQueryData struct {
	Name    string
	Schema  string
	Filter  string
	Limit   uint
	Offset  uint
	Columns []string
	Order   []string
	Expect  func(sqlmock.Sqlmock)
	Run     func(
		model app.Model,
		predicates []*app.Predicate,
		limit, offset uint,
		order []string,
		columns ...string,
	) ([]*schema.Entity, error)
	ExpectError    string
	ExpectEntities []*schema.Entity
}

type MockTestCountData struct {
	Name   string
	Schema string
	Filter string
	Column string
	Unique bool
	Expect func(sqlmock.Sqlmock)
	Run    func(
		model app.Model,
		predicates []*app.Predicate,
		unique bool,
		column string,
	) (int, error)
	ExpectError string
	ExpectCount int
}

type DBTestCreateData struct {
	Name        string
	Schema      string
	InputJSON   string
	ClearTables []string
	Run         func(model app.Model, entity *schema.Entity) (*schema.Entity, error)
	Prepare     func(t *testing.T)
	Expect      func(t *testing.T, m app.Model, e *schema.Entity)
	WantErr     bool
	ExpectError error
}

type DBTestUpdateData struct {
	Name         string
	Schema       string
	InputJSON    string
	ClearTables  []string
	Run          func(model app.Model, entity *schema.Entity) (int, error)
	Expect       func(t *testing.T, m app.Model)
	Prepare      func(t *testing.T, m app.Model)
	Predicates   []*app.Predicate
	WantErr      bool
	WantAffected int
	Transaction  bool
}

type DBTestDeleteData struct {
	Name         string
	Schema       string
	ClearTables  []string
	Run          func(model app.Model) (int, error)
	Expect       func(t *testing.T, m app.Model)
	Prepare      func(t *testing.T, m app.Model)
	Predicates   []*app.Predicate
	WantErr      bool
	ExpectError  error
	WantAffected int
	Transaction  bool
}

type DBTestQueryData struct {
	Name        string
	Schema      string
	Filter      string
	Limit       uint
	Offset      uint
	Columns     []string
	Order       []string
	ClearTables []string
	Run         func(
		model app.Model,
		predicates []*app.Predicate,
		limit, offset uint,
		order []string,
		columns ...string,
	) ([]*schema.Entity, error)
	Prepare func(t *testing.T, client app.DBClient, m app.Model) []*schema.Entity
	Expect  func(
		t *testing.T,
		m app.Model,
		preparedEntities []*schema.Entity,
		results []*schema.Entity,
	)
	ExpectError string
}

type DBTestCountData struct {
	Name        string
	Schema      string
	Filter      string
	Column      string
	Unique      bool
	ClearTables []string
	Run         func(
		model app.Model,
		predicates []*app.Predicate,
		unique bool,
		column string,
	) (int, error)
	Prepare func(t *testing.T, client app.DBClient, m app.Model) int
	Expect  func(
		t *testing.T,
		m app.Model,
		preparedCount int,
		results int,
	)
	ExpectError string
}
