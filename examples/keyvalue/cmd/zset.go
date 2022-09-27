package cmd

import (
	"examples/keyvalue/db"

	"github.com/spf13/cobra"
)

var zsetCommands = []*cobra.Command{
	{Use: "add <key> <score> <member>", RunE: db.GetExecCommand("ZAdd")},
	{Use: "card <key>", RunE: db.GetExecCommand("ZCard")},
	{Use: "clear <key>", RunE: db.GetExecCommand("ZClear")},
	{Use: "exists <key>", RunE: db.GetExecCommand("ZKeyExists")},
	{Use: "expire <key> <duration>", RunE: db.GetExecCommand("ZExpire")},
	{Use: "remove <key> <member>", RunE: db.GetExecCommand("ZRem")},
	{Use: "score <key> <member>", RunE: db.GetExecCommand("ZScore")},
	{Use: "ttl <key>", RunE: db.GetExecCommand("ZTTL")},
}

var getByRankCommand = &cobra.Command{Use: "get-by-rank <key> <rank>", RunE: func(cmd *cobra.Command, args []string) error {
	reverse, _ := cmd.Flags().GetBool("reverse")
	if reverse {
		return db.GetExecCommand("ZRevGetByRank")(cmd, args)
	}
	return db.GetExecCommand("ZGetByRank")(cmd, args)
}}
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
}}
var rankCommand = &cobra.Command{Use: "rank <key> <member>", RunE: func(cmd *cobra.Command, args []string) error {
	reverse, _ := cmd.Flags().GetBool("reverse")
	if reverse {
		return db.GetExecCommand("ZRevRank")(cmd, args)
	}
	return db.GetExecCommand("ZRanks")(cmd, args)
}}
var scoreRangeCommand = &cobra.Command{Use: "score-range <key> <min> <max>", RunE: func(cmd *cobra.Command, args []string) error {
	reverse, _ := cmd.Flags().GetBool("reverse")
	if reverse {
		return db.GetExecCommand("ZRevScoreRange")(cmd, []string{args[1], args[0]})
	}
	return db.GetExecCommand("ZRevRange")(cmd, args)
}}

func init() {
	zsetCmd.AddCommand(append(zsetCommands, getByRankCommand, rangeCommand, rankCommand)...)
	getByRankCommand.Flags().BoolP("reverse", "r", false, "Display in reverse order")
	rangeCommand.Flags().BoolP("reverse", "r", false, "Display in reverse order")
	rankCommand.Flags().BoolP("reverse", "r", false, "Display in reverse order")
	scoreRangeCommand.Flags().BoolP("reverse", "r", false, "Display in reverse order")
	rangeCommand.Flags().BoolP("scores", "s", false, "Include scores")
}
