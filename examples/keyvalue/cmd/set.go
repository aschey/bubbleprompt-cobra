package cmd

import (
	"examples/keyvalue/db"

	"github.com/spf13/cobra"
)

var setCommands = []*cobra.Command{
	{
		Use:   "add <key> <members...>",
		Short: "Add the members to the key",
		RunE:  db.GetExecCommand("SAdd"),
		Args:  cobra.MinimumNArgs(2),
	},
	{
		Use:               "card <key>",
		Short:             "Get the cardinality of the key",
		RunE:              db.GetExecCommand("SCard"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.SGetKeys,
	},
	{
		Use:               "clear <key>",
		Short:             "Clear all the values from the key",
		RunE:              db.GetExecCommand("SClear"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.SGetKeys,
	},
	{
		Use:               "diff <keys...>",
		Short:             "Get the difference between the keys",
		RunE:              db.GetListExecCommand("SDiff"),
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: db.SGetKeysN(-1),
	},
	{
		Use:               "exists <key>",
		Short:             "Check if the key exists",
		RunE:              db.GetExecCommand("SKeyExists"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.SGetKeys,
	},
	{
		Use:               "expire <key> <duration>",
		Short:             "Expire the key after the specified duration",
		RunE:              db.GetExecCommand("SExpire"),
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: db.SGetKeys,
	},
	{
		Use:               "is-member <key> <member>",
		Short:             "Check if the value is a member of the key",
		RunE:              db.GetExecCommand("SIsMember"),
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: db.SGetKeys,
	},
	{
		Use:               "members <key>",
		Short:             "Get all the members of the key",
		RunE:              db.GetListExecCommand("SMembers"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.SGetKeys,
	},

	{
		Use:               "move <source> <destination> <member>",
		Short:             "Move the member from the source to the destination",
		RunE:              db.GetExecCommand("SMove"),
		Args:              cobra.ExactArgs(3),
		ValidArgsFunction: db.SGetKeysN(2),
	},
	{
		Use:               "random <key> <count>",
		Short:             "Get random members from the key",
		RunE:              db.GetExecCommand("SRandMember"),
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: db.SGetKeys,
	},
	{
		Use:               "remove <key> <members...>",
		Short:             "Remove the members from the key",
		RunE:              db.GetExecCommand("SRem"),
		Args:              cobra.MinimumNArgs(2),
		ValidArgsFunction: db.SGetKeys,
	},
	{
		Use:               "ttl <key>",
		Short:             "Get the time-to-live of the key",
		RunE:              db.GetExecCommand("STTL"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.SGetKeys,
	},
	{
		Use:               "union <keys...>",
		Short:             "Get the union of all the keys",
		RunE:              db.GetListExecCommand("SUnion"),
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: db.SGetKeysN(-1),
	},
}

func init() {
	setCmd.AddCommand(setCommands...)
}
