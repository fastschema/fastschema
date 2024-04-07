package app_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

type TestResourceInput struct {
	Field1 string
}

var ResourceResolver1 = func(c app.Context, input *TestResourceInput) (*TestResourceInput, error) {
	return input, nil
}

func TestRegisterStaticResources(t *testing.T) {
	rs := &app.ResourcesManager{}
	configs := []*app.StaticResourceConfig{
		{
			Root:       nil,
			BasePath:   "/",
			PathPrefix: "/",
		},
		{
			Root:       nil,
			BasePath:   "/",
			PathPrefix: "/",
		},
	}

	rs.RegisterStaticResources(configs...)

	assert.Equal(t, configs, rs.StaticResources, "Registered static resources should match the input configs")
}

func TestInit(t *testing.T) {
	rs := app.NewResourcesManager()

	rs.Add(app.NewResource("resource1", ResourceResolver1))

	err := rs.Init()

	assert.NoError(t, err, "Init should not return an error")
}

func TestInitDuplicateResourceID(t *testing.T) {
	rs := app.NewResourcesManager()
	rs.Add(app.NewResource("resource1", ResourceResolver1))
	rs.Add(app.NewResource("resource1", ResourceResolver1))
	err := rs.Init()
	assert.Error(t, err, "Init should return an error")
}

func TestMangerInitError(t *testing.T) {
	rs := app.NewResourcesManager()
	resourceWithoutName := app.NewResource("", ResourceResolver1)
	rs.Add(resourceWithoutName)
	err := rs.Init()
	assert.Error(t, err, "Init should return an error")
}

func TestResourceInitErrorName(t *testing.T) {
	r1 := app.NewResource("test-aaa", ResourceResolver1)
	err := r1.Init()
	assert.Error(t, err, "Init should return an error")

	r2 := app.NewResource("test", ResourceResolver1)
	r2Sub1 := app.NewResource("sub1", ResourceResolver1)
	r2Sub2 := app.NewResource("sub1", ResourceResolver1)
	r2.Add(r2Sub1)
	r2.Add(r2Sub2)
	err = r2.Init()
	assert.Error(t, err, "Init should return an error")

	r2.Remove(r2Sub2)
	r2Sub2 = app.NewResource("sub-2", ResourceResolver1)
	r2.Add(r2Sub2)
	err = r2.Init()
	assert.Error(t, err, "Init should return an error")
}

type testContext struct{}

var testInput = &TestResourceInput{Field1: "test"}

func (c *testContext) ID() string               { return "test" }
func (c *testContext) User() *app.User          { return nil }
func (c *testContext) Value(string, ...any) any { return nil }
func (c *testContext) Logger() app.Logger       { return nil }
func (c *testContext) Parse(input any) error {
	if _, ok := input.(*string); ok {
		return errors.New("error")
	}

	if i, ok := input.(*TestResourceInput); ok {
		i.Field1 = "test"
	}

	return nil
}
func (c *testContext) Context() context.Context          { return nil }
func (c *testContext) Args() map[string]string           { return nil }
func (c *testContext) Arg(string, ...string) string      { return "" }
func (c *testContext) ArgInt(string, ...int) int         { return 0 }
func (c *testContext) Entity() (*schema.Entity, error)   { return nil, nil }
func (c *testContext) Resource() *app.Resource           { return nil }
func (c *testContext) AuthToken() string                 { return "" }
func (c *testContext) Next() error                       { return nil }
func (c *testContext) Result(...*app.Result) *app.Result { return nil }
func (c *testContext) Files() ([]*app.File, error)       { return nil, nil }

func TestNewResource(t *testing.T) {
	r := app.NewResource(
		"test",
		ResourceResolver1,
		app.Map{"key": "value"},
		true,
		app.Signature{&TestResourceInput{}, &TestResourceInput{}},
	)
	assert.NotNil(t, r, "Resource should not be nil")

	var c app.Context = &testContext{}

	resolver := r.Resolver()
	result, err := resolver(c)
	assert.NoError(t, err, "Resolver should not return an error")
	assert.Equal(t, testInput, result, "Resolver should return the input")
}

func TestNewResourceResolveError(t *testing.T) {
	r := app.NewResource(
		"test",
		func(c app.Context, input *string) (*string, error) {
			return input, nil
		},
		app.Map{"key": "value"},
		true,
		app.Signature{&TestResourceInput{}, &TestResourceInput{}},
	)
	assert.NotNil(t, r, "Resource should not be nil")

	var c app.Context = &testContext{}

	resolver := r.Resolver()
	_, err := resolver(c)
	assert.Error(t, err, "Resolver should return an error")
}

func TestResourceWithParent(t *testing.T) {
	rs := app.NewResourcesManager()
	rs1 := app.NewResource("resource1", ResourceResolver1)
	rs2 := app.NewResource("resource2", ResourceResolver1)
	rs.Add(rs1)
	rs1.Add(rs2)
	err := rs.Init()
	assert.NoError(t, err, "Init should not return an error")
	assert.Equal(t, "resource1", rs1.ID(), "Resource ID should be 'resource1'")
	assert.Equal(t, "resource1.resource2", rs2.ID(), "Resource ID should be 'resource1.resource2'")
}

func TestRemoveResource(t *testing.T) {
	rs := app.NewResourcesManager()
	rs1 := app.NewResource("resource1", ResourceResolver1)
	rs.Add(rs1)
	rs.Remove(rs1)
	result := rs.Find("resource1")
	assert.Nil(t, result, "Resource should be removed")
}

func TestResourceClone(t *testing.T) {
	rs1 := app.NewResource("resource1", ResourceResolver1)
	rs2 := app.NewResource("resource2", ResourceResolver1)
	rs1 = rs1.Add(rs2)
	rsClone := rs1.Clone()
	assert.Equal(t, rs1.Name(), rsClone.Name(), "Resource name should be the same")
	assert.Equal(t, rs1.ID(), rsClone.ID(), "Resource ID should be the same")
	assert.Equal(t, rs1.Resources()[0].Name(), rsClone.Resources()[0].Name(), "Resource children should be the same")
}

func TestAddResource(t *testing.T) {
	r := app.NewResource("parent", ResourceResolver1)
	extras := []interface{}{
		app.Map{"key": "value"},
		app.Signature{"param1", "param2"},
		true,
	}

	resolver := func(c app.Context) (any, error) {
		return nil, nil
	}

	r.AddResource("child", resolver, extras...)
	child := r.Find("parent.child")

	assert.NotNil(t, child, "Resource should not be nil")
	assert.Equal(t, "child", child.Name(), "Resource name should be 'child'")
	assert.Equal(t, app.Map{"key": "value"}, child.Meta(), "Resource meta should match")
	assert.True(t, child.WhiteListed(), "Resource should be a white list")
	assert.Contains(t, r.Resources(), child, "Resource should be added to the parent's resources")
}

func TestResourceResolver(t *testing.T) {
	r := app.NewResource("test", ResourceResolver1)
	resolver := r.Resolver()
	assert.NotNil(t, resolver, "Resolver should not be nil")
}

func TestIsGroup(t *testing.T) {
	m := app.NewResourcesManager()
	assert.True(t, m.IsGroup(), "IsGroup should return true for a group resource")

	r := app.NewResource("test", ResourceResolver1)
	assert.True(t, !r.IsGroup(), "IsGroup should return false for a group resource")
}

func TestGroup(t *testing.T) {
	r := app.NewResource("parent", ResourceResolver1)
	child1 := app.NewResource("child1", ResourceResolver1)
	child2 := app.NewResource("child2", ResourceResolver1)

	group := r.Group("group", child1, child2)

	assert.NotNil(t, group, "Group should not be nil")
	assert.Equal(t, "group", group.Name(), "Group name should be 'group'")
	assert.True(t, group.IsGroup(), "Group should be a group resource")
	assert.Contains(t, r.Resources(), group, "Group should be added to the parent's resources")
	assert.Contains(t, group.Resources(), child1, "Child1 should be added to the group's resources")
	assert.Contains(t, group.Resources(), child2, "Child2 should be added to the group's resources")
}

func TestResourceString(t *testing.T) {
	r := app.NewResource("test", ResourceResolver1, app.Map{"key": "value"})

	expected := "[test] - map[key:value]"
	result := r.String()

	assert.Equal(t, expected, result, "String representation should match the expected value")
}

func TestGroupResourceString(t *testing.T) {
	r := app.NewResourcesManager()
	child1 := app.NewResource("child1", ResourceResolver1)
	child2 := app.NewResource("child2", ResourceResolver1)
	r.Add(child1)
	r.Add(child2)

	assert.Equal(t, "[]", r.String(), "String representation should match the expected value")
	assert.Equal(t, "[child1]", child1.String(), "String representation should match the expected value")
	assert.Equal(t, "[child2]", child2.String(), "String representation should match the expected value")
}
func TestPrint(t *testing.T) {
	rs := app.NewResource("test", ResourceResolver1)
	child1 := app.NewResource("child1", ResourceResolver1)
	child2 := app.NewResource("child2", ResourceResolver1)
	rs.Group("group", child1, child2)

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
	expectedOutput := "[test]\n  [group]\n    [child1]\n    [child2]\n"
	assert.Equal(t, expectedOutput, out, "Print output should match the expected output")
}
func TestResourceMarshalJSON(t *testing.T) {
	rm := app.NewResourcesManager()
	r1 := app.NewResource("test", ResourceResolver1, app.Meta{"key": "value"}, true)
	rm.Add(r1)
	child := app.NewResource("child", ResourceResolver1)
	r1.Add(child)
	expected := `{"id":"","name":"","group":true,"resources":[{"id":"test","name":"test","meta":{"key":"value"},"whitelist":true,"resources":[{"id":"test.child","name":"child"}]}]}`
	result, err := rm.MarshalJSON()
	assert.NoError(t, err, "MarshalJSON should not return an error")
	assert.Equal(t, expected, string(result), "Marshalled JSON should match the expected value")
}
