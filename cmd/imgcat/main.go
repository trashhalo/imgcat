package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	imgcat "github.com/trashhalo/imgcat/lib"
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

	m := imgcat.NewModel()

	for _, uri := range os.Args[1:len(os.Args)] {
		if strings.HasPrefix(uri, "http") {
			m.AddImageLoader(imgcat.HttpImage{URL: uri})
		} else {
			m.AddImageLoader(imgcat.FileImage{Filename: uri})
		}
	}

	p := tea.NewProgram(m)
	p.EnterAltScreen()
	defer p.ExitAltScreen()
	if err := p.Start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
