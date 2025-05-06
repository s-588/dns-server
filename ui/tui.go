package ui

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
	"github.com/prionis/dns-server/protocol"
	"github.com/prionis/dns-server/sqlite"
)

var (
	purpleColor = lipgloss.Color("#8839ef")
	textColor   = lipgloss.Color("")
	greenColor  = lipgloss.Color("#40a02b")
	blueColor   = lipgloss.Color("#1e66f5")
	pinkColor   = lipgloss.Color("#ff87d7")
	redColor    = lipgloss.Color("#d20f39")

	baseBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(purpleColor).
			Padding(1, 2) // padding inside the border

	unselectedBoarderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(pinkColor)

	selectedBoarderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(purpleColor)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Background(purpleColor).
			Padding(0, 2).
			Align(lipgloss.Center).
			Width(50)

	buttonStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(textColor)

	selectedButtonStyle = buttonStyle.
				Background(purpleColor).
				Bold(true)

	secondarySelectedButtonStyle = buttonStyle.
					Background(pinkColor).
					Bold(true)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Padding(1, 2)

	logStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00"))

	baseStyle = lipgloss.NewStyle().Padding(1, 2)

	recordStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).
			Bold(true)

	recordDetailStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF"))
)

const (
	focusTabs = iota
	focusButtons
	focusTable
)

type model struct {
	width  int
	height int

	sockConn net.Conn
	msg      chan map[string]any

	focusLayer  int
	tabs        []tab
	selectedTab int

	logTable table.Model
	rrTable  table.Model
}

type tab struct {
	name    string
	buttons []string
	cursor  int
}

func NewModel() model {
	conn, err := MakeSockConn()
	if err != nil {
		slog.Error("can't connect to socket", "error", err)
	}

	w, h, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		slog.Error("can't get term size")
	}

	db, err := sqlite.NewDB()
	if err != nil {
		slog.Error("can't connect to database", "error", err)
	}

	dbRRs, err := db.GetAllRRs()
	if err != nil {
		slog.Error("can't get records from database", "error", err)
	}

	cols := []table.Column{{"ID", 4}, {"domain", 10}, {"data", 10}, {"type", 4}, {"class", 5}, {"TimeToLive", 8}}
	rows := make([]table.Row, 0, len(dbRRs))
	for _, rr := range dbRRs {
		rows = append(rows, table.Row{
			strconv.FormatInt(rr.ID, 10),
			rr.RR.Domain,
			net.IP(rr.RR.Data).String(),
			protocol.KeyByValue(protocol.Types, rr.RR.Type),
			protocol.KeyByValue(protocol.Classes, rr.RR.Class),
			strconv.FormatInt(int64(rr.RR.TimeToLive), 10),
		})
	}
	rrTable := table.New(table.WithColumns(cols), table.WithRows(rows), table.WithHeight(10))

	rows = make([]table.Row, 0)
	file, err := os.Open("DNSServer.log")
	if err != nil {
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		log := make(map[string]any)
		err := json.Unmarshal([]byte(line), &log)
		if err != nil {
			continue
		}
		t, err := time.Parse(time.RFC3339, log["time"].(string))
		if err != nil {
			t = time.Now()
		}
		rows = append(rows, table.Row{
			t.Format(time.DateTime),
			log["level"].(string),
			log["msg"].(string),
		})
	}
	cols = []table.Column{{"time", 14}, {"level", 5}, {"message", 20}}
	logTable := table.New(table.WithColumns(cols), table.WithRows(rows), table.WithHeight(10))

	return model{
		width:  w / 3,
		height: h / 2,
		tabs: []tab{
			{
				name:    "Logs",
				buttons: []string{"View", "Filter", "Sort"},
			},
			{
				name:    "Records",
				buttons: []string{"View", "Add", "Filter", "Sort"},
			},
		},
		msg:      make(chan map[string]any),
		sockConn: conn,

		rrTable:  rrTable,
		logTable: logTable,
	}
}

func (m model) Close() {
	close(m.msg)
	m.sockConn.Close()
}

func (m model) Init() tea.Cmd {
	go m.readSocket()
	return waitForMsg(m.msg)
}

func waitForMsg(msgs chan map[string]any) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-msgs
		if !ok {
			return nil
		}
		return logMsg(msg)
	}
}

func (m model) readSocket() {
	buf := make([]byte, 1024)
	log := make(map[string]any)

	for {
		n, err := m.sockConn.Read(buf)
		if err != nil {
			close(m.msg)
			return
		}
		err = json.Unmarshal(buf[:n], &log)
		m.msg <- log
	}
}

type logMsg map[string]any

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case logMsg:
		// Append new log line
		t, err := time.Parse(time.RFC3339, msg["time"].(string))
		if err != nil {
			t = time.Now()
		}
		level := strings.ToUpper(msg["level"].(string))
		message := msg["msg"].(string)

		// Update viewport content and scroll to bottom
		m.logTable.SetRows(append(m.logTable.Rows(), table.Row{t.Format(time.DateTime), level, message}))
		m.logTable.UpdateViewport()
		return m, waitForMsg(m.msg)

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q", "esc":
			if m.focusLayer == focusTable {
				if m.getTable().Focused() {
					m.focusLayer--
					m.getTable().Blur()
					return m, nil
				}
			}
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.focusLayer != focusTable {
				if m.focusLayer > 0 {
					m.focusLayer--
				}
			} else {
				return m.updateTable(msg)
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.focusLayer != focusTable {
				if m.focusLayer < 2 {
					m.focusLayer++
				}
				if m.focusLayer == focusTable {
					if m.selectedTab == 0 {
						m.logTable.Focus()
					}
					if m.selectedTab == 1 {
						m.rrTable.Focus()
					}
				}
			} else {
				return m.updateTable(msg)
			}

		case "left", "l":
			switch m.focusLayer {
			case focusTabs:
				if m.selectedTab < len(m.tabs)-1 {
					m.selectedTab++
				}
			case focusButtons:
				if m.tabs[m.selectedTab].cursor < len(m.tabs[m.selectedTab].buttons)-1 {
					m.tabs[m.selectedTab].cursor++
				}
			case focusTable:
				if m.selectedTab == 0 {
					m.logTable, cmd = m.logTable.Update(msg)
					return m, cmd
				}
				if m.selectedTab == 1 {
					m.rrTable, cmd = m.rrTable.Update(msg)
					return m, cmd
				}
			}

		case "right", "h":
			switch m.focusLayer {
			case focusTabs:
				if m.selectedTab > 0 {
					m.selectedTab--
				}
			case focusButtons:
				if m.tabs[m.selectedTab].cursor > 0 {
					m.tabs[m.selectedTab].cursor--
				}
			case focusTable:
				if m.selectedTab == 0 {
					m.logTable, cmd = m.logTable.Update(msg)
					return m, cmd
				}
				if m.selectedTab == 1 {
					m.rrTable, cmd = m.rrTable.Update(msg)
					return m, cmd
				}
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			switch m.focusLayer {
			case focusTabs:
			case focusButtons:
				if m.selectedTab == 0 && m.tabs[0].cursor == 0 {
					m.logTable.Focus()
				}
				if m.selectedTab == 1 && m.tabs[0].cursor == 0 {
					m.rrTable.Focus()
				}
			case focusTable:
			}
		default:
			fmt.Println("uknown key")
		}
	default:
		if m.selectedTab == 0 {
			m.logTable, cmd = m.logTable.Update(msg)
			return m, cmd
		}
		if m.selectedTab == 1 {
			m.rrTable, cmd = m.rrTable.Update(msg)
			return m, cmd
		}
	}

	select {
	case log := <-m.msg:
		m, cmd := m.Update(logMsg(log))
		return m, cmd
	default:
		return m, nil
	}
}

func (m *model) getTable() *table.Model {
	var table *table.Model
	if m.selectedTab == 0 {
		table = &m.logTable
	}
	if m.selectedTab == 1 {
		table = &m.rrTable
	}
	return table
}

func (m model) updateTable(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	if m.selectedTab == 0 {
		m.logTable, cmd = m.logTable.Update(msg)
	}
	if m.selectedTab == 1 {
		m.rrTable, cmd = m.rrTable.Update(msg)
	}
	return m, cmd
}

func (m model) View() string {
	raw := m.rawView() // your existing UI rendering logic
	return baseBorderStyle.Render(raw)
}

func (m model) rawView() string {
	s := strings.Builder{}

	var table *table.Model
	switch m.selectedTab {
	case 0:
		table = &m.logTable
	case 1:
		table = &m.rrTable
	}

	// Header with border
	s.WriteString(headerStyle.Render("DNS Server Dashboard"))
	s.WriteString("\n\n")

	// Tabs row
	tabs := make([]string, len(m.tabs))
	for i, tab := range m.tabs {
		if m.selectedTab == i {
			if m.focusLayer == focusTabs {
				tabs[i] = selectedButtonStyle.Render(tab.name)
			} else {
				tabs[i] = secondarySelectedButtonStyle.Render(tab.name)
			}
		} else {
			tabs[i] = buttonStyle.Render(tab.name)
		}
	}
	buttonsAlignCenter := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center)
	s.WriteString(buttonsAlignCenter.Render(lipgloss.JoinHorizontal(lipgloss.Top, tabs...)))
	s.WriteString("\n\n")

	// Content area with border
	buttons := make([]string, len(m.tabs[m.selectedTab].buttons))
	for i, btn := range m.tabs[m.selectedTab].buttons {
		if m.tabs[m.selectedTab].cursor == i {
			if m.focusLayer == focusButtons {
				buttons[i] = selectedButtonStyle.Render(btn)
			} else {
				buttons[i] = secondarySelectedButtonStyle.Render(btn)
			}
		} else {
			buttons[i] = buttonStyle.Render(btn)
		}
	}

	// Second button row
	buttonsAlignCenter = lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center)
	s.WriteString(buttonsAlignCenter.Render(lipgloss.JoinHorizontal(lipgloss.Top, buttons...)))
	s.WriteString("\n\n")

	if table.Focused() {
		s.WriteString(selectedBoarderStyle.Render(table.View()))
	} else {
		s.WriteString(unselectedBoarderStyle.Render(table.View()))
	}

	s.WriteString("\n\n")

	// Footer with border
	s.WriteString(footerStyle.Render("Press q to quit. Use ↑/↓ to navigate."))

	return s.String()
}
