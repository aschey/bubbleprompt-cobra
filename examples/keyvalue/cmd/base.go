package cmd

import (
	"examples/keyvalue/db"
	"fmt"

	"github.com/spf13/cobra"
)

var baseCmds = []*cobra.Command{
	{
		Use:               "delete <key>",
		RunE:              db.GetExecCommand("Del"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.GetKeys,
	},
	{
		Use:               "exists <key>",
		RunE:              db.GetExecCommand("Exists"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.GetKeys,
	},
	{
		Use:               "expire <key> <duration>",
		RunE:              db.GetExecCommand("Expire"),
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: db.GetKeys,
	},
	{
		Use:               "get <key>",
		RunE:              db.GetExecCommand("Get"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.GetKeys,
	},
	{
		Use:               "ttl <key>",
		RunE:              db.GetExecCommand("TTL"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.GetKeys,
	},
}

var hashCmd = &cobra.Command{Use: "hash <subcommand>", Args: cobra.MinimumNArgs(1)}
var setCmd = &cobra.Command{Use: "set <subcommand>", Args: cobra.MinimumNArgs(1)}
var zsetCmd = &cobra.Command{Use: "zset <subcommand>", Args: cobra.MinimumNArgs(1)}

var setKeyCmd = &cobra.Command{Use: "set-key <key> <value> [flags]", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
	if ttl, _ := cmd.Flags().GetInt64("ttl"); ttl > -1 {
		return db.GetExecCommand("SetEx")(cmd, append(args, fmt.Sprintf("%d", ttl)))
	}
	return db.GetExecCommand("Set")(cmd, args)
}}

func init() {
	rootCmd.AddCommand(append(baseCmds, setKeyCmd, hashCmd, setCmd, zsetCmd)...)

	setKeyCmd.Flags().Int64P("ttl", "t", -1, "Key TTL")
}
