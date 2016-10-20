package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/iron-io/functions_go"
	"github.com/urfave/cli"
)

type appsCmd struct {
	*functions.AppsApi
}

func apps() cli.Command {
	a := appsCmd{AppsApi: functions.NewAppsApi()}

	return cli.Command{
		Name:      "apps",
		Usage:     "list apps",
		ArgsUsage: "fnclt apps",
		Flags:     append(confFlags(&a.Configuration), []cli.Flag{}...),
		Action:    a.list,
		Subcommands: []cli.Command{
			{
				Name:   "create",
				Usage:  "create a new app",
				Action: a.create,
			},
		},
	}
}

func (a *appsCmd) list(c *cli.Context) error {
	resetBasePath(&a.Configuration)

	wrapper, _, err := a.AppsGet()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting app: %v", err)
		return nil // TODO return error instead?
	}

	if len(wrapper.Apps) == 0 {
		fmt.Println("no apps found")
	}

	for _, app := range wrapper.Apps {
		fmt.Println(app.Name)
	}

	return nil
}

func (a *appsCmd) create(c *cli.Context) error {
	if c.Args().First() == "" {
		return errors.New("error: app creating takes one argument, an app name")
	}

	resetBasePath(&a.Configuration)

	name := c.Args().Get(0)
	body := functions.AppWrapper{App: functions.App{Name: name}}
	wrapper, _, err := a.AppsPost(body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting app: %v", err)
		return nil // TODO return error instead?
	}

	fmt.Println(wrapper.App.Name, "created")
	return nil
}
