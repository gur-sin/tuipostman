package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	focusMethod = iota
	focusURL
	focusTabs
)

var methods = []string{"GET", "POST", "PUT", "DELETE"}

var tabs = []string{"[h]eaders, [b]ody, [r]esponse"}

var (
	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("5")).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("5"))

	blurredStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Background(lipgloss.Color("0")).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))

	focusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			BorderForeground(lipgloss.Color("205")).
			Border(lipgloss.RoundedBorder())

	placeholderStyle = lipgloss.NewStyle().
				PaddingTop(1).
				PaddingLeft(2).
				Foreground(lipgloss.Color("245"))

	selectedTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("10")).
				Bold(true).
				Underline(true)

	blurredTabStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)

type model struct {
	methodIndex int
	focusIndex  int
	tabIndex    int
	urlInput    textinput.Model
}

func (m model) Init() tea.Cmd {
	m.urlInput = textinput.New()
	m.urlInput.Placeholder = "https://api.example.com"
	m.urlInput.Focus()

	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			m.methodIndex = (m.methodIndex + 1) % len(methods)

		case "right":
			m.methodIndex = (m.methodIndex - 1 + len(methods)) % len(methods)

		case "tab":
			m.focusIndex = (m.focusIndex + 1) % 3

			if m.focusIndex == 1 {
				m.urlInput.Focus()
			} else {
				m.urlInput.Blur()
			}

		case "shift+tab":
			m.focusIndex = (m.focusIndex - 1 + 2) % 3

			if m.focusIndex == 1 {
				m.urlInput.Focus()
			} else {
				m.urlInput.Blur()
			}

		case "h":
			if m.focusIndex == focusTabs {
				m.tabIndex = 0
			}

		case "b":
			if m.focusIndex == focusTabs {
				m.tabIndex = 1
			}

		case "r":
			if m.focusIndex == focusTabs {
				m.tabIndex = 2
			}
		}

		if m.focusIndex == 1 {
			var cmd tea.Cmd
			m.urlInput, cmd = m.urlInput.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m model) View() string {
	var methodStr string

	if m.focusIndex == 0 {
		methodStr = selectedStyle.Render(fmt.Sprintf("[ %s ]", methods[m.methodIndex]))
	} else {
		methodStr = blurredStyle.Render(fmt.Sprintf("[ %s ]", methods[m.methodIndex]))
	}

	var urlStr string
	if m.focusIndex == 1 {
		urlStr = focusedStyle.Render(m.urlInput.View())
	} else {
		urlStr = blurredStyle.Render(m.urlInput.View())
	}

	var tabLabels []string
	for i, tab := range tabs {
		if i == m.tabIndex {
			tabLabels = append(tabLabels, selectedTabStyle.Render(tab))
		} else {
			tabLabels = append(tabLabels, blurredTabStyle.Render(tab))
		}
	}
	tabBar := strings.Join(tabLabels, "   ")

	var tabContent string
	switch m.tabIndex {
	case 0:
		tabContent = placeholderStyle.Render("Headers UI goes here.")
	case 1:
		tabContent = placeholderStyle.Render("Body UI goes here.")
	case 2:
		tabContent = placeholderStyle.Render("Response will be displayed here.")
	}

	return fmt.Sprintf("%s   %s\n%s\n\n%s\n", methodStr, urlStr, tabBar, tabContent)
}

func main() {
	if _, err := tea.NewProgram(NewModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Printf("Uh oh, there was an error: %v\n", err)
		os.Exit(1)
	}
}

func NewModel() model {
	m := model{}
	m.Init() // Call Init to set up the urlInput
	return m
}
