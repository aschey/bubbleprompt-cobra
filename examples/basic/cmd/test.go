/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"strings"
	"time"

	cprompt "github.com/aschey/bubbleprompt-cobra"
	"github.com/aschey/bubbleprompt/executor"
	"github.com/spf13/cobra"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		model := executor.NewAsyncStringModel(func() (string, error) {
			time.Sleep(100 * time.Millisecond)
			return "test", nil
		})
		return cprompt.ExecModel(cmd, model)
	},
	Args: cobra.MinimumNArgs(2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var choices []string
		if (len(args) == 0 && len(toComplete) == 0) || (len(args) == 1 && len(toComplete) > 0) {
			choices = []string{"abc", "abcd"}
		} else if (len(args) == 1 && len(toComplete) == 0) || (len(args) == 2 && len(toComplete) > 0) {
			choices = []string{"def", "defg"}
		}
		//  else if (len(args) == 2 && len(toComplete) == 0) || (len(args) == 3 && len(toComplete) > 0) {
		// 	choices = []string{"hij", "hijk"}
		// }

		filtered := []string{}
		for _, c := range choices {
			if strings.HasPrefix(c, toComplete) {
				filtered = append(filtered, c)
			}
		}

		return filtered, cobra.ShellCompDirectiveDefault
	},
}

func init() {
	testCmd.Flags().IntP("testInt", "f", 1, "f flag")
	testCmd.Flags().BoolP("testBool", "b", false, "b flag")
	testCmd.Flags().BoolP("testBool2", "c", false, "c flag")
	//testCmd.Flags().Lookup("testBool").NoOptDefVal = ""
	rootCmd.AddCommand(testCmd)
	cprompt.SetPlaceholders(testCmd, "<arg1>", "<arg2>", "[arg3]")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// testCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// testCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
