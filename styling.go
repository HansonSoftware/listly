package main

import (
	"image/color"
	"os"

	"charm.land/lipgloss/v2"
)

// Terminal colors follow the user's own terminal theme: the palette below
// references the ANSI basic colors (indices 0-15), which the terminal maps
// to its configured colors. Set NO_COLOR, or run with TERM=dumb, to fall
// back to plain uncolored output.
// pick maps a terminal color to plain output when colors are unavailable.
func pick(c color.Color) color.Color {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return lipgloss.NoColor{}
	}
	return c
}

var (
	ColorBg          color.Color = lipgloss.NoColor{} // inherit terminal background
	ColorFg          color.Color = lipgloss.NoColor{} // inherit terminal foreground
	ColorMuted       color.Color = pick(lipgloss.BrightBlack)
	ColorPrimary     color.Color = pick(lipgloss.Blue)
	ColorSecondary   color.Color = pick(lipgloss.Magenta)
	ColorSuccess     color.Color = pick(lipgloss.Green)
	ColorWarning     color.Color = pick(lipgloss.Yellow)
	ColorError       color.Color = pick(lipgloss.Red)
	ColorBorder      color.Color = pick(lipgloss.BrightBlack)
	ColorBorderFocus color.Color = pick(lipgloss.Blue)
	ColorCardBg      color.Color = lipgloss.NoColor{}
	ColorSelection   color.Color = pick(lipgloss.BrightBlack)
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

	// BoardTitleStyle is the session title shown above the board.
	BoardTitleStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			Padding(0, 1)

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
			Italic(true) // same fg as the title text so it stays readable on the selection highlight

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

	// ToastStyle is the compact variant of CardStyle for corner overlays.
	ToastStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Background(ColorCardBg)

	TextAreaStyle = lipgloss.NewStyle().
			Foreground(ColorFg).
			Background(ColorBg)

	FocusedTextAreaStyle = lipgloss.NewStyle().
				Foreground(ColorFg).
				Background(ColorBg)

	// Thanks to https://www.asciiart.eu/text-to-ascii-art
	Logo = `
 ▄█        ▄█     ▄████████     ███      ▄█       ▄██   ▄  
███       ███    ███    ███ ▀█████████▄ ███       ███   ██▄
███              ███    █▀     ▀███▀▀██ ███       ███▄▄▄███
███       ███▌   ███            ███   ▀ ███       ▀▀▀▀▀▀███
███       ███▌ ▀███████████     ███     ███       ▄██   ███
███       ███           ███     ███     ███       ███   ███
███▌    ▄ ███     ▄█    ███     ███     ███▌    ▄ ███   ███
█████▄▄██ █▀    ▄████████▀     ▄████▀   █████▄▄██  ▀█████▀ 
▀                                       ▀                  
`

	// LogoStyle renders the wordmark.
	LogoStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)
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

	WelcomePromptStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true)

	SessionCountStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				MarginBottom(1)

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
