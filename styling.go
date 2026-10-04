package main

import (
	"charm.land/lipgloss/v2"
)

// Color palette
var (
	ColorBg          = lipgloss.Color("#1a1b26")
	ColorFg          = lipgloss.Color("#c0caf5")
	ColorMuted       = lipgloss.Color("#565f89")
	ColorPrimary     = lipgloss.Color("#7aa2f7")
	ColorSecondary   = lipgloss.Color("#bb9af7")
	ColorSuccess     = lipgloss.Color("#9ece6a")
	ColorWarning     = lipgloss.Color("#e0af68")
	ColorError       = lipgloss.Color("#f7768e")
	ColorBorder      = lipgloss.Color("#292e42")
	ColorBorderFocus = lipgloss.Color("#7aa2f7")
	ColorCardBg      = lipgloss.Color("#16161e")
	ColorSelection   = lipgloss.Color("#2e3c64")
)

// Base styles
var (
	BaseStyle = lipgloss.NewStyle().
			Foreground(ColorFg).
			Background(ColorBg)

	// Column styles - both have rounded borders for alignment
	ColumnStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Background(ColorBg)

	FocusedColumnStyle = lipgloss.NewStyle().
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorBorderFocus).
				Background(ColorBg)

	// Title styles for columns
	ColumnTitleStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true).
				Padding(0, 1).
				MarginBottom(1)

	FocusedColumnTitleStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true).
				Padding(0, 1).
				MarginBottom(1).
				Background(ColorSelection)

	// Task item styles
	TaskStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Margin(0, 0, 0, 0)

	SelectedTaskStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Background(ColorSelection).
				Foreground(ColorFg)

	TaskTitleStyle = lipgloss.NewStyle().
			Foreground(ColorFg).
			Bold(true)

	TaskDescStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true)

	// Card/overlay styles for floating forms
	CardStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Background(ColorCardBg).
			Width(60)

	CardTitleStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			MarginBottom(1)

	InputStyle = lipgloss.NewStyle().
			Foreground(ColorFg).
			Background(ColorBg).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder)

	FocusedInputStyle = lipgloss.NewStyle().
				Foreground(ColorFg).
				Background(ColorBg).
				Padding(0, 1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary)

	TextAreaStyle = lipgloss.NewStyle().
			Foreground(ColorFg).
			Background(ColorBg).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder)

	FocusedTextAreaStyle = lipgloss.NewStyle().
				Foreground(ColorFg).
				Background(ColorBg).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary)

	// Welcome screen styles
	WelcomeTitleStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true).
				MarginBottom(1).
				Align(lipgloss.Center)

	WelcomeSubtitleStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				MarginBottom(2).
				Align(lipgloss.Center)

	SessionItemStyle = lipgloss.NewStyle().
				Padding(0, 2).
				Margin(0, 1).
				Foreground(ColorFg)

	SelectedSessionStyle = lipgloss.NewStyle().
				Padding(0, 2).
				Margin(0, 1).
				Background(ColorSelection).
				Foreground(ColorFg).
				Bold(true)

	DailySessionBadge = lipgloss.NewStyle().
				Foreground(ColorWarning).
				MarginLeft(1)

	WelcomeHelpStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				MarginTop(2).
				Align(lipgloss.Center)

	// Help text
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			MarginTop(1)

	// ErrorBannerStyle surfaces persistence errors to the user
	ErrorBannerStyle = lipgloss.NewStyle().
		Foreground(ColorError).
		Bold(true).
		Padding(0, 1)

	// Status indicators
	StatusTodoStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	StatusCompletingStyle = lipgloss.NewStyle().
				Foreground(ColorWarning).
				Bold(true)

	StatusDoneStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)
)

// Helper to center content in available space
func CenterIn(width, height int, content string) string {
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

// Helper to center horizontally
func CenterHorizontal(width int, content string) string {
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, content)
}

// Helper for vertical centering
func CenterVertical(height int, content string) string {
	return lipgloss.PlaceVertical(height, lipgloss.Center, content)
}

// Help keybind style
var HelpKeybindStyle = lipgloss.NewStyle().
	Foreground(ColorPrimary).
	Bold(true)

var HelpDescStyle = lipgloss.NewStyle().
	Foreground(ColorMuted)
