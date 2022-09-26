package cmd

import (
	"examples/keyvalue/db"
	"fmt"

	"github.com/spf13/cobra"
)

var baseCmds = []*cobra.Command{
	{Use: "delete <key>", RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	}},
	{Use: "exists <key>", RunE: db.GetExecCommand("Exists")},
	{Use: "expire <key> <duration>", RunE: db.GetExecCommand("Expire")},
	{Use: "get <key>", RunE: db.GetExecCommand("Get")},
	{Use: "set <subcommand>"},
	{Use: "ttl <key>", RunE: db.GetExecCommand("TTL")},
	{Use: "zset <subcommand>"},
}

var ttl *int64

var hashCmd = &cobra.Command{Use: "hash <subcommand>"}

var setKeyCmd = &cobra.Command{Use: "set-key <key> <value> [flags]", RunE: func(cmd *cobra.Command, args []string) error {
	if *ttl > -1 {
		return db.GetExecCommand("SetEx")(cmd, append(args, fmt.Sprintf("%d", *ttl)))
	}
	return db.GetExecCommand("Set")(cmd, args)
}}

func init() {
	rootCmd.AddCommand(append(baseCmds, setKeyCmd, hashCmd)...)

	ttl = setKeyCmd.Flags().Int64P("ttl", "t", -1, "Key TTL")
}
