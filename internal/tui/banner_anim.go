package tui

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

const (
	paintInterval = 25 * time.Millisecond
	traceFrames   = 30
	zoneStagger   = 3 // frame offset between popsicle zones
)

// paintCell is a single cell position in the banner grid.
type paintCell struct {
	row, col int
}

// zoneTrail holds the color trail styles for a single zone.
type zoneTrail struct {
	styles  []lipgloss.Style
	settled int
}

// animations maps strategy names to trace functions.
var animations = map[string]func([][]rune, int) []paintCell{
	"paint": tracePaint,
	"pop":   tracePop,
}

// DefaultAnimation is the animation used when none is specified.
const DefaultAnimation = "pop"

// buildZoneTrails creates per-zone color trails for animation using the
// given renderer so that color capabilities match the output destination.
// settledColors controls the final resting color for each popsicle zone.
// When bold is true, settled styles include Bold(true) to match the full
// banner's static render.
func buildZoneTrails(re *lipgloss.Renderer, settledColors [5]lipgloss.Color, bold bool) [6]zoneTrail {
	var trails [6]zoneTrail

	// Popsicle zones (0-4): white flash -> tint -> settled color
	for i := 0; i < 5; i++ {
		settled := re.NewStyle().Foreground(settledColors[i])
		if bold {
			settled = settled.Bold(true)
		}
		trails[i] = zoneTrail{
			styles: []lipgloss.Style{
				re.NewStyle().Foreground(lipgloss.Color("#FFFFFF")),
				re.NewStyle().Foreground(popsicleTints[i]),
				settled,
			},
			settled: 2,
		}
	}

	// Text zone (5): dim -> foreground
	textSettled := re.NewStyle()
	if bold {
		textSettled = textSettled.Bold(true)
	}
	trails[5] = zoneTrail{
		styles: []lipgloss.Style{
			re.NewStyle().Foreground(lipgloss.Color("#666666")),
			textSettled,
		},
		settled: 1,
	}

	return trails
}

// AnimateBanner draws the banner with animation to w.
// Falls back to static plain render when not a TTY or colors are disabled.
func AnimateBanner(w io.Writer) {
	style := GetBannerStyle()
	compact := style == BannerCompact

	if !isWriterTTY(w) || noColor() {
		if compact {
			fmt.Fprint(w, "\n"+compactBannerPlain()+"\n")
		} else {
			fmt.Fprint(w, "\n"+fullBannerPlain()+"\n")
		}
		return
	}

	re := rendererForWriter(w)

	name := os.Getenv("FIZZY_ANIM")
	if name == "none" {
		if compact {
			fmt.Fprint(w, "\n"+renderCompactBannerWith(re)+"\n")
		} else {
			fmt.Fprint(w, "\n"+renderFullStyledBannerWith(re)+"\n")
		}
		return
	}
	if name == "" {
		name = DefaultAnimation
	}

	traceFn, ok := animations[name]
	if !ok {
		traceFn = animations[DefaultAnimation]
	}

	grid, numLines := bannerGrid()

	// For full mode, mark tagline positions as dense so the trace includes them.
	if !compact {
		markTaglineDense(grid)
	}

	order := traceFn(grid, numLines)

	// For compact mode, truncate grid to popsicle columns and filter trace.
	if compact {
		for i := range grid {
			if len(grid[i]) > popsicleColLimit {
				grid[i] = grid[i][:popsicleColLimit]
			}
		}
		filtered := order[:0]
		for _, c := range order {
			if c.col < popsicleColLimit {
				filtered = append(filtered, c)
			}
		}
		order = filtered
	}

	if len(order) == 0 {
		if compact {
			fmt.Fprint(w, "\n"+renderCompactBannerWith(re)+"\n")
		} else {
			fmt.Fprint(w, "\n"+renderFullStyledBannerWith(re)+"\n")
		}
		return
	}

	settledColors := PopsicleColors
	bold := false
	if !compact {
		settledColors = brightPopsicleColors
		bold = true
	}
	trails := buildZoneTrails(re, settledColors, bold)
	batchSize := max(1, len(order)/traceFrames)
	numBatches := (len(order) + batchSize - 1) / batchSize

	maxSettled := 0
	for _, t := range trails {
		if t.settled > maxSettled {
			maxSettled = t.settled
		}
	}

	cols := termWidth(w)
	var visRows int
	if compact {
		visRows = compactVisualRows(cols)
	} else {
		visRows = bannerVisualRows(cols)
	}
	revealFrame := makeRevealGrid(grid, numLines)

	// Pre-compute side/overlay text and styles.
	var compactFizzy, compactBy37 string
	var fullDimStyle lipgloss.Style
	if compact {
		compactFizzy = renderFizzyText(re)
		compactBy37 = re.NewStyle().Foreground(lipgloss.Color("#888888")).Render("by 37signals")
	} else {
		fullDimStyle = re.NewStyle().Foreground(lipgloss.Color("#888888"))
	}

	fmt.Fprintln(w) // leading blank line
	totalFrames := numBatches + maxSettled
	for frame := 0; frame < totalFrames; frame++ {
		if frame < numBatches {
			start := frame * batchSize
			end := min(start+batchSize, len(order))
			for _, c := range order[start:end] {
				revealFrame[c.row][c.col] = frame
			}
		}

		if frame > 0 {
			fmt.Fprintf(w, "\033[%dA", visRows)
		}
		if compact {
			renderCompactPaintFrame(w, grid, revealFrame, frame, trails, numLines, compactFizzy, compactBy37)
		} else {
			renderFullPaintFrame(w, grid, revealFrame, frame, trails, numLines, fullDimStyle, cols)
		}
		time.Sleep(paintInterval)
	}
}

// ---------------------------------------------------------------------------
// Non-blocking animation
// ---------------------------------------------------------------------------

// AnimWriter is a mutex-protected io.Writer that tracks terminal rows written
// below the animation area. The animation goroutine uses this count to
// cursor-up past both the logo and any caller output, then cursor-down
// to restore the caller's write position.
type AnimWriter struct {
	w          io.Writer
	mu         sync.Mutex
	linesBelow int
	done       chan struct{}
	numLines   int // height of the animation area
	cols       int // terminal width; 0 = fall back to \n counting
	pendingW   int // visual width of the current unterminated line
}

// Write writes p through to the underlying writer and counts visual rows.
func (aw *AnimWriter) Write(p []byte) (int, error) {
	aw.mu.Lock()
	defer aw.mu.Unlock()
	n, err := aw.w.Write(p)
	rows, pw := visualLines(string(p[:n]), aw.cols, aw.pendingW)
	aw.linesBelow += rows
	aw.pendingW = pw
	return n, err
}

// Wait blocks until the animation goroutine finishes. Idempotent.
func (aw *AnimWriter) Wait() {
	<-aw.done
}

// AnimateBannerAsync starts the animation in a background goroutine and
// returns a writer for subsequent output plus a wait function. Output written
// through the returned writer appears below the animating banner. Call wait
// before any interactive prompts or output that bypasses the returned writer.
func AnimateBannerAsync(w io.Writer) (io.Writer, func()) {
	style := GetBannerStyle()
	compact := style == BannerCompact

	if !isWriterTTY(w) || noColor() {
		if compact {
			fmt.Fprint(w, "\n"+compactBannerPlain()+"\n")
		} else {
			fmt.Fprint(w, "\n"+fullBannerPlain()+"\n")
		}
		return w, func() {}
	}

	re := rendererForWriter(w)

	name := os.Getenv("FIZZY_ANIM")
	if name == "none" {
		if compact {
			fmt.Fprint(w, "\n"+renderCompactBannerWith(re)+"\n")
		} else {
			fmt.Fprint(w, "\n"+renderFullStyledBannerWith(re)+"\n")
		}
		return w, func() {}
	}
	if name == "" {
		name = DefaultAnimation
	}

	traceFn, ok := animations[name]
	if !ok {
		traceFn = animations[DefaultAnimation]
	}

	grid, numLines := bannerGrid()

	if !compact {
		markTaglineDense(grid)
	}

	order := traceFn(grid, numLines)

	if compact {
		for i := range grid {
			if len(grid[i]) > popsicleColLimit {
				grid[i] = grid[i][:popsicleColLimit]
			}
		}
		filtered := order[:0]
		for _, c := range order {
			if c.col < popsicleColLimit {
				filtered = append(filtered, c)
			}
		}
		order = filtered
	}

	if len(order) == 0 {
		if compact {
			fmt.Fprint(w, "\n"+renderCompactBannerWith(re)+"\n")
		} else {
			fmt.Fprint(w, "\n"+renderFullStyledBannerWith(re)+"\n")
		}
		return w, func() {}
	}

	settledColors := PopsicleColors
	bold := false
	if !compact {
		settledColors = brightPopsicleColors
		bold = true
	}
	trails := buildZoneTrails(re, settledColors, bold)
	batchSize := max(1, len(order)/traceFrames)
	numBatches := (len(order) + batchSize - 1) / batchSize

	maxSettled := 0
	for _, t := range trails {
		if t.settled > maxSettled {
			maxSettled = t.settled
		}
	}

	cols := termWidth(w)
	var visRows int
	if compact {
		visRows = compactVisualRows(cols)
	} else {
		visRows = bannerVisualRows(cols)
	}
	revealFrame := makeRevealGrid(grid, numLines)

	var compactFizzy, compactBy37 string
	var fullDimStyle lipgloss.Style
	if compact {
		compactFizzy = renderFizzyText(re)
		compactBy37 = re.NewStyle().Foreground(lipgloss.Color("#888888")).Render("by 37signals")
	} else {
		fullDimStyle = re.NewStyle().Foreground(lipgloss.Color("#888888"))
	}

	// Render frame 0 synchronously
	fmt.Fprintln(w) // leading blank line
	end := min(batchSize, len(order))
	for _, c := range order[:end] {
		revealFrame[c.row][c.col] = 0
	}
	if compact {
		renderCompactPaintFrame(w, grid, revealFrame, 0, trails, numLines, compactFizzy, compactBy37)
	} else {
		renderFullPaintFrame(w, grid, revealFrame, 0, trails, numLines, fullDimStyle, cols)
	}

	aw := &AnimWriter{
		w:        w,
		done:     make(chan struct{}),
		numLines: visRows,
		cols:     cols,
	}

	totalFrames := numBatches + maxSettled
	go func() {
		defer close(aw.done)

		for frame := 1; frame < totalFrames; frame++ {
			if frame < numBatches {
				start := frame * batchSize
				end := min(start+batchSize, len(order))
				for _, c := range order[start:end] {
					revealFrame[c.row][c.col] = frame
				}
			}

			aw.mu.Lock()
			fmt.Fprintf(w, "\033[%dA", visRows+aw.linesBelow)
			if compact {
				renderCompactPaintFrame(w, grid, revealFrame, frame, trails, numLines, compactFizzy, compactBy37)
			} else {
				renderFullPaintFrame(w, grid, revealFrame, frame, trails, numLines, fullDimStyle, cols)
			}
			if aw.linesBelow > 0 {
				fmt.Fprintf(w, "\033[%dB", aw.linesBelow)
			}
			if aw.pendingW > 0 {
				fmt.Fprintf(w, "\033[%dC", aw.pendingW)
			}
			aw.mu.Unlock()
			time.Sleep(paintInterval)
		}
	}()

	return aw, aw.Wait
}

// ---------------------------------------------------------------------------
// Animation strategies
// ---------------------------------------------------------------------------

// tracePaint reveals popsicles top-to-bottom with stagger cascade,
// then FIZZY text as a diagonal curtain unfurling left-to-right.
func tracePaint(grid [][]rune, numLines int) []paintCell {
	return traceWithOrder(grid, numLines, false, true)
}

// tracePop reveals popsicles bottom-to-top with stagger cascade,
// then FIZZY text left-to-right.
func tracePop(grid [][]rune, numLines int) []paintCell {
	return traceWithOrder(grid, numLines, true, false)
}

func traceWithOrder(grid [][]rune, numLines int, reverse bool, curtainText bool) []paintCell {
	// Collect non-blank cells per zone, sorted top-to-bottom (natural grid order)
	var zoneCells [6][]paintCell
	for r := 0; r < numLines; r++ {
		for c, ch := range grid[r] {
			if isDense(ch) {
				zoneCells[popsicleZone(c)] = append(zoneCells[popsicleZone(c)], paintCell{r, c})
			}
		}
	}

	// For pop strategy, reverse within-zone order (bottom-to-top)
	if reverse {
		for z := 0; z < 5; z++ {
			cells := zoneCells[z]
			for i, j := 0, len(cells)-1; i < j; i, j = i+1, j-1 {
				cells[i], cells[j] = cells[j], cells[i]
			}
		}
	}

	// Sort text zone cells
	if curtainText {
		// Radial unfurl from top-left corner of text zone. The wavefront
		// is a quarter-ellipse expanding outward — left edge drops first,
		// then the rest curls around to the right. Stretch factor makes
		// row distance comparable to column distance so the reveal
		// sweeps both down and across at a balanced rate.
		stretch := max(1, (BannerWidth-26)/numLines)
		sort.Slice(zoneCells[5], func(i, j int) bool {
			ci, cj := zoneCells[5][i], zoneCells[5][j]
			dxi, dyi := ci.col-26, ci.row*stretch
			dxj, dyj := cj.col-26, cj.row*stretch
			di := dxi*dxi + dyi*dyi
			dj := dxj*dxj + dyj*dyj
			if di != dj {
				return di < dj
			}
			return ci.col < cj.col
		})
	} else {
		// Column sweep: left-to-right, top-to-bottom within each column.
		sort.Slice(zoneCells[5], func(i, j int) bool {
			if zoneCells[5][i].col != zoneCells[5][j].col {
				return zoneCells[5][i].col < zoneCells[5][j].col
			}
			return zoneCells[5][i].row < zoneCells[5][j].row
		})
	}

	// Compute popsicle order span
	maxPopsicleEnd := 0
	for z := 0; z < 5; z++ {
		end := z*zoneStagger + len(zoneCells[z])
		if end > maxPopsicleEnd {
			maxPopsicleEnd = end
		}
	}

	// Build ordered cell list with logical frame numbers
	type orderedCell struct {
		cell  paintCell
		order int
	}
	var all []orderedCell
	for z := 0; z < 5; z++ {
		offset := z * zoneStagger
		for j, c := range zoneCells[z] {
			all = append(all, orderedCell{c, offset + j})
		}
	}

	// Simultaneous: scale text indices to span [0, maxPopsicleEnd]
	// so popsicles and text reveal together.
	n := len(zoneCells[5])
	for j, c := range zoneCells[5] {
		order := 0
		if n > 1 {
			order = j * maxPopsicleEnd / (n - 1)
		}
		all = append(all, orderedCell{c, order})
	}

	sort.SliceStable(all, func(i, j int) bool {
		return all[i].order < all[j].order
	})

	result := make([]paintCell, len(all))
	for i, oc := range all {
		result[i] = oc.cell
	}
	return result
}

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

// markTaglineDense replaces braille chars at the tagline overlay positions
// with full braille blocks so the trace includes them in the reveal order.
func markTaglineDense(grid [][]rune) {
	const tagRow = 10
	const tagCol = 55
	const tagLen = 12 // len("by 37signals")
	if tagRow < len(grid) {
		for i := 0; i < tagLen; i++ {
			col := tagCol + i
			if col < len(grid[tagRow]) {
				grid[tagRow][col] = '\u28FF' // ⣿ — all 8 dots raised
			}
		}
	}
}

// makeRevealGrid creates a frame-index grid initialized to -1 (unrevealed).
func makeRevealGrid(grid [][]rune, numLines int) [][]int {
	revealFrame := make([][]int, numLines)
	for r, row := range grid {
		revealFrame[r] = make([]int, len(row))
		for c := range revealFrame[r] {
			revealFrame[r][c] = -1
		}
	}
	return revealFrame
}

// renderFullPaintFrame renders one animation frame for the full banner,
// overlaying revealed tagline characters at the y-descender baseline.
// Uses ANSI CHA to overwrite the tagline region after the line is rendered.
// The overlay is skipped when the terminal is narrower than BannerWidth
// because soft-wrapping would place the CHA on the wrong physical row.
func renderFullPaintFrame(w io.Writer, grid [][]rune, revealFrame [][]int, frame int, trails [6]zoneTrail, numLines int, dimStyle lipgloss.Style, cols int) {
	const tagRow = 10
	const tagCol = 55
	tagline := []rune("by 37signals")
	canOverlay := cols == 0 || cols >= BannerWidth

	for r := 0; r < numLines; r++ {
		line := renderPaintLine(grid[r], revealFrame[r], frame, trails)
		if r == tagRow && canOverlay {
			// Count revealed tagline chars (contiguous from left)
			revealed := 0
			for i := range tagline {
				col := tagCol + i
				if col < len(revealFrame[r]) && revealFrame[r][col] >= 0 {
					revealed = i + 1
				}
			}
			if revealed > 0 {
				fmt.Fprintf(w, "\r%s\033[K\033[%dG%s\n", line, tagCol+1, dimStyle.Render(string(tagline[:revealed])))
			} else {
				fmt.Fprintf(w, "\r%s\033[K\n", line)
			}
		} else {
			fmt.Fprintf(w, "\r%s\033[K\n", line)
		}
	}
}

// renderPaintLine renders a single line, grouping consecutive characters at the
// same zone and color stage into single styled runs to minimize ANSI escapes.
func renderPaintLine(row []rune, dist []int, frame int, trails [6]zoneTrail) string {
	var b strings.Builder
	i := 0
	n := len(row)

	for i < n {
		ch := row[i]
		if !isDense(ch) || dist[i] < 0 {
			// Run of blank/sparse/unrevealed characters
			j := i + 1
			for j < n && (!isDense(row[j]) || dist[j] < 0) {
				j++
			}
			for k := i; k < j; k++ {
				b.WriteRune(blankBraille)
			}
			i = j
			continue
		}

		// Painted cell — determine zone and age stage
		z := popsicleZone(i)
		trail := trails[z]
		age := frame - dist[i]
		if age > trail.settled {
			age = trail.settled
		}

		// Group consecutive painted cells with same zone AND stage
		j := i + 1
		for j < n && isDense(row[j]) && dist[j] >= 0 {
			jz := popsicleZone(j)
			ja := frame - dist[j]
			if ja > trails[jz].settled {
				ja = trails[jz].settled
			}
			if jz != z || ja != age {
				break
			}
			j++
		}

		var run strings.Builder
		for k := i; k < j; k++ {
			run.WriteRune(row[k])
		}
		b.WriteString(trail.styles[age].Render(run.String()))
		i = j
	}

	return b.String()
}

// renderCompactPaintFrame renders one animation frame for the compact banner,
// appending "Fizzy" and "by 37signals" to the appropriate rows.
func renderCompactPaintFrame(w io.Writer, grid [][]rune, revealFrame [][]int, frame int, trails [6]zoneTrail, numLines int, fizzyText, by37Text string) {
	fizzyRow := numLines - 3
	by37Row := numLines - 2
	for r := 0; r < numLines; r++ {
		line := renderPaintLine(grid[r], revealFrame[r], frame, trails)
		switch r {
		case fizzyRow:
			fmt.Fprintf(w, "\r%s   %s\033[K\n", line, fizzyText)
		case by37Row:
			fmt.Fprintf(w, "\r%s   %s\033[K\n", line, by37Text)
		default:
			fmt.Fprintf(w, "\r%s\033[K\n", line)
		}
	}
}

// visualLines counts terminal rows consumed by s, accounting for soft-wrapping
// at cols. pendingW is the visual width already on the current line from a
// previous write. Returns the number of rows and the updated pending width.
func visualLines(s string, cols, pendingW int) (rows, newPendingW int) {
	if cols <= 0 {
		// Terminal width unknown: count hard newlines and track the visual
		// width on the last line so async animations can restore the cursor
		// column correctly.
		rows = strings.Count(s, "\n")
		for _, seg := range strings.SplitAfter(s, "\n") {
			if seg == "" {
				continue
			}
			if strings.HasSuffix(seg, "\n") {
				pendingW = 0
			} else {
				pendingW += lipgloss.Width(seg)
			}
		}
		return rows, pendingW
	}

	for _, seg := range strings.SplitAfter(s, "\n") {
		if seg == "" {
			continue
		}
		w := lipgloss.Width(strings.TrimSuffix(seg, "\n"))
		total := pendingW + w

		if strings.HasSuffix(seg, "\n") {
			rows += max(1, (total+cols-1)/cols)
			pendingW = 0
		} else {
			// Partial line: count wraps, update column position.
			// Use > (not >=) because terminals defer wrap at the right
			// margin — the cursor stays at cols until the next printable
			// character arrives, so exact-fit doesn't consume a row yet.
			if total > cols {
				wrapped := (total - 1) / cols
				rows += wrapped
				pendingW = total - wrapped*cols
			} else {
				pendingW = total
			}
		}
	}
	return rows, pendingW
}

// isWriterTTY returns true if the writer is backed by a terminal file descriptor.
func isWriterTTY(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
	}
	return false
}
