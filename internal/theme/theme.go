package theme

import "sort"

type Theme struct {
	Name       string
	Session    string
	Warning    string
	Success    string
	Accent     string
	Muted      string
	Text       string
	Background string
}

var themes = map[string]Theme{
	"default": {
		Name:       "default",
		Session:    "#339AF0",
		Warning:    "#FF6B6B",
		Success:    "#51CF66",
		Accent:     "#FFD43B",
		Muted:      "#666666",
		Text:       "#FFFFFF",
		Background: "#1A1A1A",
	},
	"catppuccin-mocha": {
		Name:       "catppuccin-mocha",
		Session:    "#89B4FA",
		Warning:    "#F38BA8",
		Success:    "#A6E3A1",
		Accent:     "#F9E2AF",
		Muted:      "#585B70",
		Text:       "#CDD6F4",
		Background: "#1E1E2E",
	},
	"dracula": {
		Name:       "dracula",
		Session:    "#8BE9FD",
		Warning:    "#FF5555",
		Success:    "#50FA7B",
		Accent:     "#F1FA8C",
		Muted:      "#6272A4",
		Text:       "#F8F8F2",
		Background: "#282A36",
	},
	"gruvbox": {
		Name:       "gruvbox",
		Session:    "#83A598",
		Warning:    "#FB4934",
		Success:    "#B8BB26",
		Accent:     "#FABD2F",
		Muted:      "#665C54",
		Text:       "#EBDBB2",
		Background: "#282828",
	},
	"nord": {
		Name:       "nord",
		Session:    "#88C0D0",
		Warning:    "#BF616A",
		Success:    "#A3BE8C",
		Accent:     "#EBCB8B",
		Muted:      "#4C566A",
		Text:       "#ECEFF4",
		Background: "#2E3440",
	},
	"tokyo-night": {
		Name:       "tokyo-night",
		Session:    "#7AA2F7",
		Warning:    "#F7768E",
		Success:    "#9ECE6A",
		Accent:     "#E0AF68",
		Muted:      "#565F89",
		Text:       "#C0CAF5",
		Background: "#1A1B26",
	},
	"solarized": {
		Name:       "solarized",
		Session:    "#268BD2",
		Warning:    "#DC322F",
		Success:    "#859900",
		Accent:     "#B58900",
		Muted:      "#586E75",
		Text:       "#FDF6E3",
		Background: "#002B36",
	},
}

func Names() []string {
	names := make([]string, 0, len(themes))
	for n := range themes {
		if n != "default" {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	return append([]string{"default"}, names...)
}

func Get(name string) (Theme, bool) {
	t, ok := themes[name]
	return t, ok
}
