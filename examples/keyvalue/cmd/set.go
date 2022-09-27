package cmd

import (
	"examples/keyvalue/db"

	"github.com/spf13/cobra"
)

var setCommands = []*cobra.Command{
	{Use: "add <key> <members...>", RunE: db.GetExecCommand("SAdd")},
	{Use: "card <key>", RunE: db.GetExecCommand("SCard")},
	{Use: "clear <key>", RunE: db.GetExecCommand("SClear")},
	{Use: "diff <keys...>", RunE: db.GetExecCommand("SDiff")},
	{Use: "exists <key>", RunE: db.GetExecCommand("SKeyExists")},
	{Use: "expire <key> <duration>", RunE: db.GetExecCommand("SExpire")},
	{Use: "is-member <key> <member>", RunE: db.GetExecCommand("SIsMember")},
	{Use: "members <key>", RunE: db.GetExecCommand("SMembers")},
	{Use: "move <source> <destination> <member>", RunE: db.GetExecCommand("SMove")},
	{Use: "random <key> <count>", RunE: db.GetExecCommand("SRandMember")},
	{Use: "remove <key> <members...>", RunE: db.GetExecCommand("SRandMember")},
	{Use: "ttl <key>", RunE: db.GetExecCommand("STTL")},
	{Use: "union <keys...>", RunE: db.GetExecCommand("SUnion")},
}

func init() {
	setCmd.AddCommand(setCommands...)
}
