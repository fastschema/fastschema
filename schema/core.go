package schema

var PermissionSchema = `{
  "name": "permission",
  "namespace": "permissions",
  "label_field": "resource",
  "is_system_schema": true,
  "db": {
    "indexes": [
      {
        "name": "role_id_resource",
        "unique": true,
        "columns": [
          "role_id",
          "resource"
        ]
      }
    ]
  },
  "fields": [
    {
      "is_system_field": true,
      "name": "resource",
      "label": "Resource",
      "type": "string"
    },
    {
      "is_system_field": true,
      "name": "value",
      "label": "Value",
      "type": "string"
    },
    {
      "is_system_field": true,
      "name": "role",
      "label": "Role",
      "type": "relation",
      "optional": true,
      "relation": {
        "type": "o2m",
        "schema": "role",
        "field": "permissions",
        "owner": false
      }
    }
  ]
}`

var RoleSchema = `{
  "name": "role",
  "namespace": "roles",
  "label_field": "name",
  "is_system_schema": true,
  "fields": [
    {
      "is_system_field": true,
      "name": "name",
      "label": "Name",
      "type": "string",
			"unique": true,
			"optional": false
    },
    {
      "is_system_field": true,
      "name": "description",
      "label": "Description",
      "type": "string",
			"optional": true
    },
    {
      "is_system_field": true,
      "name": "root",
      "label": "Root",
      "type": "bool",
			"optional": true
    },
    {
      "is_system_field": true,
      "name": "permissions",
      "label": "Permissions",
      "type": "relation",
      "optional": true,
      "relation": {
        "type": "o2m",
        "schema": "permission",
        "field": "role",
        "owner": true
      }
    },
		{
      "is_system_field": true,
      "name": "users",
      "label": "Users",
      "type": "relation",
      "optional": true,
      "relation": {
        "type": "m2m",
        "schema": "user",
        "field": "roles",
        "owner": true
      }
    }
  ]
}`

var UserSchema = `{
  "name": "user",
  "namespace": "users",
  "label_field": "username",
  "is_system_schema": true,
  "fields": [
    {
      "is_system_field": true,
      "name": "username",
      "label": "Username",
      "type": "string",
      "sortable": true,
      "filterable": true,
      "unique": true
    },
    {
      "is_system_field": true,
      "name": "email",
      "label": "Email",
      "type": "string",
      "sortable": true,
      "filterable": true,
      "unique": true,
      "optional": true
    },
    {
      "is_system_field": true,
      "name": "password",
      "label": "Password",
      "type": "string",
			"optional": true
    },
    {
      "is_system_field": true,
      "name": "active",
      "label": "Active",
      "type": "bool",
      "optional": true
    },
    {
      "is_system_field": true,
      "name": "provider",
      "label": "Provider",
      "type": "string",
      "sortable": true,
      "filterable": true,
			"optional": true
    },
    {
      "is_system_field": true,
      "name": "provider_id",
      "label": "Provider ID",
      "type": "string",
      "optional": true
    },
    {
      "is_system_field": true,
      "name": "provider_username",
      "label": "Provider Username",
      "type": "string",
      "optional": true
    },
		{
      "is_system_field": true,
      "name": "roles",
      "label": "Roles",
      "type": "relation",
      "optional": true,
      "relation": {
        "type": "m2m",
        "schema": "role",
        "field": "users",
        "owner": false
      }
    },
    {
      "is_system_field": true,
      "name": "medias",
      "label": "Medias",
      "type": "relation",
      "optional": true,
      "relation": {
        "type": "o2m",
        "schema": "media",
        "field": "user",
        "owner": true
      }
    }
  ]
}`

var MediaSchema = `{
  "name": "media",
  "namespace": "medias",
  "label_field": "name",
  "is_system_schema": true,
  "fields": [
    {
      "is_system_field": true,
      "name": "disk",
      "label": "Disk",
      "type": "string",
			"optional": false
    },
    {
      "is_system_field": true,
      "name": "name",
      "label": "Name",
      "type": "string",
			"optional": false
    },
    {
      "is_system_field": true,
      "name": "path",
      "label": "Path",
      "type": "string",
      "optional": false
    },
    {
      "is_system_field": true,
      "name": "type",
      "label": "Type",
      "type": "string",
      "optional": false
    },
    {
      "is_system_field": true,
      "name": "size",
      "label": "Size",
      "type": "uint64",
      "optional": false
    },
    {
      "is_system_field": true,
      "name": "user",
      "label": "User",
      "type": "relation",
      "optional": true,
      "relation": {
        "type": "o2m",
        "schema": "user",
        "field": "medias",
        "owner": false
      }
    }
  ]
}`
