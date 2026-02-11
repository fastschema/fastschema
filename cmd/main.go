package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/fastschema/fastschema"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
	toolservice "github.com/fastschema/fastschema/services/tool"
	"github.com/google/uuid"
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
					app := utils.Must(fastschema.New(&fs.Config{
						Dir: c.Args().Get(0),
					}))

					return toolservice.Setup(
						c.Context,
						app.DB(),
						app.Logger(),
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
						Usage:   "Admin user id (UUID)",
					},
				},
				Action: func(c *cli.Context) error {
					app := utils.Must(fastschema.New(&fs.Config{
						Dir: c.Args().Get(0),
					}))

					idStr := c.String("id")
					if idStr == "" {
						return fmt.Errorf("admin user id is required")
					}

					id, err := uuid.Parse(idStr)
					if err != nil {
						return fmt.Errorf("invalid admin user id: %w", err)
					}

					return toolservice.ResetAdminPassword(
						c.Context,
						app.DB(),
						c.String("password"),
						id,
					)
				},
			},
			{
				Name:  "migration",
				Usage: "Manage database migrations",
				Subcommands: []*cli.Command{
					{
						Name:  "create",
						Usage: "Create empty migration files for custom SQL",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "name",
								Aliases:  []string{"n"},
								Usage:    "Migration name",
								Required: true,
							},
						},
						Action: func(c *cli.Context) error {
							app := utils.Must(fastschema.New(&fs.Config{
								Dir: c.Args().Get(0),
							}))

							_, err := toolservice.MigrationNew(
								c.Context,
								app.DB(),
								c.String("name"),
							)
							return err
						},
					},
					{
						Name:  "generate",
						Usage: "Generate migration from schema diff",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "name",
								Aliases:  []string{"n"},
								Usage:    "Migration name",
								Required: true,
							},
						},
						Action: func(c *cli.Context) error {
							app := utils.Must(fastschema.New(&fs.Config{
								Dir: c.Args().Get(0),
							}))

							return toolservice.MigrationGenerate(
								c.Context,
								app.DB(),
								c.String("name"),
							)
						},
					},
					{
						Name:  "up",
						Usage: "Apply pending migrations",
						Flags: []cli.Flag{
							&cli.IntFlag{
								Name:    "count",
								Aliases: []string{"c"},
								Usage:   "Number of migrations to apply (0 = all)",
								Value:   0,
							},
						},
						Action: func(c *cli.Context) error {
							app := utils.Must(fastschema.New(&fs.Config{
								Dir: c.Args().Get(0),
							}))

							_, err := toolservice.MigrationUp(
								c.Context,
								app.DB(),
								c.Int("count"),
							)
							return err
						},
					},
					{
						Name:  "down",
						Usage: "Roll back migrations",
						Flags: []cli.Flag{
							&cli.IntFlag{
								Name:    "count",
								Aliases: []string{"c"},
								Usage:   "Number of migrations to roll back",
								Value:   1,
							},
						},
						Action: func(c *cli.Context) error {
							app := utils.Must(fastschema.New(&fs.Config{
								Dir: c.Args().Get(0),
							}))

							_, err := toolservice.MigrationDown(
								c.Context,
								app.DB(),
								c.Int("count"),
							)
							return err
						},
					},
					{
						Name:  "status",
						Usage: "Show migration status",
						Action: func(c *cli.Context) error {
							app := utils.Must(fastschema.New(&fs.Config{
								Dir: c.Args().Get(0),
							}))

							_, _, err := toolservice.MigrationStatus(
								c.Context,
								app.DB(),
							)
							return err
						},
					},
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
