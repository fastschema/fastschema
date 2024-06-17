package main

import (
	"log"

	"github.com/fastschema/fastschema"
	"github.com/fastschema/fastschema/fs"
)

type Student struct {
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	Profile *Profile `json:"profile" fs.relation:"{'type':'o2o','schema':'profile','field':'student','owner':true}"`
}

type Profile struct {
	ID        int      `json:"id"`
	Address   string   `json:"address"`
	StudentID int      `json:"student_id"`
	Student   *Student `json:"student" fs.relation:"{'type':'o2o','schema':'student','field':'profile'}"`
}

type Post struct {
	ID         int         `json:"id"`
	Title      string      `json:"title"`
	Comments   []*Comment  `json:"comments" fs.relation:"{'type':'o2m','schema':'comment','field':'post','owner':true}"`
	Categories []*Category `json:"categories" fs.relation:"{'type':'m2m','schema':'category','field':'posts'}"`
}

type Comment struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
	PostID  int    `json:"post_id"`
	Post    *Post  `json:"post" fs.relation:"{'type':'o2o','schema':'post','field':'comments'}"`
}

type Category struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Posts []*Post `json:"posts" fs.relation:"{'type':'m2m','schema':'post','field':'categories'}"`
}

func main() {
	app, err := fastschema.New(&fs.Config{
		SystemSchemas: []any{
			Student{},
			Profile{},
			Post{},
			Comment{},
			Category{},
		},
	})
	if err != nil {
		panic(err)
	}

	log.Fatal(app.Start())
}
