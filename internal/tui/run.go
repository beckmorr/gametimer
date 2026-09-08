package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/beckmorr/gametimer/internal/config"
	"github.com/beckmorr/gametimer/internal/session"
	"github.com/beckmorr/gametimer/internal/theme"
)

func Run(game config.Game, th theme.Theme) (session.Status, error) {
	m := New(game, th)
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return "", err
	}
	return final.(Model).FinalStatus(), nil
}
