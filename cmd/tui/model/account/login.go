// Package account provides a model for the login page of the TUI application.
package account

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prionis/dns-server/cmd/tui/model/popup"
	"github.com/prionis/dns-server/cmd/tui/structs"
	"github.com/prionis/dns-server/cmd/tui/style"
	"github.com/prionis/dns-server/cmd/tui/transport"
)

// LoginModel represents the state and behavior of the login page in the TUI application.
type LoginModel struct {
	// Width and height of the screen.
	width, height int
	// Focused input field
	focusIndex int
	// Domain, data, type, class and ttl inputs
	inputFields []textinput.Model

	// Add, Clear and Cancel buttons.
	buttons []string
	// Selected button.
	cursor int
	// Are buttons focused or the inputs.
	focusButtons bool

	transport *transport.Transport
}

// NewLoginModel creates a new login model.
func NewLoginModel(t *transport.Transport, w, h int) LoginModel {
	inputs := make([]textinput.Model, 0, 2)

	loginInput := textinput.New()
	loginInput.Placeholder = "Login"
	loginInput.Width = w / 3
	loginInput.Focus()
	inputs = append(inputs, loginInput)

	passwordInput := textinput.New()
	passwordInput.Placeholder = "Password"
	passwordInput.Width = w / 3
	passwordInput.EchoMode = textinput.EchoPassword
	passwordInput.EchoCharacter = '*'
	inputs = append(inputs, passwordInput)

	return LoginModel{
		width:       w,
		height:      h,
		inputFields: inputs,
		buttons:     []string{"Login", "Exit"},
		transport:   t,
	}
}

// Init initializes the model.
func (dm LoginModel) Init() tea.Cmd {
	return nil
}

// Update handles messages that coming from the user interactions.
func (m LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "up", "shift+tab", "down", "tab":
			m = processVerticals(m, msg.String())
		case "left":
			if m.focusButtons && m.cursor > 0 {
				m.cursor--
			}
		case "right":
			if m.focusButtons && m.cursor < len(m.buttons)-1 {
				m.cursor++
			}
		case "ctrl+c", "esc":
			return m, func() tea.Msg {
				return LoginCancelMsg{}
			}
		case "enter", " ":
			var cmd tea.Cmd
			m, cmd = processEnter(m, msg)
			cmds = append(cmds, cmd)
		default:
			if !m.focusButtons {
				var cmd tea.Cmd
				m.inputFields[m.focusIndex], cmd = m.inputFields[m.focusIndex].Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}
	return m, tea.Batch(cmds...)
}

func processVerticals(m LoginModel, button string) LoginModel {
	switch button {
	case "up", "shift+tab":
		if !m.focusButtons {
			if m.focusIndex > 0 {
				m.inputFields[m.focusIndex].Blur()
				m.focusIndex--
				m.inputFields[m.focusIndex].Focus()
			}
		} else {
			m.focusButtons = false
			m.focusIndex = len(m.inputFields) - 1
			m.inputFields[len(m.inputFields)-1].Focus()
		}

	case "down", "tab":
		if !m.focusButtons {
			if m.focusIndex < len(m.inputFields)-1 {
				m.inputFields[m.focusIndex].Blur()
				m.focusIndex++
				m.inputFields[m.focusIndex].Focus()
			} else {
				m.inputFields[m.focusIndex].Blur()
				m.focusButtons = true
				m.cursor = 0
			}
		} else {
			m.focusButtons = false
			m.focusIndex = 0
			m.inputFields[0].Focus()
		}
	}
	return m
}

func processEnter(m LoginModel, msg tea.KeyMsg) (LoginModel, tea.Cmd) {
	if m.focusButtons {
		if m.cursor == 0 { // Yes button
			login := m.inputFields[0].Value()
			password := m.inputFields[1].Value()

			user, err := m.login(login, password)
			if err != nil {
				return m, popup.Error(fmt.Sprintf("can't login: %v", err), "")
			}

			return m, tea.Batch(
				popup.Success(fmt.Sprintf("Hi %s %s!", user.FirstName, user.LastName), ""),
				func() tea.Msg {
					return LoginSuccessMsg{
						User: user,
					}
				},
			)

		} else { // No button
			return m, func() tea.Msg {
				return LoginCancelMsg{}
			}
		}
	}
	var cmd tea.Cmd
	m.inputFields[m.focusIndex], cmd = m.inputFields[m.focusIndex].Update(msg)
	return m, cmd
}

func (m LoginModel) login(login, password string) (*structs.User, error) {
	user, err := m.transport.Login(login, password)
	if err != nil {
		return nil, fmt.Errorf("can't login: %w", err)
	}
	return &structs.User{
		Login:     user.Login,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
	}, nil
}

// View renders page on the screen.
func (m LoginModel) View() string {
	s := strings.Builder{}
	s.WriteString("\n")

	for _, i := range m.inputFields {
		if i.Focused() {
			s.WriteString(style.FocusedInputStyle.Render(i.View()))
		} else {
			s.WriteString(style.BlurredInputStyle.Render(i.View()))
		}
		s.WriteRune('\n')
	}
	s.WriteRune('\n')

	styledButtons := make([]string, len(m.buttons))
	for i, btn := range m.buttons {
		if m.cursor == i {
			if m.focusButtons {
				styledButtons[i] = style.SelectedButtonStyle.Render(btn)
			} else {
				styledButtons[i] = style.SecondarySelectedButtonStyle.Render(btn)
			}
		} else {
			styledButtons[i] = style.ButtonStyle.Render(btn)
		}
	}
	buttonsAlignCenter := lipgloss.NewStyle().Width(m.width - 4).Align(lipgloss.Center)
	s.WriteString(buttonsAlignCenter.Render(lipgloss.JoinHorizontal(lipgloss.Top, styledButtons...)))

	return lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center).Render(s.String())
}

// LoginSuccessMsg is a message that is sent when the user successfully logs in.
type LoginSuccessMsg struct {
	User *structs.User
}

// LoginCancelMsg is a message that is sent when the user cancels the login process.
type LoginCancelMsg struct{}
