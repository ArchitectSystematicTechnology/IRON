package main

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"strings"

	vers "github.com/iron-io/functions/api/version"
	functions "github.com/iron-io/functions_go"
	"github.com/urfave/cli"
	"log"
)

var aliases = map[string]cli.Command{
	"build":  build(),
	"bump":   bump(),
	"deploy": deploy(),
	"push":   push(),
	"run":    run(),
	"call":   call(),
}

var API_VERSION = "/v1"
var SSL_SKIP_VERIFY = (os.Getenv("SSL_SKIP_VERIFY") == "true")
var API_URL = "http://localhost:8080"
var SCHEME = "http"
var HOST string
var BASE_PATH string

func getBasePath(version string) string {
	u, err := url.Parse(API_URL)
	if err != nil {
		log.Fatalln("Couldn't parse API URL:", err)
	}
	HOST = u.Host
	SCHEME = u.Scheme
	u.Path = version
	return u.String()
}

func init() {
	if os.Getenv("API_URL") != "" {
		API_URL = os.Getenv("API_URL")
	}
	BASE_PATH = getBasePath(API_VERSION)
}

func main() {
	app := newFn()
	app.Run(os.Args)
}

func resetBasePath(c *functions.Configuration) error {
	c.BasePath = BASE_PATH
	return nil
}

func aliasesFn() []cli.Command {
	cmds := []cli.Command{}
	for alias, cmd := range aliases {
		cmd.Name = alias
		cmd.Hidden = true
		cmds = append(cmds, cmd)
	}
	return cmds
}

func newFn() *cli.App {
	app := cli.NewApp()
	app.Name = "fn"
	app.Version = vers.Version
	app.Authors = []cli.Author{{Name: "iron.io"}}
	app.Description = "IronFunctions command line tools"
	app.UsageText = `Check the manual at https://github.com/iron-io/functions/blob/master/fn/README.md`

	cli.AppHelpTemplate = `{{.Name}} {{.Version}}{{if .Description}}

{{.Description}}{{end}}

USAGE:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}

ENVIRONMENT VARIABLES:
   API_URL - IronFunctions remote API address{{if .VisibleCommands}}

COMMANDS:{{range .VisibleCategories}}{{if .Name}}
   {{.Name}}:{{end}}{{range .VisibleCommands}}
     {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{end}}{{if .VisibleFlags}}

ALIASES:
     build    (images build)
     bump     (images bump)
     deploy   (images deploy)
     run      (images run)
     call     (routes call)
     push     (images push)

GLOBAL OPTIONS:
   {{range $index, $option := .VisibleFlags}}{{if $index}}
   {{end}}{{$option}}{{end}}{{end}}
`

	app.CommandNotFound = func(c *cli.Context, cmd string) {
		fmt.Fprintf(os.Stderr, "command not found: %v\n", cmd)
	}
	app.Commands = []cli.Command{
		initFn(),
		apps(),
		routes(),
		images(),
		lambda(),
		version(),
	}
	app.Commands = append(app.Commands, aliasesFn()...)

	prepareCmdArgsValidation(app.Commands)

	return app
}

func parseArgs(c *cli.Context) ([]string, []string) {
	args := strings.Split(c.Command.ArgsUsage, " ")
	var reqArgs []string
	var optArgs []string
	for _, arg := range args {
		if strings.HasPrefix(arg, "[") {
			optArgs = append(optArgs, arg)
		} else if strings.Trim(arg, " ") != "" {
			reqArgs = append(reqArgs, arg)
		}
	}
	return reqArgs, optArgs
}

func prepareCmdArgsValidation(cmds []cli.Command) {
	// TODO: refactor fn to use urfave/cli.v2
	// v1 doesn't let us validate args before the cmd.Action

	for i, cmd := range cmds {
		prepareCmdArgsValidation(cmd.Subcommands)
		if cmd.Action == nil {
			continue
		}
		action := cmd.Action
		cmd.Action = func(c *cli.Context) error {
			reqArgs, _ := parseArgs(c)
			if c.NArg() < len(reqArgs) {
				var help bytes.Buffer
				cli.HelpPrinter(&help, cli.CommandHelpTemplate, c.Command)
				return fmt.Errorf("ERROR: Missing required arguments: %s\n\n%s", strings.Join(reqArgs[c.NArg():], " "), help.String())
			}
			return cli.HandleAction(action, c)
		}
		cmds[i] = cmd
	}
}
