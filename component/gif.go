package component

import (
	"context"
	"image/gif"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type gifMsg struct {
	gif    *gif.GIF
	frame  int
	frames []string
	ctx    context.Context
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