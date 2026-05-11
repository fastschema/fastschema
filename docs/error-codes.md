# FastSchema Error Codes Reference

Canonical reference for all error codes emitted by the `schema` package. Codes use **dotted-lowercase** naming: `category.subject.specifier` for hierarchical filtering.

## How to consume

### Go (programmatic)

```go
import (
    "errors"
    "github.com/fastschema/fastschema/schema"
)

err := mySchema.Validate()
if err != nil {
    // Walk the typed batch
    var batch *schema.SchemaErrors
    if errors.As(err, &batch) {
        for _, fe := range batch.FieldErrors {
            fmt.Printf("[%s] field=%s message=%s\n", fe.Code, fe.Field, fe.Message)
        }
        // Filter by code
        if batch.HasCode(schema.CodeFieldTypeInvalid) {
            // ...
        }
    }

    // Walk individual entries via standard library Unwrap chain
    var fieldErr *schema.FieldError
    if errors.As(err, &fieldErr) {
        // first matching FieldError
    }
}
```

### REST / JSON

Validation errors at `POST /schema` and similar endpoints return HTTP 422 with structured payload:

```json
{
  "code": "422",
  "message": "schema validation failed",
  "data": {
    "schema": "post",
    "field_errors": [
      {
        "code": "field.type.invalid",
        "field": "status",
        "message": "field type 'enmu' is not recognized"
      },
      {
        "code": "schema.label_field.not_found",
        "message": "label_field 'foo' is not a string/text field; available: title, body"
      }
    ]
  }
}
```

Frontend / API consumers should:
1. Read `data.schema` for the affected schema name.
2. Iterate `data.field_errors[]` for per-field issues.
3. Map each `code` to a localized/help message via your own dictionary (codes are stable; messages may evolve).

### Type model

| Type | Use case | Schema context |
|------|----------|-----------------|
| `FieldError` | per-field error inside a `Schema.Validate()` batch | supplied by enclosing `SchemaErrors` |
| `SchemaError` | cross-schema, builder-level, or standalone error | explicit on the error |
| `SchemaErrors` | batch wrapper from `Schema.Validate()` | yes (`Schema` field) |
| `BuilderErrors` | aggregation across multiple schemas at build time | each entry sets its own |

`SchemaErrors.Unwrap() []error` enables stdlib `errors.As` traversal of individual entries.

## Codes by category

### `schema.*` — schema-level errors

| Code | Description |
|------|-------------|
| `schema.name.required` | schema is missing 'name' property |
| `schema.label_field.required` | schema is missing 'label_field' property |
| `schema.namespace.required` | schema is missing 'namespace' property |
| `schema.label_field.not_found` | `label_field` references a field that does not exist (or is not string/text); `Message` lists available alternatives |
| `schema.label_field.system_schema` | system schema (`user`, `role`, `file`) has a fixed `label_field`; assigning another value is invalid |
| `schema.primary_field.not_found` | declared `primary_field` references a field that does not exist |
| `schema.primary_field.required` | schema has no primary key field (no `id`, no `primary_field`) |
| `schema.io.read_error` | schema directory read failed (wraps underlying IO error in `Cause`) |
| `schema.init.unknown` | unrecognised error escaped a stage of `Schema.Init`; wraps the underlying error in `Cause` |

### `field.*` — field-level errors

| Code | Description |
|------|-------------|
| `field.name.required` | field at the given index is missing `name` |
| `field.type.invalid` | field declares a type identifier that is not recognized |
| `field.type.missing` | field has `TypeInvalid` (no type set) |
| `field.type.parse_error` | JSON unmarshalling encountered an unknown type identifier |
| `field.enum.required` | enum field is missing the `enums` array |
| `field.relation.required` | relation field is missing the `relation` object |
| `field.relation.schema.required` | `relation.schema` not set |
| `field.relation.type.required` | `relation.type` not set or invalid (must be `o2o`, `o2m`, or `m2m`) |
| `field.relation.field.required` | `relation.field` not set |
| `field.not_found` | field lookup by name failed within a known schema |
| `field.file.schema.required` | file field needs an owning schema name during initialization |
| `field.setter.compile_error` | field's `setter` expression failed to compile (wraps `expr.Compile` error in `Cause`) |
| `field.getter.compile_error` | field's `getter` expression failed to compile (wraps in `Cause`) |

### `relation.*` — cross-schema / FK errors

| Code | Description |
|------|-------------|
| `relation.target.not_found` | `relation.schema` references a schema that is not defined |
| `relation.back_ref.missing` | back-reference field not found on the target schema (or does not point back correctly) |
| `relation.fk.target.not_found` | foreign-key target field not found on target schema |
| `relation.fk.clone_failed` | foreign-key field clone failed (internal) |
| `relation.config.missing` | field has `type: relation` but no `relation` configuration |

### `builder.*` — builder-level errors

| Code | Description |
|------|-------------|
| `builder.schema.duplicate` | system schema declared more than once |
| `builder.schema.not_found` | requested schema not found in builder; `Message` lists available |
| `builder.schema.primary_key.missing` | schema is missing a primary-key field at build time |
| `builder.relation.not_m2m` | attempted to create an m2m junction for a non-m2m relation |
| `builder.junction_field.create_failed` | junction-field creation failed (internal) |

## Out of scope

`ErrInvalidFieldValue` (`schema/field.go`) is **not** part of this model. It represents runtime field-value validation (HTTP 400 user-input errors) rather than schema-definition validation. A separate `FieldValueError` type with HTTP 400 mapping may follow in a future release.

## See also

- [Schema validation rules](./schema.md#validation-rules)
- Source: `schema/errors.go`
