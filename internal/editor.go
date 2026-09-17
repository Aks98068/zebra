package shell

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/term"

	"zebra/internal/ui"
)

var ErrInterrupted = errors.New("input interrupted")
var ErrEOF = errors.New("end of input")

type Editor struct {
	history []string
	index   int

	candidates []string

	prompt string
	buffer []rune
	cursor int
}

func NewEditor(candidates []string) *Editor {
	return &Editor{
		history:    make([]string, 0),
		index:      -1,
		candidates: candidates,
	}
}

func (e *Editor) ReadLine(prompt string) (string, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return e.readFallback(prompt)
	}

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return "", fmt.Errorf("unable to enable terminal raw mode: %w", err)
	}

	defer term.Restore(int(os.Stdin.Fd()), oldState)

	e.prompt = prompt
	e.buffer = nil
	e.cursor = 0
	e.index = -1

	fmt.Print(prompt)

	for {
		key, err := readKey()
		if err != nil {
			return "", err
		}

		switch key.kind {

		case keyCharacter:
			e.insert(key.rune)
			e.render()

		case keyBackspace:
			e.backspace()
			e.render()

		case keyDelete:
			e.delete()
			e.render()

		case keyLeft:
			if e.cursor > 0 {
				e.cursor--
				e.renderCursor()
			}

		case keyRight:
			if e.cursor < len(e.buffer) {
				e.cursor++
				e.renderCursor()
			}

		case keyHome:
			e.cursor = 0
			e.renderCursor()

		case keyEnd:
			e.cursor = len(e.buffer)
			e.renderCursor()

		case keyUp:
			e.historyUp()
			e.render()

		case keyDown:
			e.historyDown()
			e.render()

		case keyTab:
			e.autocomplete()

		case keyEnter:
			line := string(e.buffer)

			fmt.Print("\r\n")

			line = strings.TrimSpace(line)

			if line != "" {
				e.addHistory(line)
			}

			return line, nil

		case keyCtrlC:
			fmt.Print("^C\r\n")
			return "", ErrInterrupted

		case keyCtrlL:
			e.clearScreen()
			e.render()

		case keyCtrlD:
			if len(e.buffer) == 0 {
				fmt.Print("\r\n")
				return "", ErrEOF
			}

			e.delete()
			e.render()
		}
	}
}

func (e *Editor) readFallback(prompt string) (string, error) {
	fmt.Print(prompt)

	var input string

	_, err := fmt.Scanln(&input)

	if err != nil {
		if errors.Is(err, io.EOF) {
			return "", ErrEOF
		}

		return "", err
	}

	return input, nil
}

// ============================================================
// Editing
// ============================================================

func (e *Editor) insert(r rune) {
	if !unicode.IsControl(r) {
		e.buffer = append(
			e.buffer[:e.cursor],
			append([]rune{r}, e.buffer[e.cursor:]...)...,
		)

		e.cursor++
	}
}

func (e *Editor) backspace() {
	if e.cursor <= 0 {
		return
	}

	e.buffer = append(
		e.buffer[:e.cursor-1],
		e.buffer[e.cursor:]...,
	)

	e.cursor--
}

func (e *Editor) delete() {
	if e.cursor >= len(e.buffer) {
		return
	}

	e.buffer = append(
		e.buffer[:e.cursor],
		e.buffer[e.cursor+1:]...,
	)
}

// ============================================================
// History
// ============================================================

func (e *Editor) addHistory(command string) {
	if command == "" {
		return
	}

	if len(e.history) > 0 &&
		e.history[len(e.history)-1] == command {
		return
	}

	e.history = append(e.history, command)

	if len(e.history) > 1000 {
		e.history = e.history[len(e.history)-1000:]
	}
}

func (e *Editor) historyUp() {
	if len(e.history) == 0 {
		return
	}

	if e.index == -1 {
		e.index = len(e.history) - 1
	} else if e.index > 0 {
		e.index--
	}

	e.buffer = []rune(e.history[e.index])
	e.cursor = len(e.buffer)
}

func (e *Editor) historyDown() {
	if len(e.history) == 0 {
		return
	}

	if e.index == -1 {
		return
	}

	if e.index < len(e.history)-1 {
		e.index++

		e.buffer = []rune(e.history[e.index])
		e.cursor = len(e.buffer)

		return
	}

	e.index = -1
	e.buffer = nil
	e.cursor = 0
}

// ============================================================
// Autocomplete
// ============================================================

func (e *Editor) autocomplete() {
	line := string(e.buffer)

	if strings.TrimSpace(line) == "" {
		return
	}

	start := e.cursor

	for start > 0 && !unicode.IsSpace(e.buffer[start-1]) {
		start--
	}

	token := string(e.buffer[start:e.cursor])

	if token == "" {
		return
	}

	matches := e.findMatches(token)

	if len(matches) == 0 {
		return
	}

	if len(matches) == 1 {
		match := matches[0]

		e.buffer = append(
			e.buffer[:start],
			append(
				[]rune(match),
				e.buffer[e.cursor:]...,
			)...,
		)

		e.cursor = start + len([]rune(match))

		if !strings.HasSuffix(match, string(os.PathSeparator)) {
			if fileExists(match) {
				e.buffer = append(
					e.buffer[:e.cursor],
					append(
						[]rune{' '},
						e.buffer[e.cursor:]...,
					)...,
				)

				e.cursor++
			}
		}

		e.render()
		return
	}

	fmt.Print("\r\n")

	for _, match := range matches {
		fmt.Printf(
			"%s%s%s  ",
			ui.BrightCyan,
			match,
			ui.Reset,
		)
	}

	fmt.Print("\r\n")

	e.render()
}

func (e *Editor) findMatches(token string) []string {
	var matches []string

	// Command completion for the first token.
	if !strings.Contains(token, string(os.PathSeparator)) &&
		!strings.Contains(token, "/") &&
		!strings.Contains(token, `\`) {

		for _, candidate := range e.candidates {
			if strings.HasPrefix(
				strings.ToLower(candidate),
				strings.ToLower(token),
			) {
				matches = append(matches, candidate)
			}
		}
	}

	// Filesystem completion.
	if len(matches) == 0 {
		matches = findPathMatches(token)
	}

	return matches
}

func findPathMatches(input string) []string {
	input = strings.TrimSpace(input)

	if input == "" {
		return nil
	}

	expanded := input

	if input == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			expanded = home
		}
	} else if strings.HasPrefix(input, "~/") ||
		strings.HasPrefix(input, `~\`) {

		home, err := os.UserHomeDir()
		if err == nil {
			expanded = filepath.Join(home, input[2:])
		}
	}

	dir := filepath.Dir(expanded)
	prefix := filepath.Base(expanded)

	if prefix == "." || prefix == string(filepath.Separator) {
		dir = expanded
		prefix = ""
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var matches []string

	for _, entry := range entries {
		name := entry.Name()

		if !strings.HasPrefix(
			strings.ToLower(name),
			strings.ToLower(prefix),
		) {
			continue
		}

		fullPath := filepath.Join(dir, name)

		if entry.IsDir() {
			fullPath += string(os.PathSeparator)
		}

		matches = append(matches, fullPath)
	}

	return matches
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ============================================================
// Rendering
// ============================================================

func (e *Editor) render() {
	fmt.Print("\r")

	fmt.Print("\033[2K")

	fmt.Print(e.prompt)

	fmt.Print(colorizeInput(string(e.buffer)))

	fmt.Print("\033[0K")

	e.renderCursor()
}

func (e *Editor) renderCursor() {
	visibleBeforeCursor := colorizeInput(
		string(e.buffer[:e.cursor]),
	)

	visibleWidth := visibleLength(visibleBeforeCursor)

	currentWidth := visibleLength(
		colorizeInput(string(e.buffer)),
	)

	moveLeft := currentWidth - visibleWidth

	if moveLeft > 0 {
		fmt.Printf("\033[%dD", moveLeft)
	}
}

func (e *Editor) clearScreen() {
	fmt.Print("\033[2J")
	fmt.Print("\033[H")
}

// ============================================================
// Input Coloring
// ============================================================

func colorizeInput(input string) string {
	if input == "" {
		return ""
	}

	tokens := strings.Fields(input)

	if len(tokens) == 0 {
		return input
	}

	var result strings.Builder

	position := 0

	for _, token := range tokens {

		index := strings.Index(
			input[position:],
			token,
		)

		if index < 0 {
			continue
		}

		index += position

		result.WriteString(input[position:index])

		switch {
		case position == 0:
			result.WriteString(
				ui.BrightCyan +
					ui.Bold +
					token +
					ui.Reset,
			)

		case strings.HasPrefix(token, "-"):
			result.WriteString(
				ui.BrightYellow +
					token +
					ui.Reset,
			)

		case looksLikePath(token):
			result.WriteString(
				ui.BrightGreen +
					token +
					ui.Reset,
			)

		case looksLikeNumber(token):
			result.WriteString(
				ui.BrightMagenta +
					token +
					ui.Reset,
			)

		default:
			result.WriteString(
				ui.BrightWhite +
					token +
					ui.Reset,
			)
		}

		position = index + len(token)
	}

	if position < len(input) {
		result.WriteString(input[position:])
	}

	return result.String()
}

func looksLikePath(value string) bool {
	return strings.Contains(value, `/`) ||
		strings.Contains(value, `\`) ||
		strings.HasPrefix(value, ".") ||
		strings.HasPrefix(value, "~") ||
		strings.Contains(value, ":")
}

func looksLikeNumber(value string) bool {
	if value == "" {
		return false
	}

	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}

func visibleLength(value string) int {
	length := 0
	inEscape := false

	for _, r := range value {
		if r == '\033' {
			inEscape = true
			continue
		}

		if inEscape {
			if r == 'm' {
				inEscape = false
			}

			continue
		}

		length++
	}

	return length
}

// ============================================================
// Keyboard Input
// ============================================================

type keyKind int

const (
	keyCharacter keyKind = iota
	keyEnter
	keyBackspace
	keyDelete
	keyLeft
	keyRight
	keyUp
	keyDown
	keyHome
	keyEnd
	keyTab
	keyCtrlC
	keyCtrlL
	keyCtrlD
)

type key struct {
	kind keyKind
	rune rune
}

func readKey() (key, error) {
	buffer := make([]byte, 1)

	_, err := os.Stdin.Read(buffer)

	if err != nil {
		return key{}, err
	}

	b := buffer[0]

	switch b {

	case 3:
		return key{kind: keyCtrlC}, nil

	case 4:
		return key{kind: keyCtrlD}, nil

	case 9:
		return key{kind: keyTab}, nil

	case 10, 13:
		return key{kind: keyEnter}, nil

	case 12:
		return key{kind: keyCtrlL}, nil

	case 127, 8:
		return key{kind: keyBackspace}, nil

	case 27:
		return readEscapeSequence()
	}

	if b >= 32 && b <= 126 {
		return key{
			kind: keyCharacter,
			rune: rune(b),
		}, nil
	}

	return key{}, nil
}

func readEscapeSequence() (key, error) {
	buffer := make([]byte, 2)

	_, err := io.ReadFull(
		os.Stdin,
		buffer,
	)

	if err != nil {
		return key{}, err
	}

	if buffer[0] != '[' {
		return key{}, nil
	}

	switch buffer[1] {

	case 'A':
		return key{kind: keyUp}, nil

	case 'B':
		return key{kind: keyDown}, nil

	case 'C':
		return key{kind: keyRight}, nil

	case 'D':
		return key{kind: keyLeft}, nil

	case 'H':
		return key{kind: keyHome}, nil

	case 'F':
		return key{kind: keyEnd}, nil

	case '3':
		next := make([]byte, 1)

		if _, err := os.Stdin.Read(next); err != nil {
			return key{}, err
		}

		if next[0] == '~' {
			return key{kind: keyDelete}, nil
		}
	}

	return key{}, nil
}