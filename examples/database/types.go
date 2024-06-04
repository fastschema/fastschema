package main

type Tag struct {
	ID    uint64  `json:"id"`
	Name  string  `json:"name" fs:"unique"`
	Desc  string  `json:"desc" fs:"optional"`
	Blogs []*Blog `json:"blogs" fs.relation:"{'type':'m2m','schema':'blog','field':'tags','owner':true}"`
}

type Blog struct {
	ID    uint64 `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body" fs:"optional;type=text" fs.renderer:"{'class':'editor'}"`
	Vote  int    `json:"vote" fs:"optional"`
	Tags  []*Tag `json:"tags" fs:"optional" fs.relation:"{'type':'m2m','schema':'tag','field':'blogs'}"`
}

type Payload struct {
	ID      int    `json:"id"`
	Comment string `json:"comment"`
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
