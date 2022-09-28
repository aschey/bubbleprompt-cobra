/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"time"

	cprompt "github.com/aschey/bubbleprompt-cobra"
	"github.com/aschey/bubbleprompt/executor"
	"github.com/spf13/cobra"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test <arg1> <arg2> [arg3]",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		model := executor.NewAsyncStringModel(func() (string, error) {
			time.Sleep(100 * time.Millisecond)
			return "done", nil
		})
		return cprompt.ExecModel(cmd, model)
	},
	Args: cobra.MinimumNArgs(2),
	ValidArgsFunction: func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {

		var choices []string
		completeIndex := len(args)
		if completeIndex == 0 {
			choices = cprompt.FilterShellCompletions([]string{"abc", "abcd"}, toComplete)
		} else if completeIndex == 1 {
			choices = cprompt.FilterShellCompletions([]string{"def", "defg"}, toComplete)
		}

		return choices, cobra.ShellCompDirectiveDefault
	},
}

func init() {
	testCmd.Flags().IntP("testInt", "f", 1, "f flag")
	testCmd.Flags().BoolP("testBool", "b", false, "b flag")
	testCmd.Flags().BoolP("testBool2", "c", false, "c flag")
	//testCmd.Flags().Lookup("testBool").NoOptDefVal = ""
	rootCmd.AddCommand(testCmd)
	//cprompt.SetPlaceholders(testCmd, "<arg1>", "<arg2>", "[arg3]")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// testCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// testCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
