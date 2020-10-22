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
	p := tea.NewProgram(model{url: os.Args[1]})
	p.EnterAltScreen()
	err := p.Start()
	p.ExitAltScreen()
	if err != nil {
		panic(err)
	}
}

const sparkles = "✨"

type model struct {
	url   string
	image string
}

func (m model) Init() tea.Cmd {
	return load(m.url)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case errMsg:
		return model{url: m.url, image: msg.Error()}, nil
	case loadMsg:
		defer msg.Body.Close()
		img, err := readerToImage(msg.Body)
		if err != nil {
			return m, func() tea.Msg { return errMsg(err) }
		}
		return model{url: m.url, image: img}, nil
	}
	return m, nil
}

func (m model) View() string {
	if m.image == "" {
		return fmt.Sprintf("loading %s %s", m.url, sparkles)
	}
	return m.image
}

type loadMsg *http.Response
type errMsg error

func load(url string) tea.Cmd {
	return func() tea.Msg {
		resp, err := http.Get(url)
		if err != nil {
			return errMsg(err)
		}
		return loadMsg(resp)
	}
}

func readerToImage(r io.Reader) (string, error) {
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
	str.WriteString("q to quit")
	return str.String(), nil
}
