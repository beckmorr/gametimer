package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/beckmorr/gametimer/internal/config"
	"github.com/beckmorr/gametimer/internal/theme"
	"github.com/beckmorr/gametimer/internal/tui"
)

var (
	startProcess string
	startSession int
	startExtra   int
	startTheme   string
)

var startCmd = &cobra.Command{
	Use:   "start <game>",
	Short: "Starts a timed session for a game",
	Long: `Starts the timer for <game> in a full-screen TUI.

If the game has been used before, the saved values (process, session
duration, extra time) are reused automatically, no need to repeat the
flags. Any flag passed here only overrides this run and updates what's
saved for next time.

Without --process, the session ends with a warning and there's no
auto-close. With --process, after the session you get the extra time
(--extra) to save before the game gets closed on its own.`,
	Example: `  gametimer start "Elden Ring" --process eldenring.exe --session 90 --extra 15 --theme dracula
  gametimer start "Elden Ring"
  gametimer start "Stardew Valley" --session 45`,
	Args: cobra.ExactArgs(1),
	RunE: runStart,
}

func init() {
	startCmd.Flags().StringVarP(&startProcess, "process", "p", "", "game process pattern (pgrep -f), for auto-close")
	startCmd.Flags().IntVarP(&startSession, "session", "s", 0, "session duration in minutes")
	startCmd.Flags().IntVarP(&startExtra, "extra", "e", 0, "extra time to save, in minutes")
	startCmd.Flags().StringVarP(&startTheme, "theme", "t", "", "theme for this session, doesn't change the default (see: gametimer theme list)")
	rootCmd.AddCommand(startCmd)
}

func runStart(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}
	games, err := config.LoadGames()
	if err != nil {
		return fmt.Errorf("reading game list: %w", err)
	}

	game, found := config.FindGame(games, name)
	if !found {
		game = config.Game{
			Name:       name,
			SessionMin: cfg.DefaultSessionMin,
			ExtraMin:   cfg.DefaultExtraMin,
		}
	}

	if cmd.Flags().Changed("process") {
		game.Process = startProcess
	}
	if cmd.Flags().Changed("session") {
		game.SessionMin = startSession
	}
	if cmd.Flags().Changed("extra") {
		game.ExtraMin = startExtra
	}
	if game.SessionMin <= 0 {
		return fmt.Errorf("session duration must be greater than zero (use --session)")
	}
	if game.ExtraMin < 0 {
		return fmt.Errorf("extra time can't be negative (use --extra)")
	}

	if _, err := config.UpsertGame(game); err != nil {
		return fmt.Errorf("saving game: %w", err)
	}

	themeName := cfg.DefaultTheme
	if startTheme != "" {
		themeName = startTheme
	}
	th, ok := theme.Get(themeName)
	if !ok {
		return fmt.Errorf("theme %q doesn't exist, see `gametimer theme list`", themeName)
	}

	if game.Process == "" {
		fmt.Printf("Starting session for %s: %d min (no process monitored, no auto-close).\n", game.Name, game.SessionMin)
	} else {
		fmt.Printf("Starting session for %s: %d min + %d min extra to save (process: %s).\n", game.Name, game.SessionMin, game.ExtraMin, game.Process)
	}

	status, err := tui.Run(game, th)
	if err != nil {
		return fmt.Errorf("running TUI: %w", err)
	}

	switch status {
	case "saved_in_time":
		fmt.Println("✔ Session ended. You saved in time.")
	case "completed":
		fmt.Println("✔ Session ended.")
	case "cancelled":
		fmt.Println("■ Session cancelled.")
	}
	return nil
}
