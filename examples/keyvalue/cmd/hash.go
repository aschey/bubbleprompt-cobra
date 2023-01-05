package cmd

import (
	"examples/keyvalue/db"

	cprompt "github.com/aschey/bubbleprompt-cobra"
	"github.com/spf13/cobra"
)

var hashCmds = []*cobra.Command{
	{
		Use:               "clear <key>",
		Short:             "Remove all the values from the key",
		RunE:              db.GetExecCommand("HClear"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.HGetKeys,
	},
	{
		Use:               "delete <key> <values...>",
		Short:             "Delete the values from the key",
		RunE:              db.GetExecCommand("HDel"),
		Args:              cobra.MinimumNArgs(2),
		ValidArgsFunction: db.HGetKeys,
	},
	{
		Use:   "exists <key> [field]",
		Short: "Check if the key exists",
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
		Short:             "Expire the key after the specified duration",
		RunE:              db.GetExecCommand("HExpire"),
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: db.HGetKeys,
	},
	{
		Use:               "fields <key>",
		Short:             "Get all the fields from the key",
		RunE:              db.GetListExecCommand("HFields"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.HGetKeys,
	},
	{
		Use:               "length <key>",
		Short:             "Check how many values are in the key",
		RunE:              db.GetExecCommand("HLen"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.HGetKeys,
	},
	{
		Use:   "set <key> <field> <value>",
		Short: "Set the values for the key",
		RunE:  db.GetExecCommand("HSet"),
		Args:  cobra.ExactArgs(3),
	},
	{
		Use:               "ttl <key>",
		Short:             "Check the time-to-live for the key",
		RunE:              db.GetExecCommand("HTTL"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.HGetKeys},
	{
		Use:               "values <key>",
		Short:             "Get all the values for the key",
		RunE:              db.GetListExecCommand("HVals"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.HGetKeys,
	},
	getCmd,
}

var getCmd = &cobra.Command{
	Use:               "get <key> <field|-a>",
	Short:             "Get the values for the key",
	ValidArgsFunction: db.HGetKeys,
	RunE: func(cmd *cobra.Command, args []string) error {
		if all, _ := cmd.Flags().GetBool("all"); all {
			return db.GetExecCommand("HGetAll")(cmd, args)
		}
		return db.GetExecCommand("HGet")(cmd, args)
	},
}

func init() {
	hashCmd.AddCommand(hashCmds...)
	getCmd.Flags().BoolP("all", "a", false, "Get all fields")
	//cprompt.PreservePlaceholder(getCmd, "all")
	cprompt.ShowFlagPlaceholder(getCmd, false)
}
