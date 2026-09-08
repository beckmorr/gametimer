package tui

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
)

// bigTitle renders name as a small block-letter banner using blockGlyphs (an
// embedded ASCII/Unicode font, no external font file). Falls back to plain
// styled text if the banner would be wider than maxWidth or contains no
// renderable glyphs.
func bigTitle(name, hex, bg string, maxWidth int) string {
	style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(hex)).Background(lipgloss.Color(bg))

	rows := [7]strings.Builder{}
	width := 0
	rendered := false
	for _, r := range strings.ToUpper(name) {
		g := glyphFor(unicode.ToUpper(r))
		if g != blankGlyph {
			rendered = true
		}
		glyphWidth := len(g[0])
		for i := range rows {
			rows[i].WriteString(g[i])
			rows[i].WriteByte(' ')
		}
		width += glyphWidth + 1
	}

	if !rendered || width == 0 || (maxWidth > 0 && width > maxWidth) {
		return style.Render(name)
	}

	lines := make([]string, 7)
	for i, row := range rows {
		s := row.String()
		s = strings.ReplaceAll(s, "#", "█")
		s = strings.ReplaceAll(s, ".", " ")
		lines[i] = style.Render(s)
	}
	return strings.Join(lines, "\n")
}
