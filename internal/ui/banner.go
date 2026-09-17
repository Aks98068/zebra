package ui

import (
	"fmt"
	"strings"
)

const (
	Version = "1.0.0"
	Author  = "Abhishekh Kumar Sah"
	Github  = "github.com/Aks98068"
)

// ============================================================
// ZEBRA BANNER
// ============================================================

func PrintBanner() {
	fmt.Println()

	printLogo()

	fmt.Println()

	printInfoBox()

	fmt.Println()

	fmt.Println(
		BrightGreen + Bold + "  ✔ " + Reset +
			BrightWhite + "Welcome to " + Reset +
			BrightCyan + Bold + "Zebra" + Reset +
			BrightWhite + " — your cross-platform command shell." + Reset,
	)

	fmt.Println()

	fmt.Println(
		BrightYellow + Bold + "  ➜ " + Reset +
			BrightWhite + "Type " + Reset +
			BrightCyan + Bold + "'help'" + Reset +
			BrightWhite + " to see available commands." + Reset,
	)

	fmt.Println()
}

// ============================================================
// LOGO
// ============================================================

func printLogo() {
	logo := []string{
		"███████╗███████╗██████╗ ██████╗  █████╗ ",
		"╚══███╔╝██╔════╝██╔══██╗██╔══██╗██╔══██╗",
		"  ███╔╝ █████╗  ██████╔╝██████╔╝███████║",
		" ███╔╝  ██╔══╝  ██╔══██╗██╔══██╗██╔══██║",
		"███████╗███████╗██████╔╝██████╔╝██║  ██║",
		"╚══════╝╚══════╝╚═════╝ ╚═════╝ ╚═╝  ╚═╝",
	}

	colors := []string{
		BrightCyan,
		BrightBlue,
		BrightMagenta,
		BrightCyan,
		BrightBlue,
		BrightMagenta,
	}

	for i, line := range logo {
		fmt.Println(
			"  " +
				colors[i%len(colors)] +
				Bold +
				line +
				Reset,
		)
	}

	fmt.Println()

	fmt.Println(
		"  " +
			BrightBlack +
			"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" +
			Reset,
	)

	fmt.Println(
		"  " +
			BrightCyan +
			Bold +
			"Z E B R A" +
			Reset +
			"  " +
			BrightWhite +
			"Cross-Platform Terminal Environment" +
			Reset,
	)

	fmt.Println(
		"  " +
			BrightBlack +
			"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" +
			Reset,
	)
}

// ============================================================
// INFORMATION BOX
// ============================================================

func printInfoBox() {
	const width = 64

	line := strings.Repeat("═", width)

	fmt.Println(
		BrightCyan +
			"╔" +
			line +
			"╗" +
			Reset,
	)

	printBoxLine(
		width,
		BrightYellow+Bold+"ZEBRA CLI TOOL"+Reset+
			"   "+
			BrightWhite+"v"+Version+Reset,
	)

	printBoxLine(
		width,
		BrightGreen+"AUTHOR"+Reset+
			"    : "+
			BrightWhite+Author+Reset,
	)

	printBoxLine(
		width,
		BrightBlue+"GITHUB"+Reset+
			"    : "+
			BrightCyan+Github+Reset,
	)

	printBoxLine(
		width,
		BrightMagenta+"PLATFORM"+Reset+
			"  : "+
			BrightWhite+"Windows / Linux / macOS"+Reset,
	)

	printBoxLine(
		width,
		BrightRed+"STATUS"+Reset+
			"    : "+
			BrightGreen+"● Ready"+Reset,
	)

	fmt.Println(
		BrightCyan +
			"╚" +
			line +
			"╝" +
			Reset,
	)
}

// ============================================================
// BOX LINE
// ============================================================

func printBoxLine(width int, text string) {
	visibleText := stripANSI(text)

	padding := width - len(visibleText) - 2

	if padding < 0 {
		padding = 0
	}

	fmt.Printf(
		BrightCyan+"║"+Reset+
			" %s%s "+
			BrightCyan+"║"+Reset+"\n",
		text,
		strings.Repeat(" ", padding),
	)
}

// ============================================================
// ANSI STRIPPER
// ============================================================

func stripANSI(text string) string {
	sequences := []string{
		Reset,
		Bold,
		Dim,
		Underline,

		Black,
		Red,
		Green,
		Yellow,
		Blue,
		Magenta,
		Cyan,
		White,

		BrightBlack,
		BrightRed,
		BrightGreen,
		BrightYellow,
		BrightBlue,
		BrightMagenta,
		BrightCyan,
		BrightWhite,
	}

	for _, sequence := range sequences {
		text = strings.ReplaceAll(text, sequence, "")
	}

	return text
}

