<p align="center">
  <a href="https://fastschema.com">
    <picture>
      <source height="125" media="(prefers-color-scheme: light)" srcset="https://fastschema.com/fastschema-logo-small.svg">
      <img height="125" alt="FastSchema" src="https://fastschema.com/fastschema-logo-small.svg">
    </picture>
  </a>
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/fastschema/fastschema#section-readme" target="_blank" rel="noopener">
    <img src="https://img.shields.io/badge/go.dev-reference-blue?logo=go&logoColor=white" alt="Go.Dev reference" />
  </a>
  <a href="https://goreportcard.com/report/github.com/fastschema/fastschema" target="_blank" rel="noopener">
    <img src="https://goreportcard.com/badge/github.com/fastschema/fastschema" alt="go report card" />
  </a>
  <a href="https://codecov.io/gh/fastschema/fastschema" target="_blank" rel="noopener">
    <img src="https://codecov.io/gh/fastschema/fastschema/graph/badge.svg?token=TPU5QN6E4Z" alt="codecov" />
  </a>
  <a href="https://github.com/fastschema/fastschema/actions" target="_blank" rel="noopener">
    <img src="https://github.com/fastschema/fastschema/actions/workflows/ci.yml/badge.svg" alt="test status" />
  </a>
  <a href="https://opensource.org/licenses/MIT" target="_blank" rel="noopener">
    <img src="https://img.shields.io/badge/license-MIT-brightgreen.svg" alt="MIT license" />
  </a>
</p>

_FastSchema is a Go web framework and headless CMS for building dynamic web applications. Designed to simplify the creation and management of structured content, FastSchema automates the generation of databases and provides CRUD APIs effortlessly._

## Try it out

Launch a headless CMS in seconds or utilize as a web framework.

### Run the Docker Container:

```bash
docker pull ghcr.io/fastschema/fastschema:latest
docker run \
  -p 8000:8000 \
  -v ./data:/fastschema/data \
  ghcr.io/fastschema/fastschema:latest
```

**Example output:**

```
> APP_KEY is not set. A new key is generated and saved to /fastschema/data/.env
> Using the default sqlite db file path: /fastschema/data/fastschema.db
> Visit the following URL to setup the app: http://localhost:8000/dash/setup/?token=lUDRgoTUUNDsjCcitgGFTqwMZQPmYvlU
```

Now you can access to the FastSchema setup page by visiting [http://localhost:8000/setup?token=\{token\}](http://localhost:8000?token={token}) (The setup token is displayed in the terminal).

> **Note:** FastSchema is currently in beta and under active development. We welcome feedback, contributions, and suggestions from the community to help improve the platform and make it more robust and feature-rich.

## Overview

FastSchema core features are built on top of `schema`, a blueprint that outlines the structure of your content. This schema acts as the foundation upon which FastSchema builds your database tables and API endpoints, streamlining the development process and allowing you to focus on creating rich, dynamic content.

<p style="text-align: center;">
  <img src="https://fastschema.com/static/images/fastschema.png" alt="FastSchema Overview" />
</p>

## Use Cases

- **A Headless CMS (No-Code Solution)**

  FastSchema is an ideal solution for building headless CMS applications that require dynamic content modeling without writing a line of code.

  It is designed to support API-first development, allowing you to define your content models and generate RESTful APIs effortlessly.

  With two line of commands, you can create a new project, define your content models, and start creating content instantly.

- **A Framework for web development**

  FastSchema is designed to be used as a framework for building web applications. It provides a set of tools and packages that simplify the development process.

  `Resource` is a core concept that represents an access point to your data. By defining resources, you can create custom endpoints and customize the behavior of your APIs.

  `Hooks` are functions that are executed before or after an operation is performed on a resource. A `Hook` can be a `resolver hook`, `database hook`, or `application hook`. They allow you to add custom logic to your APIs and extend the functionality of FastSchema.

  `ORM` is a powerful tool that simplifies the interaction with your database. It provides a set of methods for querying, creating, updating, and deleting records in your database.

## Web Framework

```go
package main

import (
	"fmt"
	"log"

	"github.com/fastschema/fastschema"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
)

func main() {
	app, _ := fastschema.New(&fs.Config{
		SystemSchemas: []any{Tag{}, Blog{}},
	})

	app.API().Add(fs.Post("/blogvote", func(c fs.Context, vote *Payload) (*Response, error) {
		_, err := db.Mutation[Blog](app.DB()).
			Where(db.EQ("id", vote.ID)).
			Update(c.Context(), fs.Map{
				"$expr": fs.Map{"vote": "vote + 1"},
			})

		return &Response{
			Success: err == nil,
			Message: fmt.Sprintf("Vote for %d: %v", vote.ID, err),
		}, nil
	}))

	log.Fatal(app.Start())
}
```

## Features

Fastschema offers a comprehensive suite of features designed to streamline and simplify the process of building and managing dynamic web applications.

- **Automated Database Generation:** FastSchema automatically generates the necessary database tables based on your schema definition with flexible relationships model, eliminating the need for manual setup.

- **RESTful API Generation:** RESTful APIs are automatically generated based on the schema definition. Whenever you create or update a schema, the corresponding API endpoints are updated accordingly.

- **Dynamic Content Modeling:** Easily create and modify content models through the intuitive admin UI, with changes reflected instantly in the schema definition file.

- **Built-in File Management:** FastSchema provides a built-in file manager to manage media assets, enabling you to upload, organize, and serve files seamlessly.

- **Built-in Admin Control Panel:** FastSchema comes with a built-in admin control panel that allows you to manage content, users, manage permissions, and more.

- **Database Support:** MySQL, PostgreSQL, SQLite.

- **Role-Based Access Control:** Define roles and permissions to control access to content and features.

- **OpenAPI Specification (OAS) Generation:** FastSchema automatically generates OpenAPI Specification (OAS) documentation for your APIs, making it easy to understand and consume your APIs.

- **Extensible and Flexible:** Extend and customize FastSchema with Go code, build extensive features by leveraging the power of Resources, Hooks, ORM, and more.

## Documentation

For more information on how to get started with FastSchema, check out our [documentation](https://fastschema.com).


## Roadmap

- [x] Improve documentation and testing.
- [ ] Add auth provider.
- [ ] Plugin system.
- [x] OpenAPI generator.
- [ ] Real-time updates.
- [ ] GraphQL support.
- [ ] Webhooks.
- [ ] Client SDKs.

## Testing

FastSchema comes with a suite of automated tests to ensure the stability and reliability of the platform.

_Fastschema come with integration tests that require a database connection. You can use the following command to create DB containers._

```bash
cd tests/integration
docker compose up -d
```

To run the tests, execute the following command:

```bash
go test ./...
```

You can skip the integration tests by running tests for packages only.

```bash
./tests/test.sh ./schema
```

## Dependencies

FastSchema is built using the Go programming language and leverages a number of open-source libraries to provide its core functionality. Some of the key dependencies include:

- [Fiber](https://gofiber.io/)
- [Ent](https://entgo.io/)
- [Rclone](https://rclone.org/)
- [Zap](https://pkg.go.dev/go.uber.org/zap)
- [Next.js](https://nextjs.org/)
- [Shadcn](https://ui.shadcn.com/)
- [TipTap](https://www.tiptap.dev/)

## Contributing

We welcome contributions from the community and encourage developers to get involved in the project. Whether you're a seasoned developer or just getting started, there are plenty of ways to contribute to FastSchema.
