package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

func main() {
	fmt.Println("Starting up...")

	store, err := NewStore()
	if err != nil {
		fmt.Println("Failed to initialize database:", err)
		os.Exit(1)
	}

	m := New(store)
	program := tea.NewProgram(m)

	if _, err := program.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
