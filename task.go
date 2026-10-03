package main

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
)

type Task struct {
	status      status
	title       string
	description string
}

func NewTask(status status, title string, description string) Task {
	return Task{status: status, title: title, description: description}
}

func (t Task) Status() status {
	return t.status
}

func (t Task) Title() string {
	return t.title
}

func (t Task) Description() string {
	return t.description
}

func (t Task) FilterValue() string {
	return t.title
}

func (t *Task) Next() {
	if t.status == done {
		t.status = todo
	} else {
		t.status++
	}
}

// Custom delegate for task list items
type TaskDelegate struct{}

func (d TaskDelegate) Height() int                             { return 2 }
func (d TaskDelegate) Spacing() int                            { return 0 }
func (d TaskDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d TaskDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	task, ok := item.(Task)
	if !ok {
		return
	}

	selected := index == m.Index()

	// Use base styles WITHOUT padding - we control width exactly
	var titleStyle lipgloss.Style
	if selected {
		titleStyle = SelectedTaskStyle.
			Padding(0, 0). // Remove padding - we control width
			Margin(0, 0)
	} else {
		titleStyle = TaskStyle.
			Padding(0, 0).
			Margin(0, 0)
	}

	// Get available width for content (list internal width minus status indicator)
	listWidth := m.Width()
	contentWidth := listWidth - 2
	if contentWidth < 10 {
		contentWidth = 10
	}

	title := task.Title()
	if task.Description() != "" {
		title += "  " + TaskDescStyle.Render(task.Description())
	}

	// Truncate title to fit
	if lipgloss.Width(title) > contentWidth {
		title = lipgloss.NewStyle().Width(contentWidth).Render(title)
	}

	var statusIndicator string
	switch task.Status() {
	case todo:
		statusIndicator = StatusTodoStyle.Render("● ")
	case completing:
		statusIndicator = StatusCompletingStyle.Render("◐ ")
	case done:
		statusIndicator = StatusDoneStyle.Render("✓ ")
	}

	content := statusIndicator + titleStyle.Render(title)
	fmt.Fprint(w, content)
}
