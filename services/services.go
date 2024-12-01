package services

import (
	"errors"

	"github.com/fastschema/fastschema/fs"
	authservice "github.com/fastschema/fastschema/services/auth"
	contentservice "github.com/fastschema/fastschema/services/content"
	fileservice "github.com/fastschema/fastschema/services/file"
	realtimeservice "github.com/fastschema/fastschema/services/realtime"
	roleservice "github.com/fastschema/fastschema/services/role"
	schemaservice "github.com/fastschema/fastschema/services/schema"
	toolservice "github.com/fastschema/fastschema/services/tool"
)

type File = fileservice.FileService
type Role = roleservice.RoleService
type Schema = schemaservice.SchemaService
type Content = contentservice.ContentService
type Tool = toolservice.ToolService
type Auth = authservice.AuthService
type Realtime = realtimeservice.RealtimeService

type Services struct {
	file     *File
	role     *Role
	schema   *Schema
	content  *Content
	tool     *Tool
	auth     *Auth
	realtime *Realtime
}

type ServiceType interface {
	File | Role | Schema | Content | Tool | Auth | Realtime
}

type ServicesProvider interface {
	Services() *Services
}

func Get[T ServiceType](app any) (*T, error) {
	appServices, ok := app.(ServicesProvider)
	if !ok {
		return nil, errors.New("app does not implement ServicesProvider")
	}

	var st T
	var services = appServices.Services()

	switch any(st).(type) {
	case authservice.AuthService:
		return any(services.Auth()).(*T), nil
	case contentservice.ContentService:
		return any(services.Content()).(*T), nil
	case fileservice.FileService:
		return any(services.File()).(*T), nil
	case realtimeservice.RealtimeService:
		return any(services.Realtime()).(*T), nil
	case roleservice.RoleService:
		return any(services.Role()).(*T), nil
	case schemaservice.SchemaService:
		return any(services.Schema()).(*T), nil
	case toolservice.ToolService:
		return any(services.Tool()).(*T), nil
	}

	return nil, errors.New("service not found")
}

func New(app fs.App) *Services {
	return &Services{
		file:     fileservice.New(app),
		role:     roleservice.New(app),
		schema:   schemaservice.New(app),
		content:  contentservice.New(app),
		tool:     toolservice.New(app),
		auth:     authservice.New(app),
		realtime: realtimeservice.New(app),
	}
}

func (s *Services) File() *fileservice.FileService {
	return s.file
}

func (s *Services) Role() *roleservice.RoleService {
	return s.role
}

func (s *Services) Schema() *schemaservice.SchemaService {
	return s.schema
}

func (s *Services) Content() *contentservice.ContentService {
	return s.content
}

func (s *Services) Tool() *toolservice.ToolService {
	return s.tool
}

func (s *Services) Auth() *authservice.AuthService {
	return s.auth
}

func (s *Services) Realtime() *realtimeservice.RealtimeService {
	return s.realtime
}
