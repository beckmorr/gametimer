package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/beckmorr/gametimer/internal/config"
	"github.com/beckmorr/gametimer/internal/theme"
)

var themeCmd = &cobra.Command{
	Use:   "theme",
	Short: "Lists, previews, or sets the default theme",
	Long: `Available themes: default, catppuccin-mocha, dracula, gruvbox, nord,
tokyo-night, solarized.

The default theme (set with "theme set") is used by "gametimer start"
whenever --theme isn't passed.`,
}

var themeListCmd = &cobra.Command{
	Use:     "list",
	Short:   "Lists the available themes",
	Example: `  gametimer theme list`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		for _, name := range theme.Names() {
			marker := "  "
			if name == cfg.DefaultTheme {
				marker = "* "
			}
			fmt.Println(marker + name)
		}
		return nil
	},
}

var themePreviewCmd = &cobra.Command{
	Use:     "preview <theme>",
	Short:   "Shows a theme's colors in the terminal",
	Example: `  gametimer theme preview dracula`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		th, ok := theme.Get(args[0])
		if !ok {
			return fmt.Errorf("theme %q doesn't exist, see `gametimer theme list`", args[0])
		}
		swatch := func(label, hex string) string {
			return lipgloss.NewStyle().Foreground(lipgloss.Color(hex)).Render(fmt.Sprintf("%-14s %s", label, hex))
		}
		fmt.Printf("Theme: %s\n\n", th.Name)
		fmt.Println(swatch("Session:", th.Session))
		fmt.Println(swatch("Warning/extra:", th.Warning))
		fmt.Println(swatch("Success:", th.Success))
		fmt.Println(swatch("Accent:", th.Accent))
		fmt.Println(swatch("Muted:", th.Muted))
		fmt.Println(swatch("Text:", th.Text))
		fmt.Println(swatch("Background:", th.Background))
		return nil
	},
}

var themeSetCmd = &cobra.Command{
	Use:     "set <theme>",
	Short:   "Sets the default theme",
	Example: `  gametimer theme set nord`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, ok := theme.Get(args[0]); !ok {
			return fmt.Errorf("theme %q doesn't exist, see `gametimer theme list`", args[0])
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cfg.DefaultTheme = args[0]
		if err := config.Save(cfg); err != nil {
			return err
		}
		fmt.Printf("Default theme set to %q.\n", args[0])
		return nil
	},
}

func init() {
	themeCmd.AddCommand(themeListCmd, themePreviewCmd, themeSetCmd)
	rootCmd.AddCommand(themeCmd)
}
