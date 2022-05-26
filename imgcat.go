package imgcat

import (
	"fmt"
	_ "image/jpeg"
	_ "image/png"

	"github.com/trashhalo/imgcat/pkg/image"

	tea "github.com/charmbracelet/bubbletea"
)

const sparkles = "✨"

type Model struct {
	selected int
	urls     []string
	image    image.Model
	err      error
}

func NewModel(urls []string) Model {
	image := image.New(1, 1, urls[0]);
	return Model{
		selected: 0,
		urls: urls,
		image: image,
	}
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
		return m, m.image.Redraw(uint(msg.Width), uint(msg.Height), m.urls[m.selected])
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
			return m, m.image.UpdateUrl(m.urls[m.selected])
		case "k", "up":
			if m.selected-1 != -1 {
				m.selected--
			} else {
				m.selected = len(m.urls) - 1
			}
			return m, m.image.UpdateUrl(m.urls[m.selected])
		}
	case errMsg:
		m.err = msg
		return m, nil
	}
	var cmd tea.Cmd
	m.image, cmd = m.image.Update(msg)
	return m, cmd
}


func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("couldn't load image(s): %v\n\npress any key to exit", m.err)
	}
	if m.image.IsLoading() {
		return fmt.Sprintf("loading %s %s", m.urls[m.selected], sparkles)
	}
	return m.image.View()
}

type errMsg struct{ error }