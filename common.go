/*
 * Copyright 2018. Akamai Technologies, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package akamai_cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/edgegrid"
	spnr "github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli"
)

var (
	spinner *spnr.Spinner
	App     *cli.App
)

type CommandLocator func() ([]cli.Command, error)

func CreateApp(command_name, usage, description, version string, defaultSection string, locator CommandLocator) {
	_, in_cli := os.LookupEnv("AKAMAI_CLI")

	app_name := "akamai"
	if command_name != "" {
		app_name = "akamai-" + command_name
		if in_cli {
			app_name = "akamai " + command_name
		}
	}

	App = cli.NewApp()
	App.Name = app_name
	App.HelpName = app_name
	App.Usage = usage
	App.Description = description
	App.Version = version
	App.Copyright = "Copyright (C) Akamai Technologies, Inc"
	App.Writer = colorable.NewColorableStdout()
	App.ErrWriter = colorable.NewColorableStderr()
	App.EnableBashCompletion = true

	if defaultSection != "" {
		dir, _ := homedir.Dir()
		dir += string(os.PathSeparator) + ".edgerc"
		App.Flags = []cli.Flag{
			cli.StringFlag{
				Name:   "edgerc",
				Usage:  "Location of the credentials file",
				Value:  dir,
				EnvVar: "AKAMAI_EDGERC",
			},
			cli.StringFlag{
				Name:   "section",
				Usage:  "Section of the credentials file",
				Value:  defaultSection,
				EnvVar: "AKAMAI_EDGERC_SECTION",
			},
			cli.StringFlag{
				Name:   "accountkey",
				Usage:  "Account switch key",
				EnvVar: "AKAMAI_ACCOUNTKEY",
			},
		}
	}

	App.BashComplete = DefaultAutoComplete

	commands, err := locator()
	if err != nil {
		fmt.Fprintln(App.ErrWriter, color.RedString("An error occurred initializing commands"))
	}

	if len(commands) > 0 {
		App.Commands = commands
	}

	cli.VersionFlag = cli.BoolFlag{
		Name:   "version",
		Hidden: true,
	}

	cli.BashCompletionFlag = cli.BoolFlag{
		Name:   "generate-auto-complete",
		Hidden: true,
	}
	cli.HelpFlag = cli.BoolFlag{
		Name:  "help",
		Usage: "show help",
	}

	setHelpTemplates()
}

func IsInteractive(c *cli.Context) bool {
	if !isTTY() {
		return false
	}

	if c != nil && c.IsSet("non-interactive") && c.Bool("non-interactive") {
		return false
	}

	return true
}

func isTTY() bool {
	if !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		return false
	}

	return true
}

func DefaultAutoComplete(ctx *cli.Context) {
	if ctx.Command.Name == "help" {
		var args []string
		args = append(args, os.Args[0])
		if len(os.Args) > 2 {
			args = append(args, os.Args[2:]...)
		}

		ctx.App.Run(args)
		return
	}

	commands := make([]cli.Command, 0)
	flags := make([]cli.Flag, 0)

	if ctx.Command.Name == "" {
		commands = ctx.App.Commands
		flags = ctx.App.Flags
	} else {
		if len(ctx.Command.Subcommands) != 0 {
			commands = ctx.Command.Subcommands
		}

		if len(ctx.Command.Flags) != 0 {
			flags = ctx.Command.Flags
		}
	}

	for _, command := range commands {
		if command.Hidden {
			continue
		}

		for _, name := range command.Names() {
			fmt.Fprintln(ctx.App.Writer, name)
		}
	}

	for _, flag := range flags {
	nextFlag:
		for _, name := range strings.Split(flag.GetName(), ",") {
			name = strings.TrimSpace(name)

			if name == cli.BashCompletionFlag.GetName() {
				continue
			}

			for _, arg := range os.Args {
				if arg == "--"+name || arg == "-"+name {
					continue nextFlag
				}
			}

			switch len(name) {
			case 0:
			case 1:
				fmt.Fprintln(ctx.App.Writer, "-"+name)
			default:
				fmt.Fprintln(ctx.App.Writer, "--"+name)
			}
		}
	}
}

func GetEdgegridConfig(c *cli.Context) (edgegrid.Config, error) {
	config, err := edgegrid.Init(c.GlobalString("edgerc"), c.GlobalString("section"))
	if err != nil {
		return edgegrid.Config{}, cli.NewExitError(err.Error(), 1)
	}

	if len(c.GlobalString("accountkey")) > 0 {
		config.AccountKey = c.GlobalString("accountkey")
	}

	return config, nil
}

func setHelpTemplates() {
	cli.AppHelpTemplate = "" +
		color.YellowString("Usage: \n") +
		color.BlueString("	{{if .UsageText}}"+
			"{{.UsageText}}"+
			"{{else}}"+
			"{{.HelpName}} "+
			"{{if .VisibleFlags}}[global flags]{{end}}"+
			"{{if .Commands}} command [command flags]{{end}} "+
			"{{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}"+
			"\n\n{{end}}") +

		"{{if .Description}}\n\n" +
		color.YellowString("Description:\n") +
		"   {{.Description}}" +
		"\n\n{{end}}" +

		"{{if .VisibleCommands}}" +
		color.YellowString("Built-In Commands:\n") +
		"{{range .VisibleCategories}}" +
		"{{if .Name}}" +
		"\n{{.Name}}\n" +
		"{{end}}" +
		"{{range .VisibleCommands}}" +
		color.GreenString("  {{.Name}}") +
		"{{if .Aliases}} ({{ $length := len .Aliases }}{{if eq $length 1}}alias:{{else}}aliases:{{end}} " +
		"{{range $index, $alias := .Aliases}}" +
		"{{if $index}}, {{end}}" +
		color.GreenString("{{$alias}}") +
		"{{end}}" +
		"){{end}}\n" +
		"{{end}}" +
		"{{end}}" +
		"{{end}}\n" +

		"{{if .VisibleFlags}}" +
		color.YellowString("Global Flags:\n") +
		"{{range $index, $option := .VisibleFlags}}" +
		"{{if $index}}\n{{end}}" +
		"   {{$option}}" +
		"{{end}}" +
		"\n\n{{end}}" +

		"{{if .Copyright}}" +
		color.HiBlackString("{{.Copyright}}") +
		"{{end}}\n"

	cli.CommandHelpTemplate = "" +
		color.YellowString("Name: \n") +
		"   {{.HelpName}}\n\n" +

		color.YellowString("Usage: \n") +
		color.BlueString("   {{.HelpName}}{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}\n\n") +

		"{{if .Category}}" +
		color.YellowString("Type: \n") +
		"   {{.Category}}\n\n{{end}}" +

		"{{if .Description}}" +
		color.YellowString("Description: \n") +
		"   {{.Description}}\n\n{{end}}" +

		"{{if .VisibleFlags}}" +
		color.YellowString("Flags: \n") +
		"{{range .VisibleFlags}}   {{.}}\n{{end}}{{end}}" +

		"{{if .UsageText}}{{.UsageText}}\n{{end}}"

	cli.SubcommandHelpTemplate = "" +
		color.YellowString("Name: \n") +
		"   {{.HelpName}} - {{.Usage}}\n\n" +

		color.YellowString("Usage: \n") +
		color.BlueString("   {{.HelpName}}{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}\n\n") +

		color.YellowString("Commands:\n") +
		"{{range .VisibleCategories}}" +
		"{{if .Name}}" +
		"{{.Name}}:" +
		"{{end}}" +
		"{{range .VisibleCommands}}" +
		`{{join .Names ", "}}{{"\t"}}{{.Usage}}` +
		"{{end}}\n\n" +
		"{{end}}" +

		"{{if .VisibleFlags}}" +
		color.YellowString("Flags:\n") +
		"{{range .VisibleFlags}}{{.}}\n{{end}}{{end}}"
}
