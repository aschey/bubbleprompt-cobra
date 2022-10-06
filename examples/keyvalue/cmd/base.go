package cmd

import (
	"examples/keyvalue/db"
	"fmt"

	"github.com/spf13/cobra"
)

var baseCmds = []*cobra.Command{
	{
		Use:               "delete <key>",
		Short:             "Delete the key from the database",
		RunE:              db.GetExecCommand("Delete"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.GetKeys,
	},
	{
		Use:               "exists <key>",
		Short:             "Check if the key exists in the database",
		RunE:              db.GetExecCommand("Exists"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.GetKeys,
	},
	{
		Use:               "expire <key> <duration>",
		Short:             "Expire the key after the specified duration",
		RunE:              db.GetExecCommand("Expire"),
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: db.GetKeys,
	},
	{
		Use:               "get <key>",
		Short:             "Get the key from the database",
		RunE:              db.GetExecCommand("Get"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.GetKeys,
	},
	{
		Use:               "ttl <key>",
		Short:             "Check the time-to-live of the key",
		RunE:              db.GetExecCommand("TTL"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.GetKeys,
	},
	setKeyCmd,
	hashCmd,
	setCmd,
	zsetCmd,
}

var hashCmd = &cobra.Command{
	Use:   "hash <command>",
	Short: "Hash store operations",
	Args:  cobra.MinimumNArgs(1),
}
var setCmd = &cobra.Command{
	Use:   "set <command>",
	Short: "Set store operations",
	Args:  cobra.MinimumNArgs(1),
}
var zsetCmd = &cobra.Command{
	Use:   "zset <subcommand>",
	Short: "ZSet operations",
	Args:  cobra.MinimumNArgs(1),
}
var setKeyCmd = &cobra.Command{
	Use:   "set-key <key> <value> [flags]",
	Short: "Set the value for the key",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if ttl, _ := cmd.Flags().GetInt64("ttl"); ttl > -1 {
			return db.GetExecCommand("SetEx")(cmd, append(args, fmt.Sprintf("%d", ttl)))
		}
		return db.GetExecCommand("Set")(cmd, args)
	}}

func init() {
	rootCmd.AddCommand(baseCmds...)

	setKeyCmd.Flags().Int64P("ttl", "t", -1, "Set the time-to-live for the key")
}
