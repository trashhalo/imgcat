package lib

import (
	"context"
	"fmt"
	"image"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/termenv"
	"github.com/nfnt/resize"
)

type Model struct {
	selected int
	entries  []ImageLoader
	image    string
	width    uint
	height   uint
	err      error

	cancelAnimation context.CancelFunc
}

type ImageLoader interface {
	Image() (image.Image, error)
}

type GifLoader interface {
	ImageLoader
	Gif() (*gif.GIF, error)
}

// LoadingMsg implement this interface in your ImageLoader if you want to set a loading message which will be displayed
// before Image() or Gif() (from GifLoader) have returned.
type LoadingMsg interface {
	LoadingMsg() string
}

// FooterMsg implement this interface in your ImageLoader if you want to set a footer message below the displayed image
type FooterMsg interface {
	Footer() string
}

func NewModel() Model {
	return Model{}
}

func (m *Model) AddImage(img image.Image) {
	m.entries = append(m.entries, &staticImage{img})
}

func (m *Model) AddImageLoader(img ImageLoader) {
	m.entries = append(m.entries, img)
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.err != nil {
		if _, ok := msg.(tea.KeyMsg); ok {
			return m, tea.Quit
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = uint(msg.Width)
		m.height = uint(msg.Height)
		return m, load(m.entries[m.selected])
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "j", "down":
			if m.selected+1 != len(m.entries) {
				m.selected++
			} else {
				m.selected = 0
			}
			return m, load(m.entries[m.selected])
		case "k", "up":
			if m.selected-1 != -1 {
				m.selected--
			} else {
				m.selected = len(m.entries) - 1
			}
			return m, load(m.entries[m.selected])
		}
	case errMsg:
		m.err = msg
		return m, nil
	case ImageLoader:
		return handleLoadMsg(m, msg)
	case gifMsg:
		return handleGifMsg(m, msg)
	}
	return m, nil
}

func handleGifMsg(m Model, msg gifMsg) (Model, tea.Cmd) {
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

func handleLoadMsg(m Model, loader ImageLoader) (Model, tea.Cmd) {
	if m.cancelAnimation != nil {
		m.cancelAnimation()
	}

	// blank out image so it says "loading..."
	m.image = ""

	// if our loader is implementing GifLoader and Gif() actually returns
	// non-nil we will use the gif handler, otherwise we'll use the static
	// image solution
	if gifl, ok := loader.(GifLoader); ok {
		gimg, err := gifl.Gif()
		if err != nil {
			return m, wrapErrCmd(err)
		} else if gimg != nil {
			ctx := context.Background()
			ctx, cancel := context.WithCancel(ctx)
			m.cancelAnimation = cancel

			// precompute the frames for performance reasons
			var frames []string
			for _, img := range gimg.Image {
				str := imageToString(m.width, m.height, img, loader)
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
	}

	img, err := loader.Image()
	if err != nil {
		return m, wrapErrCmd(err)
	}
	m.image = imageToString(m.width, m.height, img, loader)
	return m, nil
}

func wrapErrCmd(err error) tea.Cmd {
	return func() tea.Msg { return errMsg{err} }
}

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("couldn't load image(s): %v\n\npress any key to exit", m.err)
	}
	if m.image == "" {
		entry := m.entries[m.selected]
		if loading, ok := entry.(LoadingMsg); ok {
			return loading.LoadingMsg()
		}
		return ""
	}
	return m.image
}

type gifMsg struct {
	gif    *gif.GIF
	frame  int
	frames []string
	ctx    context.Context
}

type errMsg struct{ error }

func load(img ImageLoader) tea.Cmd {
	return func() tea.Msg {
		return img
	}
}

func imageToString(width, height uint, img image.Image, loader ImageLoader) string {
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
	str.WriteString("q to quit")
	if footer, ok := loader.(FooterMsg); ok {
		str.WriteString(fmt.Sprintf(" | %s", footer.Footer()))
	}
	str.WriteString("\n")
	return str.String()
}
