package tui

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestBannerDimensions(t *testing.T) {
	lines := strings.Split(Banner, "\n")
	if len(lines) != BannerLines {
		t.Errorf("expected %d lines, got %d", BannerLines, len(lines))
	}

	maxWidth := 0
	for _, line := range lines {
		w := len([]rune(line))
		if w > maxWidth {
			maxWidth = w
		}
	}
	if maxWidth != BannerWidth {
		t.Errorf("expected width %d, got %d", BannerWidth, maxWidth)
	}
}

func TestIsDense(t *testing.T) {
	if isDense(blankBraille) {
		t.Error("blank braille should not be dense")
	}
	// 1-dot characters: background texture
	for _, ch := range []rune{'⠠', '⠈', '⡀', '⠄', '⢀', '⠐', '⠂', '⠁'} {
		if isDense(ch) {
			t.Errorf("%c (1 dot) should not be dense", ch)
		}
	}
	// 2-dot characters: still background
	for _, ch := range []rune{'⠉', '⢈', '⡈', '⠊'} {
		if isDense(ch) {
			t.Errorf("%c (2 dots) should not be dense", ch)
		}
	}
	// 3+ dot characters: visible shapes
	for _, ch := range []rune{'⣤', '⣿', '⢸', '⡇', '⠋', '⡃', '⠪'} {
		if !isDense(ch) {
			t.Errorf("%c (3+ dots) should be dense", ch)
		}
	}
}

func TestPopsicleZoneBoundaries(t *testing.T) {
	tests := []struct {
		col  int
		zone int
	}{
		{0, 0}, {7, 0},
		{8, 1}, {12, 1},
		{13, 2}, {16, 2},
		{17, 3}, {21, 3},
		{22, 4}, {25, 4},
		{26, 5}, {50, 5}, {79, 5},
	}
	for _, tt := range tests {
		if got := popsicleZone(tt.col); got != tt.zone {
			t.Errorf("popsicleZone(%d) = %d, want %d", tt.col, got, tt.zone)
		}
	}
}

func TestRenderBannerNoColor(t *testing.T) {
	rendered := RenderBanner(nil)
	if strings.Contains(rendered, "\x1b[") {
		t.Error("no-color render should not contain ANSI escapes")
	}
	if !strings.Contains(rendered, "\u2800") {
		t.Error("render should contain braille content")
	}
}

func TestRenderBannerWithColor(t *testing.T) {
	var buf bytes.Buffer
	rendered := RenderBanner(&buf)
	// Verify structure is preserved: same number of lines
	renderedLines := strings.Split(rendered, "\n")
	bannerLines := strings.Split(Banner, "\n")
	if len(renderedLines) != len(bannerLines) {
		t.Errorf("colored render has %d lines, want %d", len(renderedLines), len(bannerLines))
	}
	// Content should still contain braille characters
	if !strings.Contains(rendered, "\u2800") {
		t.Error("colored render should contain braille content")
	}
}

func TestRenderBannerWithTrueColorEmitsANSI(t *testing.T) {
	// Force TrueColor profile so color output doesn't depend on the test
	// runner's terminal capabilities. SetColorProfile overrides the
	// auto-detected profile, whereas termenv.WithProfile is overridden
	// by lipgloss's own detection when the writer isn't a TTY.
	var buf bytes.Buffer
	re := lipgloss.NewRenderer(&buf)
	re.SetColorProfile(termenv.TrueColor)
	rendered := renderBannerWith(re)
	if !strings.Contains(rendered, "\x1b[") {
		t.Error("TrueColor render should contain ANSI escape sequences")
	}
	// Verify the dense cells are colored but sparse cells are blank braille
	lines := strings.Split(Banner, "\n")
	renderedLines := strings.Split(rendered, "\n")
	if len(renderedLines) != len(lines) {
		t.Fatalf("rendered %d lines, want %d", len(renderedLines), len(lines))
	}
}

func TestDenseZone5GoldenMask(t *testing.T) {
	// Golden count of dense cells in zone 5. This pins the visual shape of
	// the FIZZY wordmark so that changes to isDense or the Banner constant
	// are caught. The count includes a small number of cells on rows 9-10
	// where popsicle stick art extends past column 25 into the text zone.
	grid, numLines := bannerGrid()

	var count int
	rowCounts := make(map[int]int)
	for r := 0; r < numLines; r++ {
		for c, ch := range grid[r] {
			if popsicleZone(c) == 5 && isDense(ch) {
				count++
				rowCounts[r]++
			}
		}
	}

	const expectedTotal = 167
	if count != expectedTotal {
		t.Errorf("zone-5 dense cell count = %d, want %d (golden mask changed)", count, expectedTotal)
	}

	// Rows 0, 1, 11 should have zero dense zone-5 cells (pure background).
	for _, r := range []int{0, 1, 11} {
		if rowCounts[r] != 0 {
			t.Errorf("row %d has %d dense zone-5 cells, want 0", r, rowCounts[r])
		}
	}

	// Rows 9-10 should have a small number of cells (popsicle stick bleed).
	stickBleed := rowCounts[9] + rowCounts[10]
	if stickBleed == 0 || stickBleed > 10 {
		t.Errorf("rows 9-10 have %d dense zone-5 cells, expected 1-10 (popsicle stick bleed)", stickBleed)
	}
}

func TestAnimateBannerNonTTY(t *testing.T) {
	var buf bytes.Buffer
	AnimateBanner(&buf)

	output := buf.String()
	if !strings.Contains(output, "\u2800") {
		t.Error("non-TTY output should contain braille content")
	}
	// Non-TTY should not contain cursor-up sequences
	if strings.Contains(output, fmt.Sprintf("\033[%dA", BannerLines)) {
		t.Error("non-TTY output should not contain cursor movement")
	}
	// Non-TTY should not contain ANSI color escapes
	if strings.Contains(output, "\x1b[") {
		t.Error("non-TTY output should not contain ANSI escapes")
	}
}

func TestAnimateBannerAsyncNonTTY(t *testing.T) {
	var buf bytes.Buffer
	w, wait := AnimateBannerAsync(&buf)

	if w != &buf {
		t.Error("non-TTY should return original writer")
	}
	wait() // must not block

	output := buf.String()
	if !strings.Contains(output, "\u2800") {
		t.Error("non-TTY output should contain braille content")
	}
	// Non-TTY should not contain ANSI color escapes
	if strings.Contains(output, "\x1b[") {
		t.Error("non-TTY output should not contain ANSI escapes")
	}
}

func TestTracePaintCoverage(t *testing.T) {
	grid, numLines := bannerGrid()
	order := tracePaint(grid, numLines)

	dense := countDense(grid)
	if len(order) != dense {
		t.Errorf("tracePaint returned %d cells, expected %d non-blank cells", len(order), dense)
	}
}

func TestTracePopCoverage(t *testing.T) {
	grid, numLines := bannerGrid()
	order := tracePop(grid, numLines)

	dense := countDense(grid)
	if len(order) != dense {
		t.Errorf("tracePop returned %d cells, expected %d non-blank cells", len(order), dense)
	}
}

func TestTracePopReversesZoneOrder(t *testing.T) {
	grid, numLines := bannerGrid()
	paintOrder := tracePaint(grid, numLines)
	popOrder := tracePop(grid, numLines)

	if len(paintOrder) != len(popOrder) {
		t.Fatal("paint and pop should have the same number of cells")
	}
	if len(paintOrder) == 0 {
		t.Fatal("expected non-empty trace")
	}

	// First popsicle cell in paint should be from top rows,
	// first popsicle cell in pop should be from bottom rows.
	if paintOrder[0].row >= popOrder[0].row {
		t.Error("pop's first cell should be from a lower row than paint's first cell")
	}
}

func TestVisualLinesNoWrap(t *testing.T) {
	rows, pw := visualLines("hello\n", 80, 0)
	if rows != 1 || pw != 0 {
		t.Errorf("got rows=%d pw=%d, want rows=1 pw=0", rows, pw)
	}
}

func TestVisualLinesMultipleNewlines(t *testing.T) {
	rows, pw := visualLines("hello\nworld\n", 80, 0)
	if rows != 2 || pw != 0 {
		t.Errorf("got rows=%d pw=%d, want rows=2 pw=0", rows, pw)
	}
}

func TestVisualLinesSoftWrap(t *testing.T) {
	rows, pw := visualLines(strings.Repeat("x", 25)+"\n", 20, 0)
	if rows != 2 || pw != 0 {
		t.Errorf("got rows=%d pw=%d, want rows=2 pw=0", rows, pw)
	}
}

func TestVisualLinesPartialLine(t *testing.T) {
	rows, pw := visualLines("hello", 80, 0)
	if rows != 0 || pw != 5 {
		t.Errorf("got rows=%d pw=%d, want rows=0 pw=5", rows, pw)
	}
}

func TestVisualLinesZeroCols(t *testing.T) {
	rows, pw := visualLines("hello\nworld\n", 0, 0)
	if rows != 2 || pw != 0 {
		t.Errorf("got rows=%d pw=%d, want rows=2 pw=0", rows, pw)
	}
}

func TestAnimWriterCountsWraps(t *testing.T) {
	var buf bytes.Buffer
	aw := &AnimWriter{
		w:    &buf,
		done: make(chan struct{}),
		cols: 20,
	}

	fmt.Fprint(aw, strings.Repeat("x", 25)+"\n")
	if aw.linesBelow != 2 {
		t.Errorf("expected 2 lines, got %d", aw.linesBelow)
	}

	fmt.Fprint(aw, "hi\n")
	if aw.linesBelow != 3 {
		t.Errorf("expected 3 lines, got %d", aw.linesBelow)
	}
}

func TestAnimWriterTracksPendingWidth(t *testing.T) {
	var buf bytes.Buffer
	aw := &AnimWriter{
		w:    &buf,
		done: make(chan struct{}),
		cols: 80,
	}

	// Partial line: no newline, cursor stays on same line
	fmt.Fprint(aw, "hello")
	if aw.linesBelow != 0 {
		t.Errorf("partial line: linesBelow=%d, want 0", aw.linesBelow)
	}
	if aw.pendingW != 5 {
		t.Errorf("partial line: pendingW=%d, want 5", aw.pendingW)
	}

	// Complete the line
	fmt.Fprint(aw, " world\n")
	if aw.linesBelow != 1 {
		t.Errorf("after newline: linesBelow=%d, want 1", aw.linesBelow)
	}
	if aw.pendingW != 0 {
		t.Errorf("after newline: pendingW=%d, want 0", aw.pendingW)
	}
}

func TestTextZonePopScrollsLeftToRight(t *testing.T) {
	grid, numLines := bannerGrid()
	order := tracePop(grid, numLines)

	var textCells []paintCell
	for _, c := range order {
		if popsicleZone(c.col) == 5 {
			textCells = append(textCells, c)
		}
	}
	if len(textCells) < 2 {
		t.Fatal("expected text zone cells in trace")
	}

	// Pop text cells should be ordered by column (left-to-right sweep)
	for i := 1; i < len(textCells); i++ {
		if textCells[i].col < textCells[i-1].col {
			t.Errorf("text cell %d (col %d) before cell %d (col %d) — expected left-to-right",
				i-1, textCells[i-1].col, i, textCells[i].col)
			break
		}
	}
}

func TestTextZonePaintRadialUnfurl(t *testing.T) {
	grid, numLines := bannerGrid()
	order := tracePaint(grid, numLines)

	var textCells []paintCell
	for _, c := range order {
		if popsicleZone(c.col) == 5 {
			textCells = append(textCells, c)
		}
	}
	if len(textCells) < 2 {
		t.Fatal("expected text zone cells in trace")
	}

	// First revealed text cell should be from the top-left corner (smallest
	// radial distance) and the last from the bottom-right.
	first, last := textCells[0], textCells[len(textCells)-1]
	if first.col > last.col {
		t.Errorf("first text cell col %d > last col %d — expected left-to-right progression",
			first.col, last.col)
	}
	if first.row > last.row {
		t.Errorf("first text cell row %d > last row %d — expected top-to-bottom progression",
			first.row, last.row)
	}

	// The radial ordering should interleave rows: a cell at (row=0, col=X)
	// can appear after (row=1, col=Y) when Y is much smaller than X.
	sawRowDecrease := false
	for i := 1; i < len(textCells); i++ {
		if textCells[i].row < textCells[i-1].row {
			sawRowDecrease = true
			break
		}
	}
	if !sawRowDecrease {
		t.Error("paint text order should interleave rows (radial), but rows only increase")
	}
}

func TestRevealsTextAndPopsiclesSimultaneously(t *testing.T) {
	grid, numLines := bannerGrid()

	for name, traceFn := range map[string]func([][]rune, int) []paintCell{
		"paint": tracePaint,
		"pop":   tracePop,
	} {
		order := traceFn(grid, numLines)

		// In the first quarter of the trace, both popsicle and text cells
		// should appear — proving they reveal simultaneously.
		quarter := len(order) / 4
		sawPopsicle := false
		sawText := false
		for _, c := range order[:quarter] {
			z := popsicleZone(c.col)
			if z < 5 {
				sawPopsicle = true
			} else {
				sawText = true
			}
			if sawPopsicle && sawText {
				break
			}
		}
		if !sawPopsicle {
			t.Errorf("%s: first quarter should contain popsicle cells", name)
		}
		if !sawText {
			t.Errorf("%s: first quarter should contain text cells (simultaneous reveal)", name)
		}
	}
}

func TestBannerVisualRowsWide(t *testing.T) {
	// Terminal wider than banner: no wrapping
	if got := bannerVisualRows(120); got != BannerLines {
		t.Errorf("bannerVisualRows(120) = %d, want %d", got, BannerLines)
	}
	if got := bannerVisualRows(80); got != BannerLines {
		t.Errorf("bannerVisualRows(80) = %d, want %d", got, BannerLines)
	}
}

func TestBannerVisualRowsNarrow(t *testing.T) {
	// Terminal at 40 cols: each 80-wide line wraps to 2 rows
	if got := bannerVisualRows(40); got != BannerLines*2 {
		t.Errorf("bannerVisualRows(40) = %d, want %d", got, BannerLines*2)
	}
	// Terminal at 27 cols: each 80-wide line wraps to ceil(80/27)=3 rows
	if got := bannerVisualRows(27); got != BannerLines*3 {
		t.Errorf("bannerVisualRows(27) = %d, want %d", got, BannerLines*3)
	}
}

func TestBannerVisualRowsZero(t *testing.T) {
	// Unknown terminal width: fall back to line count
	if got := bannerVisualRows(0); got != BannerLines {
		t.Errorf("bannerVisualRows(0) = %d, want %d", got, BannerLines)
	}
}

func TestCompactVisualRowsWide(t *testing.T) {
	// Terminal wider than compact banner: no wrapping
	if got := compactVisualRows(80); got != BannerLines {
		t.Errorf("compactVisualRows(80) = %d, want %d", got, BannerLines)
	}
	if got := compactVisualRows(41); got != BannerLines {
		t.Errorf("compactVisualRows(41) = %d, want %d", got, BannerLines)
	}
}

func TestCompactVisualRowsNarrow(t *testing.T) {
	// At 30 cols: most lines (26 wide) don't wrap, but the 2 text lines do.
	// 10 lines at 26 cols = 10 rows, 1 line at 34 = 2 rows, 1 line at 41 = 2 rows
	got := compactVisualRows(30)
	want := 10 + 2 + 2 // 14
	if got != want {
		t.Errorf("compactVisualRows(30) = %d, want %d", got, want)
	}
}

func TestCompactVisualRowsZero(t *testing.T) {
	if got := compactVisualRows(0); got != BannerLines {
		t.Errorf("compactVisualRows(0) = %d, want %d", got, BannerLines)
	}
}

func TestTextZonePopStrictColumnOrder(t *testing.T) {
	// Verify that pop text zone cells are strictly sorted by column,
	// meaning columns never decrease. Within the same column, row order is
	// preserved (top to bottom).
	grid, numLines := bannerGrid()
	order := tracePop(grid, numLines)

	var textCells []paintCell
	for _, c := range order {
		if popsicleZone(c.col) == 5 {
			textCells = append(textCells, c)
		}
	}
	if len(textCells) < 2 {
		t.Fatal("expected text zone cells in trace")
	}

	for i := 1; i < len(textCells); i++ {
		prev, cur := textCells[i-1], textCells[i]
		if cur.col < prev.col {
			t.Errorf("text cell at (%d,%d) before (%d,%d) — columns must not decrease",
				prev.row, prev.col, cur.row, cur.col)
			break
		}
		if cur.col == prev.col && cur.row < prev.row {
			t.Errorf("same column %d: row %d before %d — rows must not decrease within a column",
				cur.col, prev.row, cur.row)
			break
		}
	}
}

func countDense(grid [][]rune) int {
	n := 0
	for _, row := range grid {
		for _, ch := range row {
			if isDense(ch) {
				n++
			}
		}
	}
	return n
}
