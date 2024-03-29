package component

import (
	"context"
	"fmt"
	_ "image/jpeg"
	_ "image/png"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	url     string
	image    string
	width    uint
	height   uint
	err      error

	cancelAnimation context.CancelFunc
}

func New(width, height uint, url string) Model {
	return Model{
		width: width,
		height: height,
		url: url,	
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case errMsg:
		m.err = msg
		return m, nil
	case rewdrawMsg:
		m.width = msg.width
		m.height = msg.height
		m.url = msg.url
		return m, loadUrl(m.url)
	case loadMsg:
		return handleLoadMsg(m, msg)
	case gifMsg:
		return handleGifMsg(m, msg)
	}
	return m, nil
}

func wrapErrCmd(err error) tea.Cmd {
	return func() tea.Msg { return errMsg{err} }
}

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("couldn't load image(s): %v", m.err)
	}
	return m.image
}

type errMsg struct{ error }

func (m Model) Redraw(width uint, height uint, url string) tea.Cmd {
	return func() tea.Msg {
		return rewdrawMsg{
			width: width,
			height: height,
			url: url,
		}
	}
}

func (m Model) UpdateUrl(url string) tea.Cmd {
	return func() tea.Msg {
		return rewdrawMsg{
			width: m.width,
			height: m.height,
			url: url,
		}
	}
}

type rewdrawMsg struct {
	width uint
	height uint
	url string
}

func (m Model) IsLoading() bool {
	return m.image == ""
}