package main

import (
	"log"
	"net/http"

	"github.com/fastschema/fastschema"
	"github.com/fastschema/fastschema/fs"
)

func main() {
	app, err := fastschema.New(&fs.Config{})
	if err != nil {
		panic(err)
	}

	// Add a private resource under /api/stats
	app.API().Add(fs.Get("/stats", func(c fs.Context, _ any) (any, error) {
		return fs.Map{"message": "Stats"}, nil
	}))

	// Add a public resource under /api that will be served at /api/hello
	app.API().Add(fs.Get("/hello", func(c fs.Context, _ any) (any, error) {
		return fs.Map{"message": "Hello, World!"}, nil
	}, &fs.Meta{Public: true}))

	// Add a public resource that will response a HTML page at /about
	app.AddResource(fs.Get("/about", func(c fs.Context, _ any) (any, error) {
		header := make(http.Header)
		header.Set("Content-Type", "text/html")

		return &fs.HTTPResponse{
			StatusCode: http.StatusOK,
			Header:     header,
			Body: []byte(`<!DOCTYPE html><html>
				<head>
					<title>About</title>
				</head>
				<body>
					<h1>About</h1>
					<p>This is the about page.</p>
				</body>
			</html>`),
		}, nil
	}, &fs.Meta{Public: true}))

	// Add a group resource under /guide
	guide := app.Resources().Group("guide", &fs.Meta{Prefix: "/guide"})

	// Add a public resource under /guide that will be served at /guide/getting-started
	guide.Add(fs.Get("/getting-started", func(c fs.Context, _ any) (any, error) {
		return fs.Map{"message": "Getting started with Fastschema"}, nil
	}, &fs.Meta{Public: true}))

	// Add a public resource that has a required argument under /blog/:name
	app.API().Add(fs.Get("/blog/:name", func(c fs.Context, args any) (any, error) {
		blog := c.Arg("name")
		return fs.Map{"message": "Blog: " + blog}, nil
	}, &fs.Meta{
		Public: true,
		Args: fs.Args{
			"name": fs.Arg{
				Type:        fs.TypeString,
				Required:    true,
				Description: "The name of the blog",
				Example:     "my-first-blog",
			},
		},
	}))

	type BlogCreate struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}

	type BlogLink struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Link string `json:"link"`
	}

	// Add a private resource that accepts JSON data under /blog
	app.API().Add(fs.Post("/blog", func(c fs.Context, data *BlogCreate) (*BlogLink, error) {
		// data is an object of BlogCreate that contains the JSON data sent by the client
		// You can use the data to create a new blog
		// For now, we will just return the data sent by the client
		return &BlogLink{
			ID:   1,
			Name: data.Name,
			Link: "/blog/" + data.Name,
		}, nil
	}))

	log.Fatal(app.Start())
}
