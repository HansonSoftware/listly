package main

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
