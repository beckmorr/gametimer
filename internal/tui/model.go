package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/beckmorr/gametimer/internal/config"
	"github.com/beckmorr/gametimer/internal/notify"
	"github.com/beckmorr/gametimer/internal/proc"
	"github.com/beckmorr/gametimer/internal/session"
	"github.com/beckmorr/gametimer/internal/theme"
)

type phase int

const (
	phaseSession phase = iota
	phaseExtra
	phaseDone
)

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

type Model struct {
	game   config.Game
	th     theme.Theme
	spin   spinner.Model
	width  int
	height int

	phase  phase
	paused bool

	sessionRemaining time.Duration
	extraRemaining   time.Duration
	playedSeconds    int

	startedAt   time.Time
	finalStatus session.Status
	finalMsg    string
	logPath     string
	logErr      error
}

func New(game config.Game, th theme.Theme) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Points
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(th.Session)).Background(lipgloss.Color(th.Background))
	logPath, _ := config.LogPath()
	return Model{
		game:             game,
		th:               th,
		spin:             sp,
		phase:            phaseSession,
		sessionRemaining: time.Duration(game.SessionMin) * time.Minute,
		startedAt:        time.Now(),
		logPath:          logPath,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spin.Tick, tickCmd())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.phase == phaseDone {
				return m, tea.Quit
			}
			m.finish(session.StatusCancelled, "Session ended manually.")
			return m, nil
		case "enter":
			if m.phase == phaseDone {
				return m, tea.Quit
			}
		case " ":
			if m.phase != phaseDone {
				m.paused = !m.paused
			}
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd

	case tickMsg:
		return m.handleTick()
	}
	return m, nil
}

func (m Model) handleTick() (tea.Model, tea.Cmd) {
	if m.phase == phaseDone {
		return m, nil
	}
	if m.paused {
		return m, tickCmd()
	}

	m.playedSeconds++

	switch m.phase {
	case phaseSession:
		m.sessionRemaining -= time.Second
		if m.sessionRemaining <= 0 {
			notify.Send("Time to stop", fmt.Sprintf("Your %d min session is over!", m.game.SessionMin))
			notify.Sound()
			if m.game.Process == "" {
				m.finish(session.StatusCompleted, "Session complete. No process monitored.")
				break
			}
			m.phase = phaseExtra
			m.extraRemaining = time.Duration(m.game.ExtraMin) * time.Minute
			notify.Send("Self-destruct engaged", fmt.Sprintf("You have %d min to save before it closes automatically!", m.game.ExtraMin))
		}

	case phaseExtra:
		if !proc.IsRunning(m.game.Process) {
			m.finish(session.StatusSavedInTime, "You saved and closed the game in time.")
			break
		}
		m.extraRemaining -= time.Second
		if m.extraRemaining <= 0 {
			if proc.IsRunning(m.game.Process) {
				notify.Send("Time's up", fmt.Sprintf("Closing %s now.", m.game.Process))
				_ = proc.Kill(m.game.Process)
				m.finish(session.StatusCompleted, "Time's up. The game was closed automatically.")
			} else {
				m.finish(session.StatusSavedInTime, "You saved and closed the game in time.")
			}
		}
	}

	if m.phase == phaseDone {
		return m, nil
	}
	return m, tickCmd()
}

func (m *Model) finish(status session.Status, msg string) {
	m.phase = phaseDone
	m.finalStatus = status
	m.finalMsg = msg

	entry := session.Entry{
		Game:       m.game.Name,
		Process:    m.game.Process,
		StartedAt:  m.startedAt,
		EndedAt:    time.Now(),
		PlannedMin: m.game.SessionMin,
		ExtraMin:   m.game.ExtraMin,
		PlayedMin:  float64(m.playedSeconds) / 60.0,
		Status:     status,
	}
	m.logErr = session.Append(entry)
}

func (m Model) FinalStatus() session.Status { return m.finalStatus }

func formatMMSS(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	total := int(d.Seconds())
	return fmt.Sprintf("%02d:%02d", total/60, total%60)
}

func formatPlayed(minutes float64) string {
	total := int(minutes + 0.5)
	h, m := total/60, total%60
	if h == 0 {
		return fmt.Sprintf("%dmin", m)
	}
	return fmt.Sprintf("%dh%02dmin", h, m)
}
