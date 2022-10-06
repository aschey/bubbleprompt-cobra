package cmd

import (
	"examples/keyvalue/db"

	"github.com/spf13/cobra"
)

var zsetCommands = []*cobra.Command{
	{
		Use:   "add <key> <score> <member>",
		Short: "Add the value to the key",
		RunE:  db.GetExecCommand("ZAdd"),
		Args:  cobra.ExactArgs(3),
	},
	{
		Use:               "card <key>",
		Short:             "Get the cardinality of the key",
		RunE:              db.GetExecCommand("ZCard"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.ZGetKeys,
	},
	{
		Use:               "clear <key>",
		Short:             "Clear all the values from the key",
		RunE:              db.GetExecCommand("ZClear"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.ZGetKeys,
	},
	{
		Use:               "exists <key>",
		Short:             "Check if the key exists",
		RunE:              db.GetExecCommand("ZKeyExists"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.ZGetKeys,
	},
	{
		Use:               "expire <key> <duration>",
		Short:             "Expire the key after the specified duration",
		RunE:              db.GetExecCommand("ZExpire"),
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: db.ZGetKeys,
	},
	{
		Use:               "remove <key> <member>",
		Short:             "Remove the member from the key",
		RunE:              db.GetExecCommand("ZRem"),
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: db.ZGetKeys,
	},
	{
		Use:               "score <key> <member>",
		Short:             "Get the score of the member",
		RunE:              db.GetExecCommand("ZScore"),
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: db.ZGetKeys,
	},
	{
		Use:               "ttl <key>",
		Short:             "Get the time-to-live of the key",
		RunE:              db.GetExecCommand("ZTTL"),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: db.ZGetKeys,
	},
	getByRankCommand,
	rangeCommand,
	rankCommand,
	scoreRangeCommand,
}

var getByRankCommand = &cobra.Command{
	Use:   "get-by-rank <key> <rank>",
	Short: "Get all values from the key that match the rank",
	RunE: func(cmd *cobra.Command,
		args []string) error {
		if reverse, _ := cmd.Flags().GetBool("reverse"); reverse {
			return db.GetExecCommand("ZRevGetByRank")(cmd, args)
		}
		return db.GetExecCommand("ZGetByRank")(cmd, args)
	},
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: db.ZGetKeys,
}
var rangeCommand = &cobra.Command{
	Use:   "range <key> <start> <stop>",
	Short: "Get all values in the within the range of ranks",
	RunE: func(cmd *cobra.Command, args []string) error {
		reverse, _ := cmd.Flags().GetBool("reverse")
		scores, _ := cmd.Flags().GetBool("scores")

		if scores && reverse {
			return db.GetExecCommand("ZRevRangeWithScores")(cmd, args)
		} else if scores {
			return db.GetExecCommand("ZRangeWithScores")(cmd, args)
		} else if reverse {
			return db.GetExecCommand("ZRevRange")(cmd, args)
		} else {
			return db.GetExecCommand("ZRange")(cmd, args)
		}
	},
	Args:              cobra.ExactArgs(3),
	ValidArgsFunction: db.ZGetKeys,
}
var rankCommand = &cobra.Command{
	Use:   "rank <key> <member>",
	Short: "Get the rank of the member",
	RunE: func(cmd *cobra.Command,
		args []string) error {
		if reverse,
			_ := cmd.Flags().GetBool("reverse"); reverse {
			return db.GetExecCommand("ZRevRank")(cmd, args)
		}
		return db.GetExecCommand("ZRank")(cmd, args)
	},
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: db.ZGetKeys,
}
var scoreRangeCommand = &cobra.Command{
	Use:   "score-range <key> <min> <max>",
	Short: "Get all values within the range of scores",
	RunE: func(cmd *cobra.Command,
		args []string) error {

		if reverse, _ := cmd.Flags().GetBool("reverse"); reverse {
			return db.GetListExecCommand("ZRevScoreRange")(cmd, []string{args[1], args[0]})
		}
		return db.GetListExecCommand("ZRevRange")(cmd, args)
	},
	Args:              cobra.ExactArgs(3),
	ValidArgsFunction: db.ZGetKeys,
}

func init() {
	zsetCmd.AddCommand(zsetCommands...)
	getByRankCommand.Flags().BoolP("reverse", "r", false, "Display in reverse order")
	rangeCommand.Flags().BoolP("reverse", "r", false, "Display in reverse order")
	rankCommand.Flags().BoolP("reverse", "r", false, "Display in reverse order")
	scoreRangeCommand.Flags().BoolP("reverse", "r", false, "Display in reverse order")
	rangeCommand.Flags().BoolP("scores", "s", false, "Include scores")
}
