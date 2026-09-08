package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/beckmorr/gametimer/internal/session"
)

const bannerMargin = 10

func (m Model) View() string {
	var content string
	switch m.phase {
	case phaseSession:
		content = m.viewSession()
	case phaseExtra:
		content = m.viewExtra()
	default:
		content = m.viewDone()
	}

	bg := lipgloss.Color(m.th.Background)
	content = lipgloss.NewStyle().Background(bg).Width(lipgloss.Width(content)).Align(lipgloss.Center).Render(content)

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content, lipgloss.WithWhitespaceBackground(bg))
	}
	return content
}

func (m Model) bannerWidth() int {
	if m.width <= 0 {
		return 76
	}
	return m.width - bannerMargin
}

func eyebrow(hex, bg, text string) string {
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(hex)).Background(lipgloss.Color(bg)).Render(text)
}

func (m Model) viewSession() string {
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color(m.th.Muted)).Background(lipgloss.Color(m.th.Background))
	text := lipgloss.NewStyle().Foreground(lipgloss.Color(m.th.Text)).Background(lipgloss.Color(m.th.Background))

	banner := bigTitle(m.game.Name, m.th.Session, m.th.Background, m.bannerWidth())

	info := fmt.Sprintf(
		"%s %s    %s %s",
		muted.Render("Session:"), text.Render(fmt.Sprintf("%d min", m.game.SessionMin)),
		muted.Render("Extra time:"), text.Render(fmt.Sprintf("%d min", m.game.ExtraMin)),
	)

	var status string
	if m.paused {
		status = lipgloss.NewStyle().Foreground(lipgloss.Color(m.th.Accent)).Background(lipgloss.Color(m.th.Background)).Render("⏸  Paused. Press space to resume.")
	} else {
		status = m.spin.View() + " " + text.Render("Session in progress. Enjoy the game.")
	}

	help := muted.Render("[space] pause/resume    [q] end now")

	return strings.Join([]string{
		eyebrow(m.th.Session, m.th.Background, "🎮  GAME SESSION"),
		"",
		banner,
		"",
		info,
		"",
		status,
		"",
		help,
	}, "\n")
}

func (m Model) viewExtra() string {
	text := lipgloss.NewStyle().Foreground(lipgloss.Color(m.th.Text)).Background(lipgloss.Color(m.th.Background))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color(m.th.Muted)).Background(lipgloss.Color(m.th.Background))
	warn := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.th.Warning)).Background(lipgloss.Color(m.th.Background))

	banner := bigTitle(m.game.Name, m.th.Warning, m.th.Background, m.bannerWidth())

	var clock string
	if m.paused {
		clock = lipgloss.NewStyle().Foreground(lipgloss.Color(m.th.Accent)).Background(lipgloss.Color(m.th.Background)).Render("⏸  Paused. Press space to resume.")
	} else {
		clock = text.Render("Closing in ") + warn.Render(formatMMSS(m.extraRemaining))
	}

	msg := text.Render("Save your progress now! The game will close automatically.")
	help := muted.Render("[space] pause/resume    [q] end now")

	return strings.Join([]string{
		eyebrow(m.th.Warning, m.th.Background, "⚠  TIME TO SAVE"),
		"",
		banner,
		"",
		clock,
		"",
		msg,
		"",
		help,
	}, "\n")
}

func (m Model) viewDone() string {
	text := lipgloss.NewStyle().Foreground(lipgloss.Color(m.th.Text)).Background(lipgloss.Color(m.th.Background))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color(m.th.Muted)).Background(lipgloss.Color(m.th.Background))

	color := m.th.Success
	label := "✔  SESSION ENDED"
	if m.finalStatus == session.StatusCancelled {
		color = m.th.Muted
		label = "■  SESSION CANCELLED"
	} else if m.finalStatus == session.StatusCompleted && m.game.Process != "" {
		color = m.th.Warning
		label = "⏻  GAME CLOSED AUTOMATICALLY"
	}

	banner := bigTitle(m.game.Name, color, m.th.Background, m.bannerWidth())
	played := text.Render("Time played: " + formatPlayed(m.playedSeconds2Minutes()))
	msg := text.Render(m.finalMsg)

	logLine := muted.Render("Log saved to " + m.logPath)
	if m.logErr != nil {
		logLine = lipgloss.NewStyle().Foreground(lipgloss.Color(m.th.Warning)).Background(lipgloss.Color(m.th.Background)).Render("Failed to save log: " + m.logErr.Error())
	}

	help := muted.Render("[q] quit")

	return strings.Join([]string{
		eyebrow(color, m.th.Background, label),
		"",
		banner,
		"",
		played,
		msg,
		"",
		logLine,
		"",
		help,
	}, "\n")
}

func (m Model) playedSeconds2Minutes() float64 {
	return float64(m.playedSeconds) / 60.0
}
