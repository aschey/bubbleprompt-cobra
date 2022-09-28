package cmd

import (
	"examples/keyvalue/db"

	"github.com/spf13/cobra"
)

var setCommands = []*cobra.Command{
	{Use: "add <key> <members...>", RunE: db.GetExecCommand("SAdd")},
	{Use: "card <key>", RunE: db.GetExecCommand("SCard"), ValidArgsFunction: db.SGetKeys},
	{Use: "clear <key>", RunE: db.GetExecCommand("SClear"), ValidArgsFunction: db.SGetKeys},
	{Use: "diff <keys...>", RunE: db.GetExecCommand("SDiff"), ValidArgsFunction: db.SGetKeysN(-1)},
	{Use: "exists <key>", RunE: db.GetExecCommand("SKeyExists"), ValidArgsFunction: db.SGetKeys},
	{Use: "expire <key> <duration>", RunE: db.GetExecCommand("SExpire"), ValidArgsFunction: db.SGetKeys},
	{Use: "is-member <key> <member>", RunE: db.GetExecCommand("SIsMember"), ValidArgsFunction: db.SGetKeys},
	{Use: "members <key>", RunE: db.GetExecCommand("SMembers"), ValidArgsFunction: db.SGetKeys},
	{Use: "move <source> <destination> <member>", RunE: db.GetExecCommand("SMove"), ValidArgsFunction: db.SGetKeysN(2)},
	{Use: "random <key> <count>", RunE: db.GetExecCommand("SRandMember"), ValidArgsFunction: db.SGetKeys},
	{Use: "remove <key> <members...>", RunE: db.GetExecCommand("SRandMember"), ValidArgsFunction: db.SGetKeys},
	{Use: "ttl <key>", RunE: db.GetExecCommand("STTL"), ValidArgsFunction: db.SGetKeys},
	{Use: "union <keys...>", RunE: db.GetExecCommand("SUnion"), ValidArgsFunction: db.SGetKeysN(-1)},
}

func init() {
	setCmd.AddCommand(setCommands...)
}
