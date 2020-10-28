package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/trashhalo/imgcat"
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

	p := tea.NewProgram(imgcat.NewModel(os.Args[1:len(os.Args)]))
	p.EnterAltScreen()
	defer p.ExitAltScreen()
	if err := p.Start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
