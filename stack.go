/*
 Copyright 2015 Stack Inc.
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

 http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package main

import (
	"errors"
	"fmt"
	"github.com/micosa/stack/cli"
	"github.com/spf13/cobra"
	"os"
	"runtime/debug"
	"strings"
)

var DispStackTrace bool = true
var ExitOnFailure bool = false
var StackVersion string = "1.0"
var StackRepo *cli.Repo
var StackLogLevel string = ""

func StackUsage(cmd *cobra.Command, err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}

	if DispStackTrace {
		debug.PrintStack()
	}

	if cmd != nil {
		cmd.Usage()
	}
	os.Exit(1)
}

func targetSetCmd(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		StackUsage(cmd,
			errors.New("Must specify two arguments (sect & k=v) to set"))
	}

	t, err := cli.LoadTarget(StackRepo, args[0])
	if err != nil {
		StackUsage(cmd, err)
	}
	ar := strings.Split(args[1], "=")

	t.Vars[ar[0]] = ar[1]

	err = t.Save()
	if err != nil {
		StackUsage(cmd, err)
	}

	fmt.Printf("Target %s successfully set %s to %s\n", args[0],
		ar[0], ar[1])
}

func targetShowCmd(cmd *cobra.Command, args []string) {
	dispSect := ""
	if len(args) == 1 {
		dispSect = args[0]
	}

	targets, err := cli.GetTargets(StackRepo)
	if err != nil {
		StackUsage(cmd, err)
	}

	for _, target := range targets {
		if dispSect == "" || dispSect == target.Vars["name"] {
			fmt.Println(target.Vars["name"])
			vars := target.GetVars()
			for k, v := range vars {
				fmt.Printf("	%s: %s\n", k, v)
			}
		}
	}
}

func targetCreateCmd(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		StackUsage(cmd, errors.New("Wrong number of args to create cmd."))
	}

	fmt.Println("Creating target " + args[0])

	if cli.TargetExists(StackRepo, args[0]) {
		StackUsage(cmd, errors.New(
			"Target already exists, cannot create target with same name."))
	}

	target := &cli.Target{
		Repo: StackRepo,
		Vars: map[string]string{},
	}
	target.Vars["name"] = args[0]
	target.Vars["arch"] = args[1]

	err := target.Save()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Target %s sucessfully created!\n", args[0])
	}
}

func targetBuildCmd(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		StackUsage(cmd, errors.New("Must specify target to build"))
	}

	t, err := cli.LoadTarget(StackRepo, args[0])
	if err != nil {
		StackUsage(cmd, err)
	}

	if len(args) > 1 && args[1] == "clean" {
		if len(args) > 2 && args[2] == "all" {
			err = t.BuildClean(true)
		} else {
			err = t.BuildClean(false)
		}
	} else {
		err = t.Build()
	}

	if err != nil {
		StackUsage(cmd, err)
	} else {
		fmt.Println("Successfully run!")
	}
}

func targetTestCmd(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		StackUsage(cmd, errors.New("Must specify target to build"))
	}

	t, err := cli.LoadTarget(StackRepo, args[0])
	if err != nil {
		StackUsage(cmd, err)
	}

	if len(args) > 1 && args[1] == "clean" {
		if len(args) > 2 && args[2] == "all" {
			err = t.Test("testclean", true)
		} else {
			err = t.Test("testclean", false)
		}
	} else {
		err = t.Test("test", ExitOnFailure)
	}

	if err != nil {
		StackUsage(cmd, err)
	} else {
		fmt.Println("Successfully run!")
	}
}

func targetAddCmds(base *cobra.Command) {
	targetCmd := &cobra.Command{
		Use:   "target",
		Short: "Set and view target information",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			var err error
			StackRepo, err = cli.NewRepo()
			if err != nil {
				StackUsage(nil, err)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}

	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Set target configuration variable",
		Run:   targetSetCmd,
	}

	targetCmd.AddCommand(setCmd)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a target",
		Run:   targetCreateCmd,
	}

	targetCmd.AddCommand(createCmd)

	showCmd := &cobra.Command{
		Use:   "show",
		Short: "View target configuration variables",
		Run:   targetShowCmd,
	}

	targetCmd.AddCommand(showCmd)

	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Build target",
		Run:   targetBuildCmd,
	}

	targetCmd.AddCommand(buildCmd)

	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Test target",
		Run:   targetTestCmd,
	}

	targetCmd.AddCommand(testCmd)

	base.AddCommand(targetCmd)
}

func repoCreateCmd(cmd *cobra.Command, args []string) {
	// must specify a repo name to create
	if len(args) != 1 {
		StackUsage(cmd, errors.New("Must specify a repo name to repo create"))
	}

	_, err := cli.CreateRepo(args[0])
	if err != nil {
		StackUsage(cmd, err)
	}

	fmt.Println("Repo " + args[0] + " successfully created!")
}

func repoAddCmds(baseCmd *cobra.Command) {
	repoCmd := &cobra.Command{
		Use:   "repo",
		Short: "Commands to manipulate the base repository",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a repository",
		Run:   repoCreateCmd,
	}

	repoCmd.AddCommand(createCmd)

	baseCmd.AddCommand(repoCmd)
}

func compilerCreateCmd(cmd *cobra.Command, args []string) {
	// must specify a compiler name to compiler create
	if len(args) != 1 {
		StackUsage(cmd, errors.New("Must specify a compiler name to compiler create"))
	}

	err := StackRepo.CreateCompiler(args[0])
	if err != nil {
		StackUsage(cmd, err)
	}

	fmt.Println("Compiler " + args[0] + " successfully created!")
}

func compilerInstallCmd(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		StackUsage(cmd, errors.New("Need to specify URL to install compiler "+
			"def from"))
	}

	var name string
	var err error

	if len(args) > 1 {
		name = args[1]
	} else {
		name, err = cli.UrlPath(args[0])
		if err != nil {
			StackUsage(cmd, err)
		}
	}

	dirName := StackRepo.BasePath + "/compiler/" + name + "/"
	if cli.NodeExist(dirName) {
		StackUsage(cmd, errors.New("Compiler "+name+" already installed."))
	}

	err = cli.CopyUrl(args[0], dirName)
	if err != nil {
		StackUsage(cmd, err)
	}

	fmt.Println("Compiler " + name + " successfully installed.")
}

func compilerAddCmds(baseCmd *cobra.Command) {
	compilerCmd := &cobra.Command{
		Use:   "compiler",
		Short: "Commands to install and create compiler definitions",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			var err error
			StackRepo, err = cli.NewRepo()
			if err != nil {
				StackUsage(nil, err)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new compiler definition",
		Run:   compilerCreateCmd,
	}

	compilerCmd.AddCommand(createCmd)

	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install a compiler from the specified URL",
		Run:   compilerInstallCmd,
	}

	compilerCmd.AddCommand(installCmd)

	baseCmd.AddCommand(compilerCmd)
}

func parseCmds() *cobra.Command {
	stackCmd := &cobra.Command{
		Use:   "stack",
		Short: "Stack is a tool to help you compose and build your own OS",
		Long: `Stack allows you to create your own embedded project based on the
		     stack operating system`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cli.Init(StackLogLevel)
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}

	stackCmd.PersistentFlags().StringVarP(&StackLogLevel, "loglevel", "l",
		"WARN", "Log level, defaults to WARN.")

	versCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the stack version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Stack version: ", StackVersion)
		},
	}

	stackCmd.AddCommand(versCmd)

	targetAddCmds(stackCmd)
	repoAddCmds(stackCmd)
	compilerAddCmds(stackCmd)

	return stackCmd
}

func main() {
	cmd := parseCmds()
	cmd.Execute()
}