package cmd

import (
	"examples/keyvalue/db"
	"fmt"

	"github.com/spf13/cobra"
)

var baseCmds = []*cobra.Command{
	{Use: "delete <key>", RunE: db.GetExecCommand("Del"), ValidArgsFunction: db.GetKeys},
	{Use: "exists <key>", RunE: db.GetExecCommand("Exists"), ValidArgsFunction: db.GetKeys},
	{Use: "expire <key> <duration>", RunE: db.GetExecCommand("Expire"), ValidArgsFunction: db.GetKeys},
	{Use: "get <key>", RunE: db.GetExecCommand("Get"), ValidArgsFunction: db.GetKeys},
	{Use: "ttl <key>", RunE: db.GetExecCommand("TTL"), ValidArgsFunction: db.GetKeys},
}

var hashCmd = &cobra.Command{Use: "hash <subcommand>"}
var setCmd = &cobra.Command{Use: "set <subcommand>"}
var zsetCmd = &cobra.Command{Use: "zset <subcommand>"}

var setKeyCmd = &cobra.Command{Use: "set-key <key> <value> [flags]", RunE: func(cmd *cobra.Command, args []string) error {
	if ttl, _ := cmd.Flags().GetInt64("ttl"); ttl > -1 {
		return db.GetExecCommand("SetEx")(cmd, append(args, fmt.Sprintf("%d", ttl)))
	}
	return db.GetExecCommand("Set")(cmd, args)
}}

func init() {
	rootCmd.AddCommand(append(baseCmds, setKeyCmd, hashCmd, setCmd, zsetCmd)...)

	setKeyCmd.Flags().Int64P("ttl", "t", -1, "Key TTL")
}
