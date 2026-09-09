package tui

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/beckmorr/gametimer/internal/session"
)

// bannerMargin reserves room for the box border, its padding, and a small
// safety buffer so the framed content never exceeds the terminal width.
const bannerMargin = 20

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

	if m.width > 0 && m.height > 0 {
		return m.placeWithStars(content)
	}
	return content
}

// placeWithStars centers box on the terminal and, instead of a flat
// background fill, scatters a sparse random field of white dots across the
// space around it. rng is re-seeded from m.starSeed on every call so the
// same dots stay put across renders rather than jittering every tick.
func (m Model) placeWithStars(box string) string {
	bg := lipgloss.Color(m.th.Background)
	starStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(bg)
	bgStyle := lipgloss.NewStyle().Background(bg)

	lines := strings.Split(box, "\n")
	boxH := len(lines)
	boxW := 0
	for _, l := range lines {
		if w := lipgloss.Width(l); w > boxW {
			boxW = w
		}
	}

	top := (m.height - boxH) / 2
	left := (m.width - boxW) / 2
	if top < 0 {
		top = 0
	}
	if left < 0 {
		left = 0
	}
	right := m.width - left - boxW
	if right < 0 {
		right = 0
	}

	rng := rand.New(rand.NewSource(m.starSeed))

	rows := make([]string, m.height)
	for y := 0; y < m.height; y++ {
		if y >= top && y < top+boxH {
			rows[y] = starRow(rng, starStyle, bgStyle, left) + lines[y-top] + starRow(rng, starStyle, bgStyle, right)
		} else {
			rows[y] = starRow(rng, starStyle, bgStyle, m.width)
		}
	}
	return strings.Join(rows, "\n")
}

// starRow renders width background cells with an occasional white "."
// sprinkled in (roughly one every 60 cells) and a much rarer white "*"
// (roughly one every 1200 cells, twenty times scarcer than the dots).
func starRow(rng *rand.Rand, starStyle, bgStyle lipgloss.Style, width int) string {
	if width <= 0 {
		return ""
	}
	var b strings.Builder
	run := 0
	flush := func() {
		if run > 0 {
			b.WriteString(bgStyle.Render(strings.Repeat(" ", run)))
			run = 0
		}
	}
	for i := 0; i < width; i++ {
		switch n := rng.Intn(1200); {
		case n == 0:
			flush()
			b.WriteString(starStyle.Render("*"))
		case n < 21:
			flush()
			b.WriteString(starStyle.Render("."))
		default:
			run++
		}
	}
	flush()
	return b.String()
}

func (m Model) bannerWidth() int {
	if m.width <= 0 {
		return 76
	}
	return m.width - bannerMargin
}

// contentWidth is the width used for the divider and progress bar: capped
// well below the banner width so the box stays a comfortable size even on
// wide terminals.
func (m Model) contentWidth() int {
	const max = 56
	if bw := m.bannerWidth(); bw > 0 && bw < max {
		return bw
	}
	return max
}

// box frames body in a rounded border colored for the current phase, giving
// the whole screen a "card" look instead of loose centered text.
func (m Model) box(hex, body string) string {
	bg := lipgloss.Color(m.th.Background)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(hex)).
		BorderBackground(bg).
		Background(bg).
		Padding(1, 3).
		Render(body)
}

func eyebrow(hex, bg, text string) string {
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(hex)).Background(lipgloss.Color(bg)).Render(text)
}

func divider(hex, bg string, width int) string {
	if width <= 0 {
		width = 40
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(hex)).Background(lipgloss.Color(bg)).Render(strings.Repeat("─", width))
}

// progressBar renders a filled/empty block bar for frac (0..1).
func progressBar(fillHex, emptyHex, bg string, width int, frac float64) string {
	if width <= 0 {
		width = 30
	}
	if frac < 0 {
		frac = 0
	} else if frac > 1 {
		frac = 1
	}
	filled := int(float64(width)*frac + 0.5)
	if filled > width {
		filled = width
	}
	fillStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(fillHex)).Background(lipgloss.Color(bg))
	emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(emptyHex)).Background(lipgloss.Color(bg))
	return fillStyle.Render(strings.Repeat("█", filled)) + emptyStyle.Render(strings.Repeat("░", width-filled))
}

// kv renders a "label   value" line with the label padded to a fixed column
// so consecutive kv lines line up like a small receipt.
func kv(labelStyle, valueStyle lipgloss.Style, label, value string) string {
	return labelStyle.Render(fmt.Sprintf("%-14s", label)) + valueStyle.Render(value)
}

func (m Model) viewSession() string {
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color(m.th.Muted)).Background(lipgloss.Color(m.th.Background))
	text := lipgloss.NewStyle().Foreground(lipgloss.Color(m.th.Text)).Background(lipgloss.Color(m.th.Background))

	banner := bigTitle(m.game.Name, m.th.Session, m.th.Background, m.bannerWidth())
	div := divider(m.th.Muted, m.th.Background, m.contentWidth())

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

	body := strings.Join([]string{
		eyebrow(m.th.Session, m.th.Background, "🎮  GAME SESSION"),
		"",
		banner,
		"",
		div,
		"",
		info,
		"",
		status,
		"",
		help,
	}, "\n")
	return m.box(m.th.Session, body)
}

func (m Model) viewExtra() string {
	text := lipgloss.NewStyle().Foreground(lipgloss.Color(m.th.Text)).Background(lipgloss.Color(m.th.Background))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color(m.th.Muted)).Background(lipgloss.Color(m.th.Background))
	warn := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.th.Warning)).Background(lipgloss.Color(m.th.Background))

	banner := bigTitle(m.game.Name, m.th.Warning, m.th.Background, m.bannerWidth())
	div := divider(m.th.Muted, m.th.Background, m.contentWidth())

	total := time.Duration(m.game.ExtraMin) * time.Minute
	var frac float64
	if total > 0 {
		frac = 1 - m.extraRemaining.Seconds()/total.Seconds()
	}
	bar := progressBar(m.th.Warning, m.th.Muted, m.th.Background, m.contentWidth(), frac)

	var clock string
	if m.paused {
		clock = lipgloss.NewStyle().Foreground(lipgloss.Color(m.th.Accent)).Background(lipgloss.Color(m.th.Background)).Render("⏸  Paused. Press space to resume.")
	} else {
		clock = text.Render("Closing in ") + warn.Render(formatMMSS(m.extraRemaining))
	}

	msg := text.Render("Save your progress now! The game will close automatically.")
	help := muted.Render("[space] pause/resume    [q] end now")

	body := strings.Join([]string{
		eyebrow(m.th.Warning, m.th.Background, "⚠  TIME TO SAVE"),
		"",
		banner,
		"",
		div,
		"",
		clock,
		bar,
		"",
		msg,
		"",
		help,
	}, "\n")
	return m.box(m.th.Warning, body)
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
	div := divider(m.th.Muted, m.th.Background, m.contentWidth())

	played := kv(muted, text, "Time played:", formatPlayed(m.playedSeconds2Minutes()))
	msg := text.Render(m.finalMsg)

	logLine := muted.Render("Log saved to " + m.logPath)
	if m.logErr != nil {
		logLine = lipgloss.NewStyle().Foreground(lipgloss.Color(m.th.Warning)).Background(lipgloss.Color(m.th.Background)).Render("Failed to save log: " + m.logErr.Error())
	}

	help := muted.Render("[q] quit")

	body := strings.Join([]string{
		eyebrow(color, m.th.Background, label),
		"",
		banner,
		"",
		div,
		"",
		played,
		msg,
		"",
		logLine,
		"",
		help,
	}, "\n")
	return m.box(color, body)
}

func (m Model) playedSeconds2Minutes() float64 {
	return float64(m.playedSeconds) / 60.0
}
