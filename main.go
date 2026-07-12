package main

import (
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kuroiko0429/tail-pulse/config"
	"github.com/kuroiko0429/tail-pulse/ui"
)

func main() {
	cfg := config.LoadConfig()
	m := ui.NewModel(cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Fatal: %v\n", err)
		os.Exit(1)
	}

	updatedModel := finalModel.(ui.Model)
	target := updatedModel.GetSSHTarget()
	if target != "" {
		fmt.Printf("\033[32m[SYS] Handing over to native SSH client for %s...\033[0m\n", target)
		cmd := exec.Command("ssh", target)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run()
	}
}