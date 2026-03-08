package tui

import (
	"io"
	"math/bits"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
)

// Banner is the braille-art popsicle + FIZZY text banner.
const Banner = "" +
	"⠀⡀⠄⠀⣤⣤⡄⠠⢠⣤⣤⠀⠄⣠⣤⣤⠀⠠⣠⣤⣄⠠⠀⣤⣤⡄⠠⠀⠠⠀⠠⠀⠠⠀⠠⠀⠠⠀⠠⠀⠠⠀⠠⠀⠠⠀⠠⠀⠠⠀⠠⠀⠠⠀⠠⠀⠠⠀⠠⠀⠠⠀⠠⠀⠠⠀⠠⠀⠠⠀⠠⠀⠠⠀⠠⠀⠠⠀⠠⠀\n" +
	"⠀⠀⠄⢸⣿⣿⡇⠠⢸⣿⣿⡇⠀⣿⣿⡿⡇⢈⣿⣿⣿⠀⢸⣿⣿⡇⢀⠈⡀⢈⠀⡈⢀⠈⡀⢈⠀⡈⢀⠈⡀⢈⠀⡈⢀⠈⡀⢈⠀⡈⢀⠈⡀⢈⠀⡈⢀⠈⡀⢈⠀⡈⢀⠈⠠⠈⢀⠈⡀⢈⠀⡈⢀⠈⡀⢈⠀⡈⠠⠀\n" +
	"⠀⢁⠠⢸⣿⡿⡇⠀⢸⣿⡿⡇⠀⣿⣿⢿⡃⠠⣿⣿⣾⠀⢸⣿⣽⡇⠀⠠⠀⠄⠠⠀⣦⣦⣦⣦⣦⣦⣦⣦⣦⠀⢰⣾⣷⡦⠀⠄⠠⠀⠄⠠⠀⠄⠠⠀⠄⠠⠀⠄⠠⠀⠄⠐⢀⠈⡀⠠⠀⠄⠠⠀⠄⠠⠀⠄⠠⠀⠐⠀\n" +
	"⠀⠠⠀⢸⣿⣿⡇⠈⢸⣿⣿⡇⠀⣿⣿⣿⠇⢈⣿⣷⣿⠀⢸⣿⣽⡇⠀⠂⠐⠀⠂⡀⣿⣿⡟⠛⠛⠛⠛⠛⠛⠀⠄⠉⠉⢁⠀⠂⠐⠀⠂⠐⠀⠂⠐⠀⠂⠐⠀⠂⠐⠀⠂⠁⠠⠀⠄⠐⠀⠂⠐⠀⠂⠐⠀⠂⠐⠈⡀⠁\n" +
	"⠀⠂⠈⢸⣿⣯⡇⠠⢸⣿⣯⡇⠀⣿⣿⣾⡇⠐⣿⣯⣿⠀⢸⣿⣽⡇⠀⡁⢈⠀⡁⠀⣿⣿⡇⠀⠄⠂⠀⠂⠐⢀⢸⣿⣿⡇⠀⡁⣿⣿⣿⣿⣿⣿⣿⣿⡃⢈⠀⣿⣿⣿⣿⣿⣿⣿⣿⠅⠈⢿⣿⣿⡀⢁⠈⡀⣽⣿⣿⠃\n" +
	"⠀⡈⠠⢸⣿⡿⡇⠀⢸⣿⡿⡇⠀⣿⣿⣽⡆⠨⣿⣟⣿⠀⢸⣯⣿⡇⠀⠄⠠⠀⠄⠂⣿⣿⢷⣶⣶⣶⣷⣾⡆⢀⢸⣿⣻⡇⠀⡀⠉⡈⠁⣁⣵⣿⣿⠋⠀⠠⠀⢉⠈⢁⢁⣵⣿⡿⠋⠀⠐⠘⣿⣿⣧⠀⠄⢠⣿⣿⠎⠀\n" +
	"⠀⠠⠀⢸⣿⣿⡇⠈⢸⣿⣿⡇⠀⣿⣿⣻⡅⠐⠿⣿⠟⠀⢸⣿⣽⡇⠀⠂⠐⠀⠂⡀⣿⣿⡟⠛⠛⠋⠛⠛⠃⠀⢸⣟⣿⡇⠀⠄⠂⠠⣰⣾⣿⠟⠁⢀⠈⠠⠈⢀⢠⣰⣿⡿⠟⠀⡀⢁⠈⡀⠸⣿⣾⡇⢀⣿⣿⡟⠀⠄\n" +
	"⠀⠂⠈⠸⣿⣯⡇⠠⢸⣿⣯⡇⠀⣿⣿⢿⡃⠠⠀⡢⠀⡁⠘⣿⣽⡇⠀⡁⢈⠀⢁⠀⣿⣿⡇⠀⠐⠀⠂⠐⠀⡁⢸⣿⣟⡇⢀⠐⣠⣶⣿⡟⠃⢀⠐⠀⡐⢀⠈⣠⣾⣿⡗⠋⠀⠄⠠⠀⠄⠀⠂⢹⣿⣷⣼⣿⡿⠀⠐⠀\n" +
	"⠀⡈⢀⠁⠉⡏⠀⠠⢸⣿⡿⡇⠀⣿⣿⣿⠇⠀⠂⢜⠀⡀⠂⠉⡏⠀⠄⠠⠀⠐⢀⠀⣿⣿⡇⠀⡁⢈⠀⡁⠄⠠⢸⣿⣽⡇⠀⡀⣿⣿⣻⣿⣿⣿⣿⣿⡇⠀⢐⣿⣿⣷⣿⣿⣿⣿⣿⡇⢀⠁⡈⠀⢹⣿⣻⣽⠁⢀⠁⠄\n" +
	"⠀⠠⠀⠐⠀⡇⠈⡀⢸⣿⣿⡇⠀⠙⢻⠊⠠⠈⡀⢕⠀⠠⠐⠀⡇⠐⠀⠂⢈⠠⠀⠄⢀⠀⡀⠄⠠⠀⠄⠠⠀⠂⢀⠀⡀⠠⠀⠄⢀⠀⡀⠀⡀⠀⡀⠀⡀⠐⢀⠀⡀⠀⡀⠀⡀⠀⡀⠀⠄⠠⣀⣈⣸⣿⣿⠃⠀⠄⠐⠀\n" +
	"⠀⠐⠈⢀⠁⡇⠠⠀⠸⢿⡯⠃⢀⠁⢸⠀⠂⠠⠀⠪⠀⠐⠀⠁⡇⢀⠁⡈⢀⠠⠐⠀⠄⠠⠀⠄⠂⠐⠀⠂⡀⢁⠠⠀⠄⠐⠀⠂⠠⠀⠄⠂⠀⠂⠀⠂⡀⢈⠀⠠⠀⠂⠀⠂⠀⠂⠀⠂⠐⠰⣿⣿⠿⠟⠁⡀⠂⡀⠁⠄\n" +
	"⠀⢁⠈⢀⠠⠐⠀⡈⠠⠀⡀⠄⠠⠐⠀⠂⠁⠐⠀⠂⡈⢀⠁⠐⡀⠄⠠⠀⠄⢀⠐⠀⠂⠐⠀⠂⡀⢁⠈⡀⠄⠠⠀⠐⢀⠈⡀⢁⠐⢀⠐⠀⡁⢈⠀⡁⠀⠄⠐⠀⠂⠁⡈⢀⠁⡈⢀⠁⡈⢀⠀⡀⠄⠂⠁⠀⠄⠠⠈⠀"

// BannerLines is the number of lines in the Banner.
const BannerLines = 12

// BannerWidth is the rune count of the widest line.
const BannerWidth = 80

// CompactBannerWidth is the visual width of the compact banner's widest line.
const CompactBannerWidth = 41

// BannerStyle selects the banner layout.
type BannerStyle int

const (
	// BannerCompact renders popsicle art with "Fizzy\nby 37signals" to the right.
	BannerCompact BannerStyle = iota
	// BannerFull renders the full 80-column banner with brighter colors and tagline.
	BannerFull
)

// GetBannerStyle returns the configured banner style from FIZZY_BANNER.
func GetBannerStyle() BannerStyle {
	if os.Getenv("FIZZY_BANNER") == "full" {
		return BannerFull
	}
	return BannerCompact
}

const blankBraille = '\u2800' // ⠀

// isDense returns true for braille characters with 3+ raised dots,
// which form the visible popsicle and text shapes. Characters with
// 0-2 dots are background texture — rendered as blank in colored mode
// to eliminate speckling and produce crisp shapes.
func isDense(ch rune) bool {
	if ch < 0x2800 || ch > 0x28FF {
		return false
	}
	return bits.OnesCount(uint(ch-0x2800)) > 2 // #nosec G115 -- range checked above
}

// PopsicleColors are the 5 brand colors, left to right.
var PopsicleColors = [5]lipgloss.Color{
	"#3B82F6", // Blue
	"#EC4899", // Magenta
	"#F97316", // Orange
	"#84CC16", // Lime
	"#06B6D3", // Cyan
}

// popsicleTints are the mid-trail tints (50% blend towards white).
var popsicleTints = [5]lipgloss.Color{
	"#9DC1FB", // Blue tint
	"#F5A3CC", // Magenta tint
	"#FCB98B", // Orange tint
	"#C2E58B", // Lime tint
	"#83DBEA", // Cyan tint
}

// brightPopsicleColors are more saturated versions for the full banner style.
var brightPopsicleColors = [5]lipgloss.Color{
	"#60A5FA", // Brighter blue
	"#F472B6", // Brighter magenta
	"#FB923C", // Brighter orange
	"#A3E635", // Brighter lime
	"#22D3EE", // Brighter cyan
}

// popsicleZone returns the zone index (0-5) for a column position.
// Zones 0-4 are popsicle sticks, zone 5 is the FIZZY text.
func popsicleZone(col int) int {
	switch {
	case col <= 7:
		return 0
	case col <= 12:
		return 1
	case col <= 16:
		return 2
	case col <= 21:
		return 3
	case col <= 25:
		return 4
	default:
		return 5
	}
}

// bannerGrid returns the banner as a rune grid and its line count.
func bannerGrid() ([][]rune, int) {
	lines := strings.Split(Banner, "\n")
	grid := make([][]rune, len(lines))
	for i, line := range lines {
		grid[i] = []rune(line)
	}
	return grid, len(lines)
}

// noColor returns true when color output should be suppressed.
func noColor() bool {
	_, ok := os.LookupEnv("NO_COLOR")
	return ok
}

// rendererForWriter returns a lipgloss renderer bound to the given writer,
// so that color-profile detection matches the actual output destination.
func rendererForWriter(w io.Writer) *lipgloss.Renderer {
	return lipgloss.NewRenderer(w)
}

// termWidth returns the terminal width for w, or 0 if unknown.
func termWidth(w io.Writer) int {
	if f, ok := w.(*os.File); ok {
		if width, _, err := term.GetSize(f.Fd()); err == nil {
			return width
		}
	}
	return 0
}

// bannerVisualRows returns the number of terminal rows the full banner
// occupies, accounting for soft-wrapping when the terminal is narrower
// than BannerWidth.
func bannerVisualRows(cols int) int {
	if cols <= 0 || cols >= BannerWidth {
		return BannerLines
	}
	rowsPerLine := (BannerWidth + cols - 1) / cols
	return BannerLines * rowsPerLine
}

// compactVisualRows computes the visual row count for the compact banner,
// accounting for the fact that most lines are popsicleColLimit (26) wide
// while two lines have appended text making them wider.
func compactVisualRows(cols int) int {
	if cols <= 0 || cols >= CompactBannerWidth {
		return BannerLines
	}
	total := 0
	for i := 0; i < BannerLines; i++ {
		w := popsicleColLimit
		switch i {
		case BannerLines - 3:
			w = popsicleColLimit + 3 + 5 // "   Fizzy"
		case BannerLines - 2:
			w = CompactBannerWidth // "   by 37signals"
		}
		total += max(1, (w+cols-1)/cols)
	}
	return total
}

// fullBannerPlain returns the full banner as plain text with the tagline overlay.
func fullBannerPlain() string {
	const tagRow = 10
	const tagCol = 55
	tagline := []rune("by 37signals")
	lines := strings.Split(Banner, "\n")
	var b strings.Builder
	for i, line := range lines {
		if i > 0 {
			b.WriteByte('\n')
		}
		if i == tagRow {
			runes := []rune(line)
			b.WriteString(string(runes[:tagCol]))
			b.WriteString(string(tagline))
			end := tagCol + len(tagline)
			if end < len(runes) {
				b.WriteString(string(runes[end:]))
			}
		} else {
			b.WriteString(line)
		}
	}
	return b.String()
}

// RenderBanner returns the banner styled according to the configured BannerStyle.
// When w is nil, returns plain text. Otherwise creates a renderer bound to w
// so that color capabilities are detected correctly.
func RenderBanner(w io.Writer) string {
	style := GetBannerStyle()
	if w == nil {
		if style == BannerFull {
			return fullBannerPlain()
		}
		return compactBannerPlain()
	}
	re := rendererForWriter(w)
	if style == BannerFull {
		return renderFullStyledBannerWith(re)
	}
	return renderCompactBannerWith(re)
}

// renderBannerWith renders the full banner using the given renderer's styles.
// This is the original rendering used by the animation system.
func renderBannerWith(re *lipgloss.Renderer) string {
	var zoneStyles [6]lipgloss.Style
	for i := 0; i < 5; i++ {
		zoneStyles[i] = re.NewStyle().Foreground(PopsicleColors[i])
	}
	zoneStyles[5] = re.NewStyle() // default foreground for text

	var b strings.Builder
	for i, line := range strings.Split(Banner, "\n") {
		if i > 0 {
			b.WriteByte('\n')
		}
		runes := []rune(line)
		j := 0
		for j < len(runes) {
			z := popsicleZone(j)
			k := j + 1
			for k < len(runes) && popsicleZone(k) == z {
				k++
			}
			var run strings.Builder
			for x := j; x < k; x++ {
				if isDense(runes[x]) {
					run.WriteRune(runes[x])
				} else {
					run.WriteRune(blankBraille)
				}
			}
			b.WriteString(zoneStyles[z].Render(run.String()))
			j = k
		}
	}
	return b.String()
}

// ---------------------------------------------------------------------------
// Compact banner (popsicle art + side text)
// ---------------------------------------------------------------------------

const popsicleColLimit = 26

// compactBannerPlain returns the compact banner as plain text (no ANSI).
func compactBannerPlain() string {
	lines := strings.Split(Banner, "\n")
	numLines := len(lines)
	var b strings.Builder
	for i, line := range lines {
		if i > 0 {
			b.WriteByte('\n')
		}
		runes := []rune(line)
		if len(runes) > popsicleColLimit {
			runes = runes[:popsicleColLimit]
		}
		b.WriteString(string(runes))
		switch i {
		case numLines - 3:
			b.WriteString("   Fizzy")
		case numLines - 2:
			b.WriteString("   by 37signals")
		}
	}
	return b.String()
}

// renderCompactBannerWith renders popsicle art with styled "Fizzy\nby 37signals".
func renderCompactBannerWith(re *lipgloss.Renderer) string {
	var zoneStyles [5]lipgloss.Style
	for i := 0; i < 5; i++ {
		zoneStyles[i] = re.NewStyle().Foreground(PopsicleColors[i])
	}

	fizzy := renderFizzyText(re)
	by37 := re.NewStyle().Foreground(lipgloss.Color("#888888")).Render("by 37signals")

	lines := strings.Split(Banner, "\n")
	numLines := len(lines)

	var b strings.Builder
	for i, line := range lines {
		if i > 0 {
			b.WriteByte('\n')
		}
		runes := []rune(line)
		if len(runes) > popsicleColLimit {
			runes = runes[:popsicleColLimit]
		}
		j := 0
		for j < len(runes) {
			z := popsicleZone(j)
			if z >= 5 {
				break
			}
			k := j + 1
			for k < len(runes) && popsicleZone(k) == z {
				k++
			}
			var run strings.Builder
			for x := j; x < k; x++ {
				if isDense(runes[x]) {
					run.WriteRune(runes[x])
				} else {
					run.WriteRune(blankBraille)
				}
			}
			b.WriteString(zoneStyles[z].Render(run.String())) // #nosec G602 -- z < 5 guarded above
			j = k
		}
		switch i {
		case numLines - 3:
			b.WriteString("   ")
			b.WriteString(fizzy)
		case numLines - 2:
			b.WriteString("   ")
			b.WriteString(by37)
		}
	}
	return b.String()
}

// renderFizzyText returns "Fizzy" with each letter in its tint color, bold.
func renderFizzyText(re *lipgloss.Renderer) string {
	var b strings.Builder
	for i, l := range []string{"F", "i", "z", "z", "y"} {
		b.WriteString(re.NewStyle().Foreground(popsicleTints[i]).Bold(true).Render(l))
	}
	return b.String()
}

// ---------------------------------------------------------------------------
// Full styled banner (brighter colors + bold + tagline)
// ---------------------------------------------------------------------------

// renderFullStyledBannerWith renders the full 80-col banner with brighter
// colors, bold, and "by 37signals" embedded at the y-descender baseline.
func renderFullStyledBannerWith(re *lipgloss.Renderer) string {
	var zoneStyles [6]lipgloss.Style
	for i := 0; i < 5; i++ {
		zoneStyles[i] = re.NewStyle().Foreground(brightPopsicleColors[i]).Bold(true)
	}
	zoneStyles[5] = re.NewStyle().Bold(true)

	dim := re.NewStyle().Foreground(lipgloss.Color("#888888"))

	const tagCol = 55
	const tagRow = 10
	tagline := "by 37signals"

	var b strings.Builder
	for i, line := range strings.Split(Banner, "\n") {
		if i > 0 {
			b.WriteByte('\n')
		}
		runes := []rune(line)
		doOverlay := i == tagRow
		overlayEnd := tagCol + len([]rune(tagline))

		j := 0
		for j < len(runes) {
			if doOverlay && j == tagCol {
				b.WriteString(dim.Render(tagline))
				j = overlayEnd
				if j >= len(runes) {
					break
				}
				continue
			}

			z := popsicleZone(j)
			k := j + 1
			limit := len(runes)
			if doOverlay && j < tagCol && tagCol < limit {
				limit = tagCol
			}
			for k < limit && popsicleZone(k) == z {
				k++
			}

			var run strings.Builder
			for x := j; x < k; x++ {
				if isDense(runes[x]) {
					run.WriteRune(runes[x])
				} else {
					run.WriteRune(blankBraille)
				}
			}
			b.WriteString(zoneStyles[z].Render(run.String()))
			j = k
		}
	}
	return b.String()
}
