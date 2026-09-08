package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/beckmorr/gametimer/internal/session"
)

var (
	statsGame   string
	statsPeriod string
	statsJSON   bool
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Shows total hours played",
	Long:  `Sums up play time from the session log, overall or by game.`,
	Example: `  gametimer stats
  gametimer stats --period week
  gametimer stats --game "Elden Ring" --period month
  gametimer stats --json`,
	RunE: runStats,
}

func init() {
	statsCmd.Flags().StringVar(&statsGame, "game", "", "filter by game name")
	statsCmd.Flags().StringVar(&statsPeriod, "period", "all", "period: today, week, month, all")
	statsCmd.Flags().BoolVar(&statsJSON, "json", false, "output as JSON")
	rootCmd.AddCommand(statsCmd)
}

func periodSince(period string) (time.Time, error) {
	now := time.Now()
	switch period {
	case "all":
		return time.Time{}, nil
	case "today":
		y, m, d := now.Date()
		return time.Date(y, m, d, 0, 0, 0, 0, now.Location()), nil
	case "week":
		return now.AddDate(0, 0, -7), nil
	case "month":
		return now.AddDate(0, -1, 0), nil
	default:
		return time.Time{}, fmt.Errorf("invalid period %q (use today, week, month or all)", period)
	}
}

func runStats(cmd *cobra.Command, args []string) error {
	since, err := periodSince(statsPeriod)
	if err != nil {
		return err
	}
	entries, err := session.Load()
	if err != nil {
		return err
	}
	stats := session.Aggregate(entries, statsGame, since)

	if statsJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(stats)
	}

	if stats.TotalSessions == 0 {
		fmt.Println("No sessions recorded for that period.")
		return nil
	}

	fmt.Printf("Sessions: %d\n", stats.TotalSessions)
	fmt.Printf("Total played: %s\n\n", formatHours(stats.TotalMinutes))

	names := make([]string, 0, len(stats.ByGame))
	for name := range stats.ByGame {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool { return stats.ByGame[names[i]] > stats.ByGame[names[j]] })
	for _, name := range names {
		fmt.Printf("  %-30s %s\n", name, formatHours(stats.ByGame[name]))
	}
	return nil
}

func formatHours(minutes float64) string {
	total := int(minutes + 0.5)
	h, m := total/60, total%60
	return fmt.Sprintf("%dh%02dmin", h, m)
}
