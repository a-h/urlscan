package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/a-h/urlscan/urlscanner"
	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
	"github.com/pkg/browser"
)

func main() {
	urls, err := urlscanner.Scan(os.Stdin)
	if err != nil {
		fmt.Println("Error reading:", err)
		os.Exit(1)
	}

	// Create a screen.
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	s, err := tcell.NewScreen()
	if err != nil {
		fmt.Println("Error creating screen:", err)
		os.Exit(1)
	}
	if err = s.Init(); err != nil {
		fmt.Println("Error initializing screen:", err)
		os.Exit(1)
	}
	defer s.Fini()
	s.SetStyle(defaultStyle)
	action := NewOptions(s, append(urls)...).Focus()
	if action == "Exit" {
		return
	}
	browser.OpenURL(action)
}

// Below are functions taken from the min browser (github.com/a-h/min).

// flow breaks up text to its maximum width.
func flow(s string, maxWidth int) []string {
	var ss []string
	flowProcessor(s, maxWidth, func(line string) {
		ss = append(ss, line)
	})
	return ss
}

func flowProcessor(s string, maxWidth int, out func(string)) {
	var buf strings.Builder
	var col int
	var lastSpace int
	for _, r := range s {
		if r == '\r' {
			continue
		}
		if r == '\n' {
			out(buf.String())
			buf.Reset()
			col = 0
			lastSpace = 0
			continue
		}
		buf.WriteRune(r)
		if unicode.IsSpace(r) {
			lastSpace = col
		}
		if col == maxWidth {
			// If the word is greater than the width, then break the word down.
			end := lastSpace
			if end == 0 {
				end = col
			}
			out(strings.TrimSpace(buf.String()[:end]))
			prefix := strings.TrimSpace(buf.String()[end:])
			buf.Reset()
			lastSpace = 0
			buf.WriteString(prefix)
			col = len(prefix)
			continue
		}
		col++
	}
	out(buf.String())
}

var defaultStyle = tcell.StyleDefault.
	Foreground(tcell.ColorWhite).
	Background(tcell.ColorBlack)

func NewText(s tcell.Screen, text string) *Text {
	return &Text{
		Screen: s,
		X:      0,
		Y:      0,
		Style:  defaultStyle,
		Text:   text,
	}
}

type Text struct {
	Screen   tcell.Screen
	X        int
	Y        int
	MaxWidth int
	Style    tcell.Style
	Text     string
}

func (t *Text) WithOffset(x, y int) *Text {
	t.X = x
	t.Y = y
	return t
}

func (t *Text) WithMaxWidth(x int) *Text {
	t.MaxWidth = x
	return t
}

func (t *Text) WithStyle(st tcell.Style) *Text {
	t.Style = st
	return t
}

func (t *Text) Draw() (x, y int) {
	maxX, _ := t.Screen.Size()
	maxWidth := maxX - t.X
	if t.MaxWidth > 0 && maxWidth > t.MaxWidth {
		maxWidth = t.MaxWidth
	}
	lines := flow(t.Text, maxWidth)
	var requiredMaxWidth int
	for lineIndex := 0; lineIndex < len(lines); lineIndex++ {
		y = t.Y + lineIndex
		x = t.X
		for _, c := range lines[lineIndex] {
			var comb []rune
			w := runewidth.RuneWidth(c)
			if w == 0 {
				comb = []rune{c}
				c = ' '
				w = 1
			}
			t.Screen.SetContent(x, y, c, comb, t.Style)
			x += w
			if x > requiredMaxWidth {
				requiredMaxWidth = x
			}
		}
	}
	return requiredMaxWidth, y
}

func NewOptions(s tcell.Screen, opts ...string) *Options {
	cancelIndex := -1
	for i, o := range opts {
		if o == "Cancel" || o == "Exit" {
			cancelIndex = i
			break
		}
	}
	return &Options{
		Screen:      s,
		Style:       defaultStyle,
		Options:     opts,
		CancelIndex: cancelIndex,
	}
}

type Choice struct {
	Index int
	Value string
}

func ChoiceOptionIndex(options ...string) (op []string) {
	op = make([]string, len(options))
	for i := range options {
		op[i] = strconv.FormatInt(int64(i), 10)
	}
	return op
}

func choiceIndex(options []string, s string) (index int, matchesPrefix bool) {
	index = -1
	for i, c := range options {
		if c == s {
			index = i
			continue
		}
		if strings.HasPrefix(c, s) {
			matchesPrefix = true
		}
	}
	return
}

func NewChoice(options ...string) (runes chan rune, selection chan Choice, closer func()) {
	var buffer string
	runes = make(chan rune)
	selection = make(chan Choice)

	var ctx context.Context
	ctx, closer = context.WithCancel(context.Background())
	go func(ctx context.Context) {
		defer close(runes)
		defer close(selection)
		index := -1
		var matchesPrefix bool
		for {
			select {
			case <-time.After(time.Millisecond * 200):
				if index > -1 {
					selection <- Choice{Index: index, Value: options[index]}
					index = -1
				}
				buffer = ""
			case r := <-runes:
				buffer += string(r)
				index, matchesPrefix = choiceIndex(options, buffer)
				if index < 0 {
					buffer = ""
					continue
				}
				if matchesPrefix {
					continue // Wait to see if any more is typed in.
				}
				selection <- Choice{Index: index, Value: options[index]}
				index = -1
				buffer = ""
			case <-ctx.Done():
				return
			}
		}
	}(ctx)
	return
}

type Options struct {
	Screen      tcell.Screen
	X           int
	Y           int
	Style       tcell.Style
	Options     []string
	ActiveIndex int
	CancelIndex int
}

func (o *Options) Draw() {
	o.Screen.Clear()
	for i, oo := range o.Options {
		style := defaultStyle
		var prefix = " "
		if i == o.ActiveIndex {
			style = defaultStyle.Background(tcell.ColorGray)
			prefix = ">"
		}
		NewText(o.Screen, fmt.Sprintf("%s [%d] %s", prefix, i, oo)).WithOffset(1, i).WithStyle(style).Draw()
	}
}

func (o *Options) Up() {
	if o.ActiveIndex == 0 {
		o.ActiveIndex = len(o.Options) - 1
		return
	}
	o.ActiveIndex--
}

func (o *Options) Down() {
	if o.ActiveIndex == len(o.Options)-1 {
		o.ActiveIndex = 0
		return
	}
	o.ActiveIndex++
}

func (o *Options) Focus() string {
	runes, selection, closer := NewChoice(ChoiceOptionIndex(o.Options...)...)
	defer closer()
	go func() {
		for choice := range selection {
			o.ActiveIndex = choice.Index
			o.Screen.PostEvent(nil)
		}
	}()
	o.Draw()
	o.Screen.Show()
	for {
		switch ev := o.Screen.PollEvent().(type) {
		case *tcell.EventResize:
			o.Screen.Sync()
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyBacktab:
				o.Up()
			case tcell.KeyTab:
				o.Down()
			case tcell.KeyUp:
				o.Up()
			case tcell.KeyDown:
				o.Down()
			case tcell.KeyRune:
				if ev.Rune() == 'q' {
					return "quit"
				}
				runes <- ev.Rune()
			case tcell.KeyEscape:
				return "quit"
			case tcell.KeyEnter:
				return o.Options[o.ActiveIndex]
			}
		}
		o.Draw()
		o.Screen.Show()
	}
}
