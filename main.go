package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/termenv"
	"github.com/nfnt/resize"
)

func main() {
	p := tea.NewProgram(model{urls: os.Args[1:len(os.Args)]})
	p.EnterAltScreen()
	err := p.Start()
	p.ExitAltScreen()
	if err != nil {
		panic(err)
	}
}

const sparkles = "✨"

type model struct {
	selected int
	urls     []string
	image    string
}

func (m model) Init() tea.Cmd {
	return load(m.urls[m.selected])
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "j", "down":
			if m.selected+1 != len(m.urls) {
				m.selected++
			} else {
				m.selected = 0
			}
			return m, load(m.urls[m.selected])
		case "k", "up":
			if m.selected-1 != -1 {
				m.selected--
			} else {
				m.selected = len(m.urls) - 1
			}
			return m, load(m.urls[m.selected])
		}
	case errMsg:
		m.image = msg.Error()
		return m, nil
	case loadMsg:
		url := m.urls[m.selected]
		if msg.resp != nil {
			defer msg.resp.Body.Close()
			img, err := readerToImage(url, msg.resp.Body)
			if err != nil {
				return m, func() tea.Msg { return errMsg(err) }
			}
			m.image = img
			return m, nil
		}
		defer msg.file.Close()
		img, err := readerToImage(url, msg.file)
		if err != nil {
			return m, func() tea.Msg { return errMsg(err) }
		}
		m.image = img
		return m, nil
	}
	return m, nil
}

func (m model) View() string {
	if m.image == "" {
		return fmt.Sprintf("loading %s %s", m.urls[m.selected], sparkles)
	}
	return m.image
}

type loadMsg struct {
	resp *http.Response
	file *os.File
}

type errMsg error

func load(url string) tea.Cmd {
	if strings.HasPrefix(url, "http") {
		return func() tea.Msg {
			resp, err := http.Get(url)
			if err != nil {
				return errMsg(err)
			}
			return loadMsg{resp: resp}
		}
	}
	return func() tea.Msg {
		file, err := os.Open(url)
		if err != nil {
			return errMsg(err)
		}
		return loadMsg{file: file}
	}
}

func readerToImage(url string, r io.Reader) (string, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return "", err
	}

	img = resize.Resize(50, 0, img, resize.Lanczos3)
	b := img.Bounds()
	w := b.Max.X
	h := b.Max.Y
	p := termenv.ColorProfile()
	str := strings.Builder{}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c, _ := colorful.MakeColor(img.At(x, y))
			color := p.Color(c.Hex())
			str.WriteString(termenv.String("  ").
				Background(color).
				String())
		}
		str.WriteString("\n")
	}
	str.WriteString(fmt.Sprintf("q to quit | %s\n", url))
	return str.String(), nil
}
