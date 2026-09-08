package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/beckmorr/gametimer/internal/config"
)

var gamesCmd = &cobra.Command{
	Use:   "games",
	Short: "Manages saved games",
	Long: `Lists or removes the games saved by "gametimer start": name, monitored
process, and each one's default session and extra time durations.`,
}

var gamesListCmd = &cobra.Command{
	Use:     "list",
	Short:   "Lists saved games and their settings",
	Example: `  gametimer games list`,
	RunE: func(cmd *cobra.Command, args []string) error {
		games, err := config.LoadGames()
		if err != nil {
			return err
		}
		if len(games) == 0 {
			fmt.Println("No games saved yet. Use `gametimer start <game> --process <pattern> --session 90 --extra 15`.")
			return nil
		}
		for _, g := range games {
			proc := g.Process
			if proc == "" {
				proc = "(none, no auto-close)"
			}
			fmt.Printf("%-30s session=%dmin extra=%dmin process=%s\n", g.Name, g.SessionMin, g.ExtraMin, proc)
		}
		return nil
	},
}

var gamesRemoveCmd = &cobra.Command{
	Use:     "remove <game>",
	Short:   "Removes a saved game",
	Example: `  gametimer games remove "Elden Ring"`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		found, err := config.RemoveGame(args[0])
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("game %q not found", args[0])
		}
		fmt.Printf("Game %q removed.\n", args[0])
		return nil
	},
}

func init() {
	gamesCmd.AddCommand(gamesListCmd, gamesRemoveCmd)
	rootCmd.AddCommand(gamesCmd)
}
