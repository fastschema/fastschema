// Package auditlog carries the request actor (who/where) from the HTTP layer
// down to the DB hooks that write audit-trail entries.
//
// The actor cannot travel via context.WithValue: fs.Context delegates Value()
// to the underlying fasthttp request context, which is unaware of arbitrary
// context wrappers. Instead the actor is stashed in fiber Locals (fasthttp
// UserValue) on the very request context object that flows unchanged from the
// handler into db/mutation -> entdbadapter -> the post-mutation hooks, so a
// later ctx.Value() reads it back.
package auditlog

import (
	"context"

	"github.com/fastschema/fastschema/fs"
	"github.com/google/uuid"
)

// actorContextKey keys the actor in fiber Locals. fs.Context.Local only accepts
// a string key, so this is a plain (namespaced) string rather than an
// unexported type; the namespace avoids collision with other Locals.
const actorContextKey = "fastschema.audit.actor"

// ActorContext is the request-scoped "who did it / from where" snapshot.
type ActorContext struct {
	UserID   *uuid.UUID
	UserType string // fs.ActivityActor{User,Guest,System}
	IP       string
	Method   string
	Path     string
	TraceID  string
}

// requestInfo exposes HTTP request details not present on the fs.Context
// interface itself; the concrete restfulresolver.Context implements it.
type requestInfo interface {
	Method() string
	Path() string
}

// ActorFromRequest builds an ActorContext from the current request. Call it
// after the user has been resolved into Locals (e.g. in ParseUser). Guests
// (no authenticated user) get UserType=guest and a nil UserID.
func ActorFromRequest(c fs.Context) *ActorContext {
	if c == nil {
		return nil
	}

	actor := &ActorContext{
		UserType: fs.ActivityActorGuest,
		IP:       c.IP(),
		TraceID:  c.TraceID(),
	}

	if user := c.User(); user != nil {
		id := user.ID
		actor.UserID = &id
		actor.UserType = fs.ActivityActorUser
	}

	if req, ok := c.(requestInfo); ok {
		actor.Method = req.Method()
		actor.Path = req.Path()
	}

	return actor
}

// WithActor stashes the actor on the request context (fiber Locals).
func WithActor(c fs.Context, actor *ActorContext) {
	if c == nil || actor == nil {
		return
	}

	c.Local(actorContextKey, actor)
}

// ActorFromContext reads the actor stashed by WithActor. It returns nil for
// non-HTTP contexts (background jobs, migrations) so callers can fall back to
// a system actor.
func ActorFromContext(ctx context.Context) *ActorContext {
	if ctx == nil {
		return nil
	}

	if actor, ok := ctx.Value(actorContextKey).(*ActorContext); ok {
		return actor
	}

	return nil
}
