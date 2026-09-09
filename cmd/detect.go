package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beckmorr/gametimer/internal/proc"
)

var detectWatch bool

var detectCmd = &cobra.Command{
	Use:   "detect [query]",
	Short: "Helps find a game's process name for --process",
	Long: `Helps find the pattern to pass to "gametimer start --process", for
when the game's process name doesn't resemble the game's name (e.g.
Persona 5 Royal running as P5R.exe).

With a query, lists running processes whose command line contains it
(case-insensitive), replacing "ps aux | grep -i <name>".

With --watch, it snapshots running processes, waits for you to launch
the game, then shows only what's new: usually the game itself, even
if its process name has nothing to do with the game's name.`,
	Example: `  gametimer detect "elden ring"
  gametimer detect --watch`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDetect,
}

func init() {
	detectCmd.Flags().BoolVarP(&detectWatch, "watch", "w", false, "snapshot processes, wait for enter, then show what's new")
	rootCmd.AddCommand(detectCmd)
}

func runDetect(cmd *cobra.Command, args []string) error {
	if detectWatch {
		return runDetectWatch()
	}
	if len(args) == 0 {
		return fmt.Errorf("pass a search term, or use --watch (see `gametimer detect --help`)")
	}

	procs, err := proc.List()
	if err != nil {
		return fmt.Errorf("listing processes: %w", err)
	}
	query := strings.ToLower(args[0])
	var matched []proc.Info
	for _, p := range procs {
		if strings.Contains(strings.ToLower(p.Args), query) {
			matched = append(matched, p)
		}
	}
	if len(matched) == 0 {
		fmt.Println("No matching processes found.")
		return nil
	}
	printProcesses(matched)
	return nil
}

func runDetectWatch() error {
	before, err := proc.List()
	if err != nil {
		return fmt.Errorf("listing processes: %w", err)
	}
	seen := make(map[int]bool, len(before))
	for _, p := range before {
		seen[p.PID] = true
	}

	fmt.Println("Launch the game now, then press enter here...")
	bufio.NewReader(os.Stdin).ReadString('\n')

	after, err := proc.List()
	if err != nil {
		return fmt.Errorf("listing processes: %w", err)
	}
	var newProcs []proc.Info
	for _, p := range after {
		if !seen[p.PID] {
			newProcs = append(newProcs, p)
		}
	}
	if len(newProcs) == 0 {
		fmt.Println("No new processes found. Try again, or use `gametimer detect <name>`.")
		return nil
	}
	fmt.Println("New processes since before you launched the game:")
	printProcesses(newProcs)
	return nil
}

func printProcesses(procs []proc.Info) {
	for _, p := range procs {
		fmt.Printf("%-8d %s\n", p.PID, p.Args)
	}
}
