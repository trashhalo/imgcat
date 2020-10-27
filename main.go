package main

import (
	"context"
	"fmt"
	"image"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/disintegration/imageorient"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/termenv"
	"github.com/nfnt/resize"
)

const usage = `imgcat [pattern|url]

Examples:
    imgcat path/to/image.jpg
    imgcat *.jpg
    imgcat https://example.com/image.jpg`

func main() {
	if len(os.Args) == 1 {
		fmt.Println(usage)
		os.Exit(1)
	}

	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Println(usage)
		os.Exit(0)
	}

	p := tea.NewProgram(model{urls: os.Args[1:len(os.Args)]})
	p.EnterAltScreen()
	defer p.ExitAltScreen()
	if err := p.Start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

const sparkles = "✨"

type model struct {
	selected int
	urls     []string
	image    string
	width    uint
	height   uint
	err      error

	cancelAnimation context.CancelFunc
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.err != nil {
		if _, ok := msg.(tea.KeyMsg); ok {
			return m, tea.Quit
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = uint(msg.Width)
		m.height = uint(msg.Height)
		return m, load(m.urls[m.selected])
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
		m.err = msg
		return m, nil
	case loadMsg:
		return handleLoadMsg(m, msg)
	case gifMsg:
		return handleGifMsg(m, msg)
	}
	return m, nil
}

func handleGifMsg(m model, msg gifMsg) (model, tea.Cmd) {
	m.image = msg.frames[msg.frame]
	return m, func() tea.Msg {
		nextFrame := msg.frame + 1
		if nextFrame == len(msg.gif.Image) {
			nextFrame = 0
		}
		select {
		case <-msg.ctx.Done():
			return nil
		case <-time.After(time.Duration(msg.gif.Delay[nextFrame]*10) * time.Millisecond):
			return gifMsg{
				ctx:    msg.ctx,
				gif:    msg.gif,
				frames: msg.frames,
				frame:  nextFrame,
			}
		}
	}
}

func handleLoadMsg(m model, msg loadMsg) (model, tea.Cmd) {
	if m.cancelAnimation != nil {
		m.cancelAnimation()
	}

	// blank out image so it says "loading..."
	m.image = ""

	selected := m.urls[m.selected]
	ext := filepath.Ext(selected)
	t := mime.TypeByExtension(ext)
	if strings.Contains(t, "gif") {
		return handleLoadMsgAnimation(m, msg)
	}
	return handleLoadMsgStatic(m, msg)
}

func handleLoadMsgStatic(m model, msg loadMsg) (model, tea.Cmd) {
	defer msg.Close()
	r := msg.Reader()
	url := m.urls[m.selected]
	img, err := readerToImage(m.width, m.height, url, r)
	if err != nil {
		return m, func() tea.Msg { return errMsg{err} }
	}
	m.image = img
	return m, nil
}

func handleLoadMsgAnimation(m model, msg loadMsg) (model, tea.Cmd) {
	defer msg.Close()
	r := msg.Reader()

	// decode the gif
	gimg, err := gif.DecodeAll(r)
	if err != nil {
		return m, wrapErrCmd(err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	m.cancelAnimation = cancel

	// precompute the frames for performance reasons
	var frames []string
	for _, img := range gimg.Image {
		str, err := imageToString(m.width, m.height, m.urls[m.selected], img)
		if err != nil {
			return m, wrapErrCmd(err)
		}
		frames = append(frames, str)
	}

	return m, func() tea.Msg {
		return gifMsg{
			gif:    gimg,
			frames: frames,
			frame:  0,
			ctx:    ctx,
		}
	}
}

func wrapErrCmd(err error) tea.Cmd {
	return func() tea.Msg { return errMsg{err} }
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("couldn't load image(s): %v\n\npress any key to exit", m.err)
	}
	if m.image == "" {
		return fmt.Sprintf("loading %s %s", m.urls[m.selected], sparkles)
	}
	return m.image
}

type gifMsg struct {
	gif    *gif.GIF
	frame  int
	frames []string
	ctx    context.Context
}

type loadMsg struct {
	resp *http.Response
	file *os.File
}

func (l loadMsg) Reader() io.ReadCloser {
	if l.resp != nil {
		return l.resp.Body
	}
	return l.file
}

func (l loadMsg) Close() {
	l.Reader().Close()
}

type errMsg struct{ error }

func load(url string) tea.Cmd {
	if strings.HasPrefix(url, "http") {
		return func() tea.Msg {
			resp, err := http.Get(url)
			if err != nil {
				return errMsg{err}
			}
			return loadMsg{resp: resp}
		}
	}
	return func() tea.Msg {
		file, err := os.Open(url)
		if err != nil {
			return errMsg{err}
		}
		return loadMsg{file: file}
	}
}

func readerToImage(width uint, height uint, url string, r io.Reader) (string, error) {
	img, _, err := imageorient.Decode(r)
	if err != nil {
		return "", err
	}

	return imageToString(width, height, url, img)
}

func imageToString(width, height uint, url string, img image.Image) (string, error) {
	img = resize.Thumbnail(width, height*2-4, img, resize.Lanczos3)
	b := img.Bounds()
	w := b.Max.X
	h := b.Max.Y
	p := termenv.ColorProfile()
	str := strings.Builder{}
	for y := 0; y < h; y += 2 {
		for x := w; x < int(width); x = x + 2 {
			str.WriteString(" ")
		}
		for x := 0; x < w; x++ {
			c1, _ := colorful.MakeColor(img.At(x, y))
			color1 := p.Color(c1.Hex())
			c2, _ := colorful.MakeColor(img.At(x, y+1))
			color2 := p.Color(c2.Hex())
			str.WriteString(termenv.String("▀").
				Foreground(color1).
				Background(color2).
				String())
		}
		str.WriteString("\n")
	}
	str.WriteString(fmt.Sprintf("q to quit | %s\n", url))
	return str.String(), nil
}
