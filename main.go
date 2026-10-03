package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	fmt.Println("Starting up...")

	store, err := NewStore()
	if err != nil {
		fmt.Println("Failed to initialize database:", err)
		os.Exit(1)
	}

	models = []tea.Model{New(store), NewForm(todo), NewSessionNameForm(nil)}
	m := models[model]
	program := tea.NewProgram(m, tea.WithAltScreen())
	tea.SetWindowTitle("Listly")

	if _, err := program.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
