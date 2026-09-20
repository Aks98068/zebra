package ui

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	Version = "1.0.0"
	Author  = "Abhishekh Kumar Sah"
	Github  = "github.com/Aks98068"
)

// ============================================================
// ANSI / TERMINAL CONTROLS
// ============================================================

const (
	ClearScreen = "\033[2J"
	CursorHome  = "\033[H"
	HideCursor  = "\033[?25l"
	ShowCursor  = "\033[?25h"
)

func moveCursor(row, col int) {
	fmt.Printf("\033[%d;%dH", row, col)
}

func clearScreen() {
	fmt.Print(ClearScreen + CursorHome)
}

// ============================================================
// MAIN BANNER
// ============================================================

func PrintBanner() {
	fmt.Print(HideCursor)
	defer fmt.Print(ShowCursor)

	clearScreen()

	// --------------------------------------------------------
	// STARTUP ANIMATION
	// --------------------------------------------------------

	animateTopStripe()
	animateZebraWithSignals()
	pulseEyes()
	animateLogoText()
	animateSeparator()
	printInfoBox()
	printWelcome()
	animateBottomStripe()

	moveCursor(34, 1)

	fmt.Print(Reset)
}

// ============================================================
// TOP STRIPE
// ============================================================

func animateTopStripe() {
	stripe := `  ▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░`

	colors := []string{
		BrightWhite,
		BrightCyan,
		BrightMagenta,
		BrightYellow,
		BrightGreen,
		BrightCyan,
	}

	for i := 0; i < len(colors); i++ {
		moveCursor(1, 1)

		fmt.Print(
			colors[i] +
				Bold +
				stripe +
				Reset,
		)

		time.Sleep(70 * time.Millisecond)
	}
}

// ============================================================
// ZEBRA + RECON SCANNER
// ============================================================

type scanLine struct {
	row    int
	face   string
	target string
	color  string
}

// ============================================================
// ZEBRA FACE
// ============================================================

func animateZebraWithSignals() {

	lines := []scanLine{

		{
			row:    3,
			face:   BrightWhite + `          _____` + Reset,
			target: `[ DNS Recon    ]`,
			color:  BrightGreen,
		},

		{
			row: 4,
			face: BrightWhite +
				`         /` +
				BrightBlack + `▓░▓░▓` +
				BrightWhite + `\` +
				Reset,
			target: `[ WHOIS        ]`,
			color:  BrightGreen,
		},

		{
			row: 5,
			face: BrightWhite +
				`        /` +
				BrightBlack + `░` +
				BrightGreen + Bold + `[◈]` +
				Reset +
				BrightBlack + `▓░▓` +
				BrightWhite + `\` +
				Reset,
			target: `[ HTTP/Web     ]`,
			color:  BrightCyan,
		},

		{
			row: 6,
			face: BrightWhite +
				`        |` +
				BrightBlack + `▓░▓░▓░▓` +
				BrightWhite + `|` +
				Reset,
			target: `[ SSL/TLS      ]`,
			color:  BrightCyan,
		},

		{
			row: 7,
			face: BrightWhite +
				`        |` +
				BrightBlack + `░▓░▓░▓░` +
				BrightWhite + `|` +
				Reset,
			target: `[ Port Scan    ]`,
			color:  BrightYellow,
		},

		{
			row: 8,
			face: BrightWhite +
				`        |` +
				BrightBlack + `▓░` +
				BrightWhite + `_____` +
				BrightBlack + `░▓` +
				BrightWhite + `|` +
				Reset,
			target: `[ Banner Grab  ]`,
			color:  BrightYellow,
		},

		{
			row: 9,
			face: BrightWhite +
				`         \` +
				BrightBlack + `░▓░▓░▓` +
				BrightWhite + `/` +
				Reset,
			target: `[ Subdomains   ]`,
			color:  BrightMagenta,
		},

		{
			row: 10,
			face: BrightWhite +
				`          \_____/` +
				Reset,
			target: `[ Email Recon  ]`,
			color:  BrightMagenta,
		},

		{
			row: 11,
			face: BrightWhite +
				`           |` +
				BrightBlack + `▓░▓` +
				BrightWhite + `|` +
				Reset,
			target: `[ OSINT        ]`,
			color:  BrightRed,
		},

		{
			row: 12,
			face: BrightWhite +
				`           |` +
				BrightBlack + `░▓░` +
				BrightWhite + `|` +
				Reset,
			target: `[ System Info  ]`,
			color:  BrightRed,
		},

		{
			row: 13,
			face: BrightWhite +
				`          /|` +
				BrightBlack + `▓░▓` +
				BrightWhite + `|\` +
				Reset,
			target: `[ File Recon   ]`,
			color:  BrightBlue,
		},

		{
			row: 14,
			face: BrightWhite +
				`         / |` +
				BrightBlack + `░▓░` +
				BrightWhite + `| \` +
				Reset,
			target: `[ Vuln Scan    ]`,
			color:  BrightBlue,
		},
	}

	// Draw zebra first.
	for _, line := range lines {
		moveCursor(line.row, 1)
		fmt.Print(line.face)
		time.Sleep(35 * time.Millisecond)
	}

	// Animate all scanner lines.
	for _, line := range lines {
		animateScan(line)
	}
}

// ============================================================
// ANIMATED SCANNER
// ============================================================

func animateScan(line scanLine) {

	const (
		startColumn = 22
		scanLength  = 24
	)

	// --------------------------------------------------------
	// Build the target position.
	// --------------------------------------------------------

	targetStart := startColumn + scanLength + 3

	// --------------------------------------------------------
	// Draw scanner moving toward target.
	// --------------------------------------------------------

	for position := 0; position <= scanLength; position++ {

		moveCursor(line.row, startColumn)

		// Signal before scanner.
		signal := strings.Repeat("≋", position)

		// Empty space after scanner.
		remaining := scanLength - position

		space := strings.Repeat(" ", remaining)

		fmt.Print(
			BrightCyan +
				signal +
				space +
				Bold +
				"►" +
				Reset,
		)

		time.Sleep(22 * time.Millisecond)
	}

	// --------------------------------------------------------
	// Draw target.
	// --------------------------------------------------------

	moveCursor(line.row, targetStart)

	fmt.Print(
		line.color +
			Bold +
			line.target +
			Reset,
	)

	time.Sleep(55 * time.Millisecond)
}

// ============================================================
// EYE PULSE
// ============================================================

func pulseEyes() {

	colors := []string{
		BrightGreen,
		BrightYellow,
		BrightRed,
		BrightMagenta,
		BrightCyan,
		BrightGreen,
	}

	for _, color := range colors {

		moveCursor(5, 9)

		fmt.Print(
			BrightWhite +
				`/` +
				BrightBlack +
				`░` +
				color +
				Bold +
				`[◈]` +
				Reset +
				BrightBlack +
				`▓░▓` +
				BrightWhite +
				`\` +
				Reset,
		)

		time.Sleep(110 * time.Millisecond)
	}
}

// ============================================================
// ZEBRA LOGO
// ============================================================

func animateLogoText() {

	logo := []string{

		`███████╗███████╗██████╗ ██████╗  █████╗`,

		`╚══███╔╝██╔════╝██╔══██╗██╔══██╗██╔══██╗`,

		`  ███╔╝ █████╗  ██████╔╝██████╔╝███████║`,

		` ███╔╝  ██╔══╝  ██╔══██╗██╔══██╗███████║`,

		`███████╗███████╗██████╔╝██████╔╝██║  ██║`,

		`╚══════╝╚══════╝╚═════╝ ╚═════╝ ╚═╝  ╚═╝`,
	}

	colors := []string{
		BrightCyan,
		BrightBlue,
		BrightMagenta,
		BrightCyan,
		BrightBlue,
		BrightMagenta,
	}

	startRow := 16

	for i, line := range logo {

		moveCursor(startRow+i, 2)

		fmt.Print(
			colors[i] +
				Bold +
				line +
				Reset,
		)

		time.Sleep(85 * time.Millisecond)
	}
}

// ============================================================
// SEPARATOR
// ============================================================

func animateSeparator() {

	separator := `━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`

	colors := []string{
		BrightCyan,
		BrightMagenta,
		BrightYellow,
		BrightGreen,
		BrightCyan,
	}

	moveCursor(23, 2)

	for _, color := range colors {

		moveCursor(23, 2)

		fmt.Print(
			color +
				Bold +
				separator +
				Reset,
		)

		time.Sleep(80 * time.Millisecond)
	}

	// --------------------------------------------------------
	// Tagline
	// --------------------------------------------------------

	tagline := `" Stripes are Signals. Hunt Without Being Seen. "`

	moveCursor(24, 2)

	tagColors := []string{
		BrightWhite,
		BrightCyan,
		BrightMagenta,
		BrightYellow,
		BrightGreen,
	}

	for i, character := range tagline {

		fmt.Print(
			tagColors[i%len(tagColors)] +
				Bold +
				string(character) +
				Reset,
		)

		time.Sleep(15 * time.Millisecond)
	}

	moveCursor(25, 2)

	fmt.Print(
		BrightCyan +
			Bold +
			separator +
			Reset,
	)
}

// ============================================================
// INFORMATION BOX
// ============================================================

func printInfoBox() {

	const width = 64

	line := strings.Repeat("═", width)

	moveCursor(27, 1)

	fmt.Print(
		BrightCyan +
			"  ╔" +
			line +
			"╗" +
			Reset +
			"\n",
	)

	printBoxLine(
		width,
		BrightYellow+
			Bold+
			"ZEBRA CLI TOOL"+
			Reset+
			"   "+
			BrightWhite+
			"v"+
			Version+
			Reset,
	)

	printBoxLine(
		width,
		BrightGreen+
			"AUTHOR"+
			Reset+
			"    : "+
			BrightWhite+
			Author+
			Reset,
	)

	printBoxLine(
		width,
		BrightBlue+
			"GITHUB"+
			Reset+
			"    : "+
			BrightCyan+
			Github+
			Reset,
	)

	printBoxLine(
		width,
		BrightMagenta+
			"PLATFORM"+
			Reset+
			"  : "+
			BrightWhite+
			"Windows / Linux / macOS"+
			Reset,
	)

	printBoxLine(
		width,
		BrightRed+
			"PURPOSE"+
			Reset+
			"   : "+
			BrightYellow+
			"Cybersecurity CLI Toolkit"+
			Reset,
	)

	printBoxLine(
		width,
		BrightRed+
			"WARNING"+
			Reset+
			"   : "+
			BrightRed+
			Bold+
			"[!] For Educational Purposes Only"+
			Reset,
	)

	printBoxLine(
		width,
		BrightRed+
			"STATUS"+
			Reset+
			"    : "+
			BrightGreen+
			Bold+
			"● Ready"+
			Reset,
	)

	fmt.Print(
		BrightCyan +
			"  ╚" +
			line +
			"╝" +
			Reset +
			"\n",
	)
}

// ============================================================
// WELCOME MESSAGE
// ============================================================

func printWelcome() {

	fmt.Println()

	fmt.Print(
		BrightGreen +
			Bold +
			"  ✔ " +
			Reset +
			BrightWhite +
			"Welcome to " +
			Reset +
			BrightCyan +
			Bold +
			"Zebra" +
			Reset +
			BrightWhite +
			" — Cybersecurity CLI Toolkit." +
			Reset +
			"\n",
	)

	fmt.Println()

	fmt.Print(
		BrightYellow +
			Bold +
			"  ➜ " +
			Reset +
			BrightWhite +
			"Type " +
			Reset +
			BrightCyan +
			Bold +
			"'help'" +
			Reset +
			BrightWhite +
			" to see all available commands." +
			Reset +
			"\n",
	)

	fmt.Println()
}

// ============================================================
// BOTTOM STRIPE
// ============================================================

func animateBottomStripe() {

	stripe := `  ▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░▓░`

	colors := []string{
		BrightCyan,
		BrightMagenta,
		BrightYellow,
		BrightGreen,
		BrightWhite,
		BrightCyan,
	}

	moveCursor(34, 1)

	for i := 0; i < len(colors); i++ {

		moveCursor(34, 1)

		fmt.Print(
			colors[i] +
				Bold +
				stripe +
				Reset,
		)

		time.Sleep(70 * time.Millisecond)
	}

	fmt.Print(Reset)
}

// ============================================================
// BOX LINE
// ============================================================

func printBoxLine(width int, text string) {

	visibleText := visibleWidth(text)

	padding := width - visibleText - 2

	if padding < 0 {
		padding = 0
	}

	fmt.Printf(
		BrightCyan+
			"  ║"+
			Reset+
			" %s%s "+
			BrightCyan+
			"║"+
			Reset+
			"\n",

		text,
		strings.Repeat(" ", padding),
	)
}

// ============================================================
// ANSI WIDTH
// ============================================================

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

func stripANSI(text string) string {
	return ansiRegexp.ReplaceAllString(text, "")
}

// ============================================================
// VISIBLE WIDTH
// ============================================================

func visibleWidth(text string) int {

	clean := stripANSI(text)

	width := 0

	for range clean {
		width++
	}

	return width
}
