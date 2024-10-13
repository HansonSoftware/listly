package main

import (
	"github.com/charmbracelet/bubbles/list"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	list list.Model
	err  error
}

func New() *Model {
	return &Model{}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) View() string {
	return m.list.View()
}

func (m *Model) CreateList(width int, height int) {
	m.list = list.New([]list.Item{}, list.NewDefaultDelegate(), width, height)

	m.list.Title = "Today's List"

	// NOTE: Dummy items for now
	m.list.SetItems([]list.Item{
		Task{status: done, title: "meeting @ 9:00am", description: "engineering team stand-up"},
		Task{status: completing, title: "implement client feedback", description: "substation modeling"},
		Task{status: todo, title: "complete c# training", description: "update cs-neetcode repo"},
		Task{status: todo, title: "lunch", description: "sushi @ 12:00pm"},
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.CreateList(msg.Width, msg.Height)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	return m, cmd
}
