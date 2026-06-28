package activityservice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigureHardSkipsActivity(t *testing.T) {
	s := &ActivityService{}

	// Even if the operator's skip list omits "activity", it stays skipped to
	// prevent recursion.
	s.Configure([]string{"foo", "bar"}, nil)

	assert.True(t, s.shouldSkip("_activity"))
	assert.True(t, s.shouldSkip("foo"))
	assert.True(t, s.shouldSkip("bar"))
	// default skips no longer apply once a custom list is given
	assert.False(t, s.shouldSkip("session"))
}

func TestConfigureEmptyUsesDefaults(t *testing.T) {
	s := &ActivityService{}
	s.Configure(nil, nil)

	assert.True(t, s.shouldSkip("_activity"))
	assert.True(t, s.shouldSkip("session"))
	assert.True(t, s.shouldSkip("migration"))
}

func TestConfigureCustomRedactFields(t *testing.T) {
	s := &ActivityService{}
	s.Configure(nil, []string{"ssn"})

	assert.True(t, s.redactor.shouldRedact("ssn"))
	assert.False(t, s.redactor.shouldRedact("password"))
}
