package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
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

var tabs = []string{"Headers", "Body", "Response"}

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

type headerField struct {
	key   textinput.Model
	value textinput.Model
}

type model struct {
	methodIndex int
	focusIndex  int
	tabIndex    int
	urlInput    textinput.Model

	headers          []headerField
	headerFocusIndex int
	bodyInput        textinput.Model

	response     string
	errorMessage error
}

type responseMsg struct {
	body     string
	err      error
	exitCode int
}

func (m model) Init() tea.Cmd {
	return m.urlInput.Focus()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case responseMsg:
		m.response = msg.body
		if msg.err != nil {
			m.errorMessage = msg.err
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "left":
			if m.focusIndex == focusMethod {
				m.methodIndex = (m.methodIndex - 1 + len(methods)) % len(methods)
			}

		case "right":
			if m.focusIndex == focusMethod {
				m.methodIndex = (m.methodIndex + 1) % len(methods)
			}

		case "tab":
			m.focusIndex = (m.focusIndex + 1) % 3

			m.urlInput.Blur()

			if m.focusIndex == focusURL {
				cmds = append(cmds, m.urlInput.Focus())
			}

		case "shift+tab":
			m.focusIndex = (m.focusIndex - 1 + 3) % 3

			m.urlInput.Blur()

			if m.focusIndex == focusURL {
				cmds = append(cmds, m.urlInput.Focus())
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

		case "ctrl+n":
			if m.focusIndex == focusTabs && m.tabIndex == 0 {
				newKey := textinput.New()
				newKey.Placeholder = "key"
				newKey.Width = 10
				newKey.Prompt = ""

				newValue := textinput.New()
				newValue.Placeholder = "value"
				newValue.Width = 20
				newValue.Prompt = ""

				newHeader := headerField{key: newKey, value: newValue}
				m.headers = append(m.headers, newHeader)

				m.headerFocusIndex = len(m.headers)*2 - 2 // focus on new key field
				return m, newHeader.key.Focus()
			}

		case "up":
			if m.focusIndex == focusTabs && m.tabIndex == 0 {
				if m.headerFocusIndex > 0 {
					m.headerFocusIndex--
				}
			}

		case "down":
			if m.focusIndex == focusTabs && m.tabIndex == 0 {
				if m.headerFocusIndex < len(m.headers)*2-1 {
					m.headerFocusIndex++
				}
			}
		}

		if msg.String() == "ctrl+r" {
			if m.focusIndex == focusTabs && m.tabIndex == 1 {
				m.tabIndex = 2 // switch to Response tab after sending
				return m, sendRequest(m)
			}
		}
	}

	if m.focusIndex == focusURL {
		var urlCmd tea.Cmd
		m.urlInput, urlCmd = m.urlInput.Update(msg)
		cmds = append(cmds, urlCmd)
	}

	if m.focusIndex == focusTabs && m.tabIndex == 0 {
		i := m.headerFocusIndex / 2
		j := m.headerFocusIndex % 2 // 0 = key, 1 = value

		var cmd tea.Cmd
		if i < len(m.headers) {
			if j == 0 {
				m.headers[i].key.Focus()
				m.headers[i].value.Blur()
				m.headers[i].key, cmd = m.headers[i].key.Update(msg)
			} else {
				m.headers[i].value.Focus()
				m.headers[i].key.Blur()
				m.headers[i].value, cmd = m.headers[i].value.Update(msg)
			}
			return m, cmd
		}
	}

	if m.focusIndex == focusTabs && m.tabIndex == 1 {
		var cmd tea.Cmd
		m.bodyInput, cmd = m.bodyInput.Update(msg)
		return m, cmd
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	var methodStr string

	if m.focusIndex == focusMethod {
		methodStr = selectedStyle.Render(methods[m.methodIndex])
	} else {
		methodStr = blurredStyle.Render(methods[m.methodIndex])
	}

	var urlStr string
	if m.focusIndex == focusURL {
		urlStr = focusedStyle.Render(m.urlInput.View())
	} else {
		urlStr = blurredStyle.Render(m.urlInput.View())
	}

	var tabLabels []string
	for i, tab := range tabs {
		var renderedTab string
		if m.focusIndex == focusTabs && i == m.tabIndex {
			renderedTab = selectedTabStyle.Render(tab)
		} else {
			renderedTab = blurredTabStyle.Render(tab)
		}
		tabLabels = append(tabLabels, renderedTab)
	}
	tabBar := strings.Join(tabLabels, "    ")

	var tabContent string
	switch m.tabIndex {
	case 0:
		var headerLines []string
		for i, h := range m.headers {
			if m.focusIndex == focusTabs && m.tabIndex == 0 {
				if m.headerFocusIndex == i*2 {
					h.key.PromptStyle = focusedStyle
				} else {
					h.key.PromptStyle = blurredStyle
				}

				if m.headerFocusIndex == i*2+1 {
					h.value.PromptStyle = focusedStyle
				} else {
					h.value.PromptStyle = blurredStyle
				}
			}

			kView := h.key.View()
			vView := h.value.View()
			line := lipgloss.JoinHorizontal(lipgloss.Top,
				lipgloss.NewStyle().Width(17).Render(kView),
				lipgloss.NewStyle().Width(2).Render(":"),
				vView,
			)
			headerLines = append(headerLines, line)
		}

		tabContent = lipgloss.NewStyle().
			PaddingTop(1).
			PaddingLeft(2).
			Render(strings.Join(headerLines, "\n"))
	case 1:
		bodyView := m.bodyInput.View()
		if m.focusIndex == focusTabs {
			tabContent = focusedStyle.Render(bodyView)
		} else {
			tabContent = blurredStyle.Render(bodyView)
		}

	case 2:
		if m.errorMessage != nil && m.errorMessage.Error() != "" {
			tabContent = focusedStyle.Render("Error: " + m.errorMessage.Error())
		} else {
			tabContent = placeholderStyle.Render(m.response)
		}
	}

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, methodStr, " ", urlStr)

	return lipgloss.JoinVertical(lipgloss.Left,
		topRow,
		tabBar,
		"",
		tabContent,
	) + "\n"
}

func main() {
	p := tea.NewProgram(NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Uh oh, there was an error: %v\n", err)
		os.Exit(1)
	}
}

func NewModel() model {
	ti := textinput.New()
	ti.Placeholder = "https://api.example.com"
	ti.Width = 30
	ti.Prompt = ""

	key := textinput.New()
	key.Placeholder = "key"
	key.Width = 20
	key.Prompt = ""
	key.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	key.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	value := textinput.New()
	value.Placeholder = "value"
	value.Width = 30
	value.Prompt = ""
	value.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	value.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	headers := []headerField{{key: key, value: value}}

	bi := textinput.New()
	bi.Placeholder = "Raw request body"
	bi.Width = 30
	bi.Prompt = ""
	bi.Focus()

	return model{
		methodIndex: 0,
		focusIndex:  focusURL,
		tabIndex:    0,
		urlInput:    ti,

		headers:          headers,
		headerFocusIndex: 0,
		bodyInput:        bi,

		response:     "",
		errorMessage: errors.New(""),
	}
}

func sendRequest(m model) tea.Cmd {
	return func() tea.Msg {
		method := methods[m.methodIndex]
		url := strings.TrimSpace(m.urlInput.Value())

		if url == "" {
			return responseMsg{body: "", err: errors.New("URL cannot be empty")}
		}

		args := []string{"--silent", "-X", method, url}

		for _, h := range m.headers {
			key := strings.TrimSpace(h.key.Value())
			value := strings.TrimSpace(h.value.Value())

			if key != "" {
				headerStr := fmt.Sprintf("%s: %s", key, value)
				args = append(args, "-H", headerStr)
			}
		}

		if method == "POST" || method == "PUT" || method == "PATCH" || method == "DELETE" {
			body := m.bodyInput.Value()

			if trimmed := strings.TrimSpace(body); trimmed != "" {
				args = append(args, "--data", trimmed)
			}
		}

		cmd := exec.Command("curl", args...)

		var outBuf, errBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf

		err := cmd.Run()

		resp := responseMsg{
			body: outBuf.String(),
			err:  nil,
		}

		if err != nil {
			resp.err = fmt.Errorf("request failed: %v", err)
		} else if errBuf.Len() > 0 {
			resp.err = fmt.Errorf(errBuf.String())
		}

		return resp
	}
}
