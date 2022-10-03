package cmd

import (
	"examples/keyvalue/db"

	cprompt "github.com/aschey/bubbleprompt-cobra"
	"github.com/spf13/cobra"
)

var hashCmds = []*cobra.Command{
	{
		Use:               "clear <key>",
		RunE:              db.GetExecCommand("HClear"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.HGetKeys,
	},
	{
		Use:               "delete <key> <values...>",
		RunE:              db.GetExecCommand("HDel"),
		Args:              cobra.MinimumNArgs(2),
		ValidArgsFunction: db.HGetKeys,
	},
	{
		Use: "exists <key> [field]",
		RunE: func(cmd *cobra.Command,
			args []string) error {
			if len(args) == 2 {
				return db.GetExecCommand("HExists")(cmd, args)
			}
			return db.GetExecCommand("HKeyExists")(cmd, args)
		},
		Args:              cobra.RangeArgs(1, 2),
		ValidArgsFunction: db.HGetKeys,
	},
	{
		Use:               "expire <key> <duration>",
		RunE:              db.GetExecCommand("HExpire"),
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: db.HGetKeys,
	},
	{
		Use:               "fields <key>",
		RunE:              db.GetListExecCommand("HFields"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.HGetKeys},

	{
		Use:               "length <key>",
		RunE:              db.GetExecCommand("HLen"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.HGetKeys,
	},
	{
		Use:  "set <key> <field> <value>",
		RunE: db.GetExecCommand("HSet"),
		Args: cobra.ExactArgs(3),
	},
	{
		Use:               "ttl <key>",
		RunE:              db.GetExecCommand("HTTL"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.HGetKeys},
	{
		Use:               "values <key>",
		RunE:              db.GetListExecCommand("HVals"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.HGetKeys,
	},
}

var getCmd = &cobra.Command{
	Use:  "get <key> [field]",
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if all, _ := cmd.Flags().GetBool("all"); all {
			return db.GetExecCommand("HGetAll")(cmd, args)
		}
		return db.GetExecCommand("HGet")(cmd, args)
	}}

func init() {
	hashCmd.AddCommand(append(hashCmds, getCmd)...)
	getCmd.Flags().BoolP("all", "a", false, "Get all fields")
	cprompt.PreservePlaceholder(getCmd, "all")
}
