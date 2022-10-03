/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	cprompt "github.com/aschey/bubbleprompt-cobra"
	"github.com/spf13/cobra"
)

// interactiveCmd represents the interactive command
var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		model, err := cprompt.NewPrompt(cmd)
		if err != nil {
			return err
		}
		return model.Start()
	},
}

func init() {
	rootCmd.AddCommand(interactiveCmd)
}
