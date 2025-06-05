package tui

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	crud "github.com/prionis/dns-server/cmd/tui/CRUD"
	"github.com/prionis/dns-server/cmd/tui/auth"
	"github.com/prionis/dns-server/cmd/tui/popup"
	"github.com/prionis/dns-server/cmd/tui/transport"
	"github.com/prionis/dns-server/proto/crud/genproto/crudpb"
	"golang.org/x/term"
)

const (
	// Focus layers represent page user on
	focusTabs = iota
	focusButtons
	focusTable
	focusAddPage
	focusDeletePage
	focusUpdatePage
	focusFilterPage
	focusSortPage
	focusLoginPage
	focusRegister

	// Minimum width and height of the screen to fit atleast one row in the tables
	minWidth  = 52
	minHeight = 22
)

// This is the main model of the user interface.
// It render all other buttons, tables and other models.
type model struct {
	// Width of the screen
	width int
	// Height of the screen
	height int

	user *crudpb.User

	// What element user use right now
	focusLayer int
	// Tabs for select table
	tabs        []tab
	selectedTab int

	loginPage auth.LoginModel

	// Table that contain logs of the server.
	logTable table.Model
	// Table that contain resource records from the database.
	rrTable table.Model

	// Model for popup notifications.
	popup popup.PopupModel

	// Model for deleting resource records from the database.
	deleteModel crud.DeleteModel
	// Model for adding resource records to the database.
	addModel crud.AddModel
	// Model for updating resource records of the database.
	updatePage crud.UpdateModel
	// Model for filtering rows of the table.
	filterPage crud.FilterModel
	// Model for sorting rows of the table.
	sortPage crud.SortModel

	// Pointer to the dabase connection.
	transport *transport.Transport
}

// Struct for representing tab on the top of the screen.
type tab struct {
	// Name of the tab.
	name string
	// Buttons for manipulation with tab data.
	buttons []string
	// Selected button.
	cursor int
}

// Creating new model of the user interface.
func NewModel() (model, error) {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		slog.Error("can't get term size")
	}
	if w < minWidth || h < minHeight {
		return model{},
			fmt.Errorf("Minimum size of the screen is %dx%d. Current is %dx%d",
				minWidth, minHeight, w, h)
	}

	t, err := transport.New("172.17.0.1:8083")
	if err != nil {
		return model{}, fmt.Errorf("can't create http transport: %w", err)
	}

	rrTable, logTable := rrTable(t, w, h), logTable(w, h)
	return model{
		loginPage:  auth.NewLoginModel(t, w, h),
		focusLayer: focusLoginPage,

		width:  w,
		height: h,
		tabs: []tab{
			{
				name: "Logs",
				buttons: []string{
					fmt.Sprintf("View%c ", '\uebb7'),
					fmt.Sprintf("Filter%c ", '\ueaf1'),
					fmt.Sprintf("Sort%c ", '\ueaf1'),
					fmt.Sprintf("Export to Word%c ", '\ue6a5'),
					fmt.Sprintf("Export to Excel%c ", '\uf1c3'),
				},
			},
			{
				name: "Records",
				buttons: []string{
					fmt.Sprintf("View%c ", '\uebb7'),
					fmt.Sprintf("Add%c ", '\uea60'),
					fmt.Sprintf("Delete%c ", '\uf00d'),
					fmt.Sprintf("Update%c ", '\uea73'),
					fmt.Sprintf("Filter%c ", '\ueaf1'),
					fmt.Sprintf("Sort%c ", '\ueaf1'),
					fmt.Sprintf("Export to Word%c ", '\ue6a5'),
					fmt.Sprintf("Export to Excel%c ", '\uf1c3'),
				},
			},
		},
		transport: t,

		rrTable:  rrTable,
		logTable: logTable,

		popup:       popup.NewPopupModel(),
		deleteModel: crud.NewDeleteModel(nil, t, w, h),
		addModel:    crud.NewAddModel(t, w, h),
		filterPage:  crud.NewFilterModel(nil, nil, w, h),
	}, nil
}

// Create new table for the resource records and fill it with data from database.
func rrTable(t *transport.Transport, w, h int) table.Model {
	dbRRs, err := t.GetAllRRs()
	if err != nil {
		slog.Error("can't get records from database", "error", err)
	}

	cols := []table.Column{
		{
			Title: "ID",
			Width: max(4, w/10-5),
		},
		{
			Title: "Domain",
			Width: max(8, w/10*3-5),
		},
		{
			Title: "Data",
			Width: max(10, w/10*3-5),
		},
		{
			Title: "Type",
			Width: max(5, w/10-5),
		},
		{
			Title: "Class",
			Width: max(2, w/10-5),
		},
		{
			Title: "TimeToLive",
			Width: max(6, w/10-5),
		},
	}
	rows := make([]table.Row, 0, len(dbRRs))
	for _, rr := range dbRRs {
		rows = append(rows, table.Row{
			strconv.FormatInt(rr.ID, 10),
			rr.Domain,
			rr.Data,
			rr.Type,
			rr.Class,
			strconv.FormatInt(int64(rr.TTL), 10),
		})
	}
	return table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithHeight(max(3, h-20)),
		table.WithWidth(w-8))
}

// Create new table for server logs and fill it with existing logs.
func logTable(w, h int) table.Model {
	rows := make([]table.Row, 0)
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
	cols := []table.Column{
		{
			Title: "Time",
			Width: max(8, w/5-5),
		},
		{
			Title: "Level",
			Width: max(5, w/5-5),
		},
		{
			Title: "Message",
			Width: max(8, (w/5)*3-5),
		},
	}
	return table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithHeight(max(3, h-20)),
		table.WithWidth(w-8),
	)
}

// Exit point of the app.
func (m model) Close() tea.Cmd {
	return tea.Quit
}

// Initialize esential things.
func (m model) Init() tea.Cmd {
	return tea.Batch(popup.ListenForPopupMsg(m.popup.MsgChan))
}
