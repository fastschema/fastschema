package main

import (
	"errors"

	"github.com/fastschema/fastschema"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
)

func main() {
	app := utils.Must(fastschema.New(&fs.Config{
		Port: "8000",
	}))

	// Log with different levels
	app.Logger().Info("Server configured successfully")

	app.Logger().Warn("A warning message")

	app.Logger().Error("An error message")

	app.Logger().Error(errors.New("an error message with error"))

	// Log with a context
	logger := app.Logger().WithContext(fs.Map{
		"request_id": "123",
	})

	logger.Info("A log message with context")
	//2024-06-04T12:36:02.649994461+07:00	info	logging/main.go:30	A log message with context	{"request_id": "123"}

	// Log within resource handler
	app.AddResource(fs.Get("/about", func(c fs.Context, _ any) (string, error) {
		c.Logger().Info("About page accessed")
		// 2024-06-04T12:37:43.76167048+07:00	info	logging/main.go:35	About page accessed	{"request_id": "cdb4653e-4663-4242-9228-49b69b1f5258"}
		return "About page", nil
	}, &fs.Meta{Public: true}))

	app.Logger().Fatal(app.Start())
}
