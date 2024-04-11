package app_test

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"reflect"
	"testing"
	"time"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

func TestEntityToRole(t *testing.T) {
	e := schema.NewEntityFromMap(map[string]any{
		"id":          1,
		"name":        "Test Role",
		"description": "Test Description",
		"root":        true,
		"permissions": []*schema.Entity{
			schema.NewEntityFromMap(map[string]any{
				"resource": "Resource 1",
				"value":    "Value 1",
			}),
			schema.NewEntityFromMap(map[string]any{
				"resource": "Resource 2",
				"value":    "Value 2",
			}),
		},
	})

	expectedRole := &app.Role{
		ID:          1,
		Name:        "Test Role",
		Description: "Test Description",
		Root:        true,
		Users:       []*app.User{},
		Permissions: []*app.Permission{
			{
				Resource: "Resource 1",
				Value:    "Value 1",
			},
			{
				Resource: "Resource 2",
				Value:    "Value 2",
			},
		},
	}

	role := app.EntityToRole(e)
	assert.Equal(t, expectedRole, role)

	createdAt := time.Now()
	updatedAt := time.Now()
	deletedAt := time.Now()

	e.Set("created_at", &createdAt)
	e.Set("updated_at", &updatedAt)
	e.Set("deleted_at", &deletedAt)

	expectedRole.CreatedAt = &createdAt
	expectedRole.UpdatedAt = &updatedAt
	expectedRole.DeletedAt = &deletedAt

	role = app.EntityToRole(e)

	if !reflect.DeepEqual(role, expectedRole) {
		t.Errorf("EntityToRole() = %v, want %v", role, expectedRole)
	}
}

func TestEntitiesToRoles(t *testing.T) {
	entities := []*schema.Entity{
		schema.NewEntityFromMap(map[string]interface{}{
			"id":          1,
			"name":        "Role 1",
			"description": "Description 1",
			"root":        true,
			"permissions": []*schema.Entity{
				schema.NewEntityFromMap(map[string]interface{}{
					"resource": "Resource 1",
					"value":    "Value 1",
				}),
			},
		}),
		schema.NewEntityFromMap(map[string]interface{}{
			"id":          2,
			"name":        "Role 2",
			"description": "Description 2",
			"root":        false,
			"permissions": []*schema.Entity{
				schema.NewEntityFromMap(map[string]interface{}{
					"resource": "Resource 2",
					"value":    "Value 2",
				}),
			},
		}),
	}

	expectedRoles := []*app.Role{
		{
			ID:          1,
			Name:        "Role 1",
			Description: "Description 1",
			Root:        true,
			Users:       []*app.User{},
			Permissions: []*app.Permission{
				{
					Resource: "Resource 1",
					Value:    "Value 1",
				},
			},
		},
		{
			ID:          2,
			Name:        "Role 2",
			Description: "Description 2",
			Root:        false,
			Users:       []*app.User{},
			Permissions: []*app.Permission{
				{
					Resource: "Resource 2",
					Value:    "Value 2",
				},
			},
		},
	}

	roles := app.EntitiesToRoles(entities)
	assert.Equal(t, expectedRoles, roles)
}

func TestEntityToUserNil(t *testing.T) {
	user := app.EntityToUser(nil)

	if user != nil {
		t.Errorf("EntityToUser() = %v, want nil", user)
	}
}

func TestEntityToUser(t *testing.T) {
	createdAt := time.Now()
	updatedAt := time.Now()
	deletedAt := time.Now()

	role1 := schema.NewEntityFromMap(map[string]any{
		"id":   1,
		"name": "Role 1",
	})

	role2 := schema.NewEntityFromMap(map[string]any{
		"id":   2,
		"name": "Role 2",
	})

	e := schema.NewEntityFromMap(map[string]any{
		"id":                1,
		"username":          "testuser",
		"email":             "test@example.com",
		"password":          "password123",
		"provider":          "local",
		"provider_id":       "123456",
		"provider_username": "testuser",
		"roles":             []*schema.Entity{role1, role2},
		"role_ids":          []uint64{1, 2},
		"active":            true,
		"created_at":        &createdAt,
		"updated_at":        &updatedAt,
		"deleted_at":        &deletedAt,
	})

	expectedUser := &app.User{
		ID:               1,
		Username:         "testuser",
		Email:            "test@example.com",
		Password:         "password123",
		Provider:         "local",
		ProviderID:       "123456",
		ProviderUsername: "testuser",
		Roles: []*app.Role{
			app.EntityToRole(role1),
			app.EntityToRole(role2),
		},
		RoleIDs:   []uint64{1, 2},
		Active:    true,
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
		DeletedAt: &deletedAt,
	}

	user := app.EntityToUser(e)
	assert.Equal(t, expectedUser, user)
}

// MockDisk is a mock implementation of the app.Disk interface.
type MockDisk struct{}

func (d *MockDisk) Name() string            { return "mock" }
func (d *MockDisk) LocalPublicPath() string { return "" }
func (d *MockDisk) Root() string            { return "" }
func (d *MockDisk) URL(filepath string) string {
	return fmt.Sprintf("http://example.com%s", filepath)
}
func (d *MockDisk) Delete(c context.Context, filepath string) error          { return nil }
func (d *MockDisk) Put(c context.Context, file *app.File) (*app.File, error) { return nil, nil }
func (d *MockDisk) PutReader(c context.Context, in io.Reader, size uint64, mime, dst string) (*app.File, error) {
	return nil, nil
}
func (d *MockDisk) PutMultipart(c context.Context, m *multipart.FileHeader, dsts ...string) (*app.File, error) {
	return nil, nil
}

func TestEntityToFile(t *testing.T) {
	createdAt := time.Now()
	updatedAt := time.Now()
	deletedAt := time.Now()
	e := schema.NewEntityFromMap(map[string]interface{}{
		"id":         1,
		"disk":       "disk1",
		"name":       "test.txt",
		"path":       "/path/to/file",
		"type":       "text/plain",
		"size":       uint64(1024),
		"user_id":    uint64(123),
		"created_at": &createdAt,
		"updated_at": &updatedAt,
		"deleted_at": &deletedAt,
	})

	expectedFile := &app.File{
		ID:        1,
		Disk:      "disk1",
		Name:      "test.txt",
		Path:      "/path/to/file",
		Type:      "text/plain",
		Size:      1024,
		UserID:    123,
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
		DeletedAt: &deletedAt,
	}

	file := app.EntityToFile(e)
	assert.Equal(t, expectedFile, file)

	var disk app.Disk = &MockDisk{}
	file2 := app.EntityToFile(e, disk)
	assert.Equal(t, "http://example.com/path/to/file", file2.URL)
}

func TestEntityToFileNil(t *testing.T) {
	file := app.EntityToFile(nil)

	if file != nil {
		t.Errorf("EntityToFile() = %v, want nil", file)
	}
}
func TestEntitiesToFiles(t *testing.T) {
	createdAt := time.Now()
	updatedAt := time.Now()
	deletedAt := time.Now()
	entities := []*schema.Entity{
		schema.NewEntityFromMap(map[string]interface{}{
			"id":         1,
			"disk":       "disk1",
			"name":       "test1.txt",
			"path":       "/path/to/file1",
			"type":       "text/plain",
			"size":       uint64(1024),
			"user_id":    uint64(123),
			"created_at": &createdAt,
			"updated_at": &updatedAt,
			"deleted_at": &deletedAt,
		}),
		schema.NewEntityFromMap(map[string]interface{}{
			"id":         2,
			"disk":       "disk2",
			"name":       "test2.txt",
			"path":       "/path/to/file2",
			"type":       "text/plain",
			"size":       uint64(2048),
			"user_id":    uint64(456),
			"created_at": &createdAt,
			"updated_at": &updatedAt,
			"deleted_at": &deletedAt,
		}),
	}

	expectedFiles := []*app.File{
		{
			ID:        1,
			Disk:      "disk1",
			Name:      "test1.txt",
			Path:      "/path/to/file1",
			Type:      "text/plain",
			Size:      1024,
			UserID:    123,
			CreatedAt: &createdAt,
			UpdatedAt: &updatedAt,
			DeletedAt: &deletedAt,
		},
		{
			ID:        2,
			Disk:      "disk2",
			Name:      "test2.txt",
			Path:      "/path/to/file2",
			Type:      "text/plain",
			Size:      2048,
			UserID:    456,
			CreatedAt: &createdAt,
			UpdatedAt: &updatedAt,
			DeletedAt: &deletedAt,
		},
	}

	files := app.EntitiesToFiles(entities)

	assert.Equal(t, expectedFiles, files)
}
