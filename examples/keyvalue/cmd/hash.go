package cmd

import (
	"examples/keyvalue/db"

	cprompt "github.com/aschey/bubbleprompt-cobra"
	"github.com/spf13/cobra"
)

var hashCmds = []*cobra.Command{
	{Use: "clear <key>", RunE: db.GetExecCommand("HClear")},
	{Use: "delete <key> <values...>", RunE: db.GetExecCommand("HDel")},
	{Use: "exists <key> [field]", RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 2 {
			return db.GetExecCommand("HExists")(cmd, args)
		}
		return db.GetExecCommand("HKeyExists")(cmd, args)
	}},
	{Use: "expire <key> <duration>", RunE: db.GetExecCommand("HExpire")},
	{Use: "fields <key>", RunE: db.GetExecCommand("HKeys")},
	{Use: "length <key>", RunE: db.GetExecCommand("HLen")},
	{Use: "set <key> <field> <value>", RunE: db.GetExecCommand("HSet")},
	{Use: "ttl <key>", RunE: db.GetExecCommand("HTTL")},
	{Use: "values <key>", RunE: db.GetExecCommand("HValues")},
}

var all *bool

var getCmd = &cobra.Command{
	Use: "get <key> [field]",
	RunE: func(cmd *cobra.Command, args []string) error {
		if *all {
			return db.GetExecCommand("HGetAll")(cmd, args)
		}
		return db.GetExecCommand("HGet")(cmd, args)
	}}

func init() {
	hashCmd.AddCommand(append(hashCmds, getCmd)...)
	all = getCmd.Flags().BoolP("all", "a", false, "Get all fields")
	cprompt.PreservePlaceholder(getCmd, "all")
}
