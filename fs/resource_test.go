package fs_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

type TestResourceInput struct {
	Field1 string
}

var ResourceResolver1 = func(c fs.Context, input *TestResourceInput) (*TestResourceInput, error) {
	return input, nil
}

func TestInit(t *testing.T) {
	rs := fs.NewResourcesManager()

	rs.Add(fs.NewResource("resource1", ResourceResolver1))

	err := rs.Init()

	assert.NoError(t, err, "Init should not return an error")
}

func TestInitDuplicateResourceID(t *testing.T) {
	rs := fs.NewResourcesManager()
	rs.Add(fs.NewResource("resource1", ResourceResolver1))
	rs.Add(fs.NewResource("resource1", ResourceResolver1))
	err := rs.Init()
	assert.Error(t, err, "Init should return an error")
}

func TestMangerInitError(t *testing.T) {
	rs := fs.NewResourcesManager()
	resourceWithoutName := fs.NewResource("", ResourceResolver1)
	rs.Add(resourceWithoutName)
	err := rs.Init()
	assert.Error(t, err, "Init should return an error")
}

func TestResourceInitErrorName(t *testing.T) {
	r2 := fs.NewResource("test", ResourceResolver1)
	r2Sub1 := fs.NewResource("sub1", ResourceResolver1)
	r2Sub2 := fs.NewResource("sub1", ResourceResolver1)
	r2.Add(r2Sub1)
	r2.Add(r2Sub2)
	err := r2.Init()
	assert.Error(t, err, "Init should return an error")

	r2.Remove(r2Sub2)
	r2Sub2 = fs.NewResource("", ResourceResolver1)
	r2.Add(r2Sub2)
	err = r2.Init()
	assert.Error(t, err, "Init should return an error")
}

type testContext struct {
	err error
}

func (c *testContext) ID() string               { return "test" }
func (c *testContext) User() *fs.User           { return nil }
func (c *testContext) Value(string, ...any) any { return nil }
func (c *testContext) Logger() logger.Logger    { return nil }
func (c *testContext) Parse(input any) error {
	if c.err != nil {
		return c.err
	}

	if i, ok := input.(*TestResourceInput); ok {
		i.Field1 = "test"
	}

	return nil
}
func (c *testContext) Context() context.Context         { return nil }
func (c *testContext) Args() map[string]string          { return nil }
func (c *testContext) Arg(string, ...string) string     { return "" }
func (c *testContext) ArgInt(string, ...int) int        { return 0 }
func (c *testContext) Entity() (*schema.Entity, error)  { return nil, nil }
func (c *testContext) Resource() *fs.Resource           { return nil }
func (c *testContext) AuthToken() string                { return "" }
func (c *testContext) Next() error                      { return nil }
func (c *testContext) Result(...*fs.Result) *fs.Result  { return nil }
func (c *testContext) Files() ([]*fs.File, error)       { return nil, nil }
func (c *testContext) Redirect(string) error            { return nil }
func (c *testContext) Cookies(string, ...string) string { return "" }

func TestNewResource(t *testing.T) {
	r := fs.NewResource(
		"test",
		ResourceResolver1,
		&fs.Meta{Get: "/get"},
	)
	assert.NotNil(t, r, "Resource should not be nil")

	var c fs.Context = &testContext{}

	resolver := r.Handler()
	result, err := resolver(c)
	assert.NoError(t, err, "Resolver should not return an error")
	assert.Equal(t, (*TestResourceInput)(nil), result, "Resolver should return the input")
}

func TestNewResourceResolveError(t *testing.T) {
	r := fs.NewResource(
		"test",
		func(c fs.Context, input *string) (*string, error) {
			return input, errors.New("error")
		},
		&fs.Meta{
			Get:    "/get",
			Public: true,
		},
	)
	assert.NotNil(t, r, "Resource should not be nil")
	var c fs.Context = &testContext{err: errors.New("error")}
	resolver := r.Handler()
	_, err := resolver(c)
	assert.Error(t, err, "Resolver should return an error")
}

func TestResourceWithParent(t *testing.T) {
	rs := fs.NewResourcesManager()
	rs1 := fs.NewResource("resource1", ResourceResolver1)
	rs2 := fs.NewResource("resource2", ResourceResolver1)
	rs.Add(rs1)
	rs1.Add(rs2)
	err := rs.Init()
	assert.NoError(t, err, "Init should not return an error")
	assert.Equal(t, "resource1", rs1.ID(), "Resource ID should be 'resource1'")
	assert.Equal(t, "resource1.resource2", rs2.ID(), "Resource ID should be 'resource1.resource2'")
}

func TestRemoveResource(t *testing.T) {
	rs := fs.NewResourcesManager()
	rs1 := fs.NewResource("resource1", ResourceResolver1)
	rs.Add(rs1)
	rs.Remove(rs1)
	result := rs.Find("resource1")
	assert.Nil(t, result, "Resource should be removed")
}

func TestResourcesManagerClone(t *testing.T) {
	rs := &fs.ResourcesManager{
		Resource:    &fs.Resource{},
		Middlewares: []fs.Middleware{func(c fs.Context) error { return nil }},
		Hooks:       func() *fs.Hooks { return nil },
	}

	clone := rs.Clone()

	// Assert that the cloned ResourcesManager is not the same instance as the original
	assert.NotEqual(t, rs, clone)

	// Assert that the cloned ResourcesManager has the same Resource instance as the original
	assert.Equal(t, rs.Resource, clone.Resource)

	// Assert that the cloned ResourcesManager has the same number of Middlewares as the original
	assert.Len(t, clone.Middlewares, len(rs.Middlewares))

	// Assert that the cloned ResourcesManager has the same Middleware instances as the original
	assert.Len(t, clone.Middlewares, len(rs.Middlewares))

	// Assert that the cloned ResourcesManager has the same Hooks instance as the original
	assert.NotNil(t, clone.Hooks)
}

func TestResourceClone(t *testing.T) {
	rs1 := fs.NewResource("resource1", ResourceResolver1, &fs.Meta{Get: "/get"})
	rs2 := fs.NewResource("resource2", ResourceResolver1)
	rs1 = rs1.Add(rs2)
	rsClone := rs1.Clone()
	assert.Equal(t, rs1.Name(), rsClone.Name(), "Resource name should be the same")
	assert.Equal(t, rs1.ID(), rsClone.ID(), "Resource ID should be the same")
	assert.Equal(t, rs1.Resources()[0].Name(), rsClone.Resources()[0].Name(), "Resource children should be the same")
}

func TestAddResource(t *testing.T) {
	r := fs.NewResource("parent", ResourceResolver1)
	signatures := fs.Signatures{"a", "a"}
	meta := &fs.Meta{
		Get:        "/get",
		Public:     true,
		Signatures: signatures,
	}

	resolver := func(c fs.Context) (any, error) {
		return nil, nil
	}

	r.AddResource("child", resolver, meta)
	child := r.Find("parent.child")

	assert.NotNil(t, child, "Resource should not be nil")
	assert.Equal(t, "child", child.Name(), "Resource name should be 'child'")
	assert.Equal(t, "/get", child.Meta().Get, "Resource meta should match")
	assert.True(t, child.IsPublic(), "Resource should be a white list")
	assert.Contains(t, r.Resources(), child, "Resource should be added to the parent's resources")
	assert.Equal(t, signatures, child.Signature(), "Resource signature should match")

	nonPublicResource := fs.NewResource("non-public", ResourceResolver1)
	assert.False(t, nonPublicResource.IsPublic(), "Resource should not be a white list")
}

func TestResourceResolver(t *testing.T) {
	r := fs.NewResource("test", ResourceResolver1)
	resolver := r.Handler()
	assert.NotNil(t, resolver, "Resolver should not be nil")
}

func TestIsGroup(t *testing.T) {
	m := fs.NewResourcesManager()
	assert.True(t, m.IsGroup(), "IsGroup should return true for a group resource")

	r := fs.NewResource("test", ResourceResolver1)
	assert.True(t, !r.IsGroup(), "IsGroup should return false for a group resource")
}

func TestGroup(t *testing.T) {
	r := fs.NewResource("parent", ResourceResolver1)
	child1 := fs.NewResource("child1", ResourceResolver1)
	child2 := fs.NewResource("child2", ResourceResolver1)

	group := r.Group("group", &fs.Meta{Prefix: "/group"})
	group.Add(child1, child2)

	assert.Equal(t, "[group] /group", group.String(), "String representation should match the expected value")
	assert.NotNil(t, group, "Group should not be nil")
	assert.Equal(t, "group", group.Name(), "Group name should be 'group'")
	assert.True(t, group.IsGroup(), "Group should be a group resource")
	assert.Contains(t, r.Resources(), group, "Group should be added to the parent's resources")
	assert.Contains(t, group.Resources(), child1, "Child1 should be added to the group's resources")
	assert.Contains(t, group.Resources(), child2, "Child2 should be added to the group's resources")
}

func TestResourceString(t *testing.T) {
	r := fs.NewResource("test", ResourceResolver1, &fs.Meta{
		Get:     "/get",
		Head:    "/head",
		Post:    "/post",
		Put:     "/put",
		Delete:  "/delete",
		Trace:   "/trace",
		Options: "/options",
		Connect: "/connect",
		Patch:   "/patch",
	})

	expected := "- test - GET: /get, HEAD: /head, POST: /post, PUT: /put, DELETE: /delete, CONNECT: /connect, OPTIONS: /options, TRACE: /trace, PATCH: /patch"
	result := r.String()

	assert.Equal(t, expected, result, "String representation should match the expected value")
}

func TestGroupResourceString(t *testing.T) {
	r := fs.NewResourcesManager()
	child1 := fs.NewResource("child1", ResourceResolver1)
	child2 := fs.NewResource("child2", ResourceResolver1)
	r.Add(child1)
	r.Add(child2)

	assert.Equal(t, "[root] /", r.String(), "String representation should match the expected value")
	assert.Equal(t, "- child1", child1.String(), "String representation should match the expected value")
	assert.Equal(t, "- child2", child2.String(), "String representation should match the expected value")
}
func TestPrint(t *testing.T) {
	rs := fs.NewResource("test", ResourceResolver1)
	child1 := fs.NewResource("child1", ResourceResolver1)
	child2 := fs.NewResource("child2", ResourceResolver1)
	g := rs.Group("group")
	g.Add(child1, child2)

	backupStdOut := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rs.Print()

	outputChannel := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outputChannel <- buf.String()
	}()

	w.Close()
	os.Stdout = backupStdOut
	out := <-outputChannel
	expectedOutput := "- test\n  [group] /group\n    - child1\n    - child2\n"
	assert.Equal(t, expectedOutput, out, "Print output should match the expected output")
}

func TestResourceMarshalJSON(t *testing.T) {
	rm := fs.NewResourcesManager()
	r1 := fs.NewResource("test", ResourceResolver1, &fs.Meta{
		Get:    "/get",
		Public: true,
	})
	rm.Add(r1)
	child := fs.NewResource("child", ResourceResolver1)
	r1.Add(child)
	expected := `{"id":"","name":"","group":true,"resources":[{"id":"test","name":"test","meta":{"get":"/get","public":true},"resources":[{"id":"test.child","name":"child"}]}]}`
	result, err := rm.MarshalJSON()
	assert.NoError(t, err, "MarshalJSON should not return an error")
	assert.Equal(t, expected, string(result), "Marshalled JSON should match the expected value")
}

func TestResourceMethods(t *testing.T) {
	signatures := fs.Signatures{"a", "a"}
	meta := &fs.Meta{
		Public:     true,
		Signatures: signatures,
	}

	resolver := func(c fs.Context, _ string) (string, error) {
		return "", nil
	}

	// No meta
	r := fs.Get("getresource", resolver)
	assert.NotNil(t, r)
	assert.Equal(t, "getresource", r.Name())
	assert.Equal(t, "getresource", r.Meta().Get)

	// Get
	r = fs.Get("getresource", resolver, meta)
	assert.NotNil(t, r)
	assert.Equal(t, "getresource", r.Name())
	assert.Equal(t, "getresource", r.Meta().Get)
	assert.True(t, r.IsPublic())
	assert.Equal(t, signatures, r.Meta().Signatures)

	// Head
	r = fs.Head("headresource", resolver, meta)
	assert.NotNil(t, r)
	assert.Equal(t, "headresource", r.Name())
	assert.Equal(t, "headresource", r.Meta().Head)
	assert.True(t, r.IsPublic())
	assert.Equal(t, signatures, r.Meta().Signatures)

	// Post
	r = fs.Post("Postresource", resolver, meta)
	assert.NotNil(t, r)
	assert.Equal(t, "Postresource", r.Name())
	assert.Equal(t, "Postresource", r.Meta().Post)
	assert.True(t, r.IsPublic())
	assert.Equal(t, signatures, r.Meta().Signatures)

	// Put
	r = fs.Put("Putresource", resolver, meta)
	assert.NotNil(t, r)
	assert.Equal(t, "Putresource", r.Name())
	assert.Equal(t, "Putresource", r.Meta().Put)
	assert.True(t, r.IsPublic())
	assert.Equal(t, signatures, r.Meta().Signatures)

	// Delete
	r = fs.Delete("Deleteresource", resolver, meta)
	assert.NotNil(t, r)
	assert.Equal(t, "Deleteresource", r.Name())
	assert.Equal(t, "Deleteresource", r.Meta().Delete)
	assert.True(t, r.IsPublic())
	assert.Equal(t, signatures, r.Meta().Signatures)

	// Connect
	r = fs.Connect("Connectresource", resolver, meta)
	assert.NotNil(t, r)
	assert.Equal(t, "Connectresource", r.Name())
	assert.Equal(t, "Connectresource", r.Meta().Connect)
	assert.True(t, r.IsPublic())
	assert.Equal(t, signatures, r.Meta().Signatures)

	// Options
	r = fs.Options("Optionsresource", resolver, meta)
	assert.NotNil(t, r)
	assert.Equal(t, "Optionsresource", r.Name())
	assert.Equal(t, "Optionsresource", r.Meta().Options)
	assert.True(t, r.IsPublic())
	assert.Equal(t, signatures, r.Meta().Signatures)

	// Trace
	r = fs.Trace("Traceresource", resolver, meta)
	assert.NotNil(t, r)
	assert.Equal(t, "Traceresource", r.Name())
	assert.Equal(t, "Traceresource", r.Meta().Trace)
	assert.True(t, r.IsPublic())
	assert.Equal(t, signatures, r.Meta().Signatures)

	// Patch
	r = fs.Patch("Patchresource", resolver, meta)
	assert.NotNil(t, r)
	assert.Equal(t, "Patchresource", r.Name())
	assert.Equal(t, "Patchresource", r.Meta().Patch)
	assert.True(t, r.IsPublic())
	assert.Equal(t, signatures, r.Meta().Signatures)
}
