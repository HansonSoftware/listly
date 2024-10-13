package main

import (
	css "github.com/charmbracelet/lipgloss"
)

var (
	columnStyle = css.NewStyle().
			Padding(1, 2).
			Border(css.HiddenBorder())
	focusedStyle = css.NewStyle().
			Padding(1, 2).
			Border(css.RoundedBorder()).
			BorderForeground(css.Color("#a6e3a1"))
	helpStyle = css.NewStyle().
			Foreground(css.Color("240"))
)
