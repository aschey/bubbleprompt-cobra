package cmd

import (
	"examples/keyvalue/db"

	"github.com/spf13/cobra"
)

var zsetCommands = []*cobra.Command{
	{Use: "add <key> <score> <member>", RunE: db.GetExecCommand("ZAdd")},
	{Use: "card <key>", RunE: db.GetExecCommand("ZCard"), ValidArgsFunction: db.ZGetKeys},
	{Use: "clear <key>", RunE: db.GetExecCommand("ZClear"), ValidArgsFunction: db.ZGetKeys},
	{Use: "exists <key>", RunE: db.GetExecCommand("ZKeyExists"), ValidArgsFunction: db.ZGetKeys},
	{Use: "expire <key> <duration>", RunE: db.GetExecCommand("ZExpire"), ValidArgsFunction: db.ZGetKeys},
	{Use: "remove <key> <member>", RunE: db.GetExecCommand("ZRem"), ValidArgsFunction: db.ZGetKeys},
	{Use: "score <key> <member>", RunE: db.GetExecCommand("ZScore"), ValidArgsFunction: db.ZGetKeys},
	{Use: "ttl <key>", RunE: db.GetExecCommand("ZTTL"), ValidArgsFunction: db.ZGetKeys},
}

var getByRankCommand = &cobra.Command{Use: "get-by-rank <key> <rank>", RunE: func(cmd *cobra.Command, args []string) error {

	if reverse, _ := cmd.Flags().GetBool("reverse"); reverse {
		return db.GetExecCommand("ZRevGetByRank")(cmd, args)
	}
	return db.GetExecCommand("ZGetByRank")(cmd, args)
}, ValidArgsFunction: db.ZGetKeys}
var rangeCommand = &cobra.Command{Use: "range <key> <start> <stop>", RunE: func(cmd *cobra.Command, args []string) error {
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
}, ValidArgsFunction: db.ZGetKeys}
var rankCommand = &cobra.Command{Use: "rank <key> <member>", RunE: func(cmd *cobra.Command, args []string) error {

	if reverse, _ := cmd.Flags().GetBool("reverse"); reverse {
		return db.GetExecCommand("ZRevRank")(cmd, args)
	}
	return db.GetExecCommand("ZRanks")(cmd, args)
}, ValidArgsFunction: db.ZGetKeys}
var scoreRangeCommand = &cobra.Command{Use: "score-range <key> <min> <max>", RunE: func(cmd *cobra.Command, args []string) error {
	if reverse, _ := cmd.Flags().GetBool("reverse"); reverse {
		return db.GetExecCommand("ZRevScoreRange")(cmd, []string{args[1], args[0]})
	}
	return db.GetExecCommand("ZRevRange")(cmd, args)
}, ValidArgsFunction: db.ZGetKeys}

func init() {
	zsetCmd.AddCommand(append(zsetCommands, getByRankCommand, rangeCommand, rankCommand)...)
	getByRankCommand.Flags().BoolP("reverse", "r", false, "Display in reverse order")
	rangeCommand.Flags().BoolP("reverse", "r", false, "Display in reverse order")
	rankCommand.Flags().BoolP("reverse", "r", false, "Display in reverse order")
	scoreRangeCommand.Flags().BoolP("reverse", "r", false, "Display in reverse order")
	rangeCommand.Flags().BoolP("scores", "s", false, "Include scores")
}
