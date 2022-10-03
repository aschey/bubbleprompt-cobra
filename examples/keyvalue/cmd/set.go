package cmd

import (
	"examples/keyvalue/db"

	"github.com/spf13/cobra"
)

var setCommands = []*cobra.Command{
	{
		Use:  "add <key> <members...>",
		RunE: db.GetExecCommand("SAdd"),
		Args: cobra.MinimumNArgs(2),
	},
	{
		Use:               "card <key>",
		RunE:              db.GetExecCommand("SCard"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.SGetKeys,
	},
	{
		Use:               "clear <key>",
		RunE:              db.GetExecCommand("SClear"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.SGetKeys,
	},
	{
		Use:               "diff <keys...>",
		RunE:              db.GetListExecCommand("SDiff"),
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: db.SGetKeysN(-1),
	},
	{
		Use:               "exists <key>",
		RunE:              db.GetExecCommand("SKeyExists"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.SGetKeys,
	},
	{
		Use:               "expire <key> <duration>",
		RunE:              db.GetExecCommand("SExpire"),
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: db.SGetKeys,
	},
	{
		Use:               "is-member <key> <member>",
		RunE:              db.GetExecCommand("SIsMember"),
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: db.SGetKeys,
	},
	{
		Use:               "members <key>",
		RunE:              db.GetListExecCommand("SMembers"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.SGetKeys,
	},

	{
		Use:               "move <source> <destination> <member>",
		RunE:              db.GetExecCommand("SMove"),
		Args:              cobra.ExactArgs(3),
		ValidArgsFunction: db.SGetKeysN(2),
	},
	{
		Use:               "random <key> <count>",
		RunE:              db.GetExecCommand("SRandMember"),
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: db.SGetKeys,
	},
	{
		Use:               "remove <key> <members...>",
		RunE:              db.GetExecCommand("SRem"),
		Args:              cobra.MinimumNArgs(2),
		ValidArgsFunction: db.SGetKeys,
	},
	{
		Use:               "ttl <key>",
		RunE:              db.GetExecCommand("STTL"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.SGetKeys,
	},
	{
		Use:               "union <keys...>",
		RunE:              db.GetListExecCommand("SUnion"),
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: db.SGetKeysN(-1),
	},
}

func init() {
	setCmd.AddCommand(setCommands...)
}
