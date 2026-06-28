package auditlog

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// fakeCtx mimics how the request context resolves a stashed value by key,
// without using context.WithValue (which vet flags for basic-type keys).
type fakeCtx struct {
	context.Context
	key string
	val any
}

func (f fakeCtx) Value(k any) any {
	if k == f.key {
		return f.val
	}

	return f.Context.Value(k)
}

func TestActorFromContextNilForBackground(t *testing.T) {
	// Non-HTTP contexts carry no actor; callers fall back to the system actor.
	assert.Nil(t, ActorFromContext(context.Background()))
	assert.Nil(t, ActorFromContext(nil))
}

func TestActorFromContextRoundTrip(t *testing.T) {
	id := uuid.New()
	want := &ActorContext{UserID: &id, UserType: "user", IP: "127.0.0.1"}

	ctx := fakeCtx{Context: context.Background(), key: actorContextKey, val: want}

	got := ActorFromContext(ctx)
	assert.NotNil(t, got)
	assert.Equal(t, want, got)
}

func TestActorFromContextWrongType(t *testing.T) {
	ctx := fakeCtx{Context: context.Background(), key: actorContextKey, val: "not-an-actor"}
	assert.Nil(t, ActorFromContext(ctx))
}
