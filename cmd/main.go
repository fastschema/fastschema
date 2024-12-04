package main

import (
	"os"
	"sort"

	"github.com/fastschema/fastschema"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
	toolservice "github.com/fastschema/fastschema/services/tool"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "fastschema",
		Usage: "BaaS",
		Commands: []*cli.Command{
			{
				Name:  "setup",
				Usage: "Setup the fastschema application",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "username",
						Aliases:  []string{"u"},
						Usage:    "Admin username",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "email",
						Aliases:  []string{"e"},
						Usage:    "Admin email",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "password",
						Aliases:  []string{"p"},
						Usage:    "Admin password",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					fs := utils.Must(fastschema.New(&fs.Config{
						Dir: c.Args().Get(0),
					}))

					return toolservice.Setup(
						c.Context,
						fs.DB(),
						fs.Logger(),
						c.String("username"), c.String("email"), c.String("password"),
					)
				},
			},
			{
				Name:  "start",
				Usage: "Start the fastschema application",
				Action: func(c *cli.Context) error {
					app := utils.Must(fastschema.New(&fs.Config{
						Dir: c.Args().Get(0),
					}))

					return app.Start()
				},
			},
			{
				Name:  "reset-admin-password",
				Usage: "Reset the admin password",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "password",
						Aliases:  []string{"p"},
						Usage:    "Admin new password",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "id",
						Aliases: []string{"i"},
						Usage:   "Admin user id",
						Value:   "1",
					},
				},
				Action: func(c *cli.Context) error {
					fs := utils.Must(fastschema.New(&fs.Config{
						Dir: c.Args().Get(0),
					}))

					_ = toolservice.ResetAdminPassword(
						c.Context,
						fs.DB(),
						c.String("password"),
						c.Int("id"),
					)

					return nil
				},
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
