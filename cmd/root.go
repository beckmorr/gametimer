package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gametimer",
	Short: "Game session timer with a TUI",
	Long: `gametimer times your game session in a full-screen TUI.

It warns you when the session ends and, if you give it the game's process
with --process, opens an extra-time window for you to save before it
closes the game automatically. Every session is logged so you can track
how many hours you've played (see "gametimer stats").`,
	Example: `  gametimer start "Elden Ring" --process eldenring.exe --session 90 --extra 15 --theme dracula
  gametimer start "Elden Ring"
  gametimer games list
  gametimer theme set dracula
  gametimer stats --period week`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
