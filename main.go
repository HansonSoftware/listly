package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	fmt.Println("Starting up...")
	initializeDB()

	models = []tea.Model{New(), NewForm(todo)}
	m := models[model]
	program := tea.NewProgram(m, tea.WithAltScreen())
	tea.SetWindowTitle("Listly")

	if _, err := program.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
