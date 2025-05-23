package tui

import (
	"bufio"
	"encoding/json"
	"log/slog"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/prionis/dns-server/protocol"
	"github.com/prionis/dns-server/sqlite"
	crud "github.com/prionis/dns-server/ui/tui/CRUD"
	"github.com/prionis/dns-server/ui/tui/popup"
	"golang.org/x/term"
)

const (
	focusTabs = iota
	focusButtons
	focusTable
	focusAddPage
	focusDeletePage

	minWidth  = 52
	minHeight = 22
)

type model struct {
	width  int
	height int

	focusLayer  int
	tabs        []tab
	selectedTab int

	logTable table.Model
	rrTable  table.Model

	popup popup.PopupModel

	deletePage crud.DeleteModel
	addPage    crud.AddModel

	sockConn   net.Conn
	logMsgChan chan map[string]any

	db *sqlite.DB
}

type tab struct {
	name    string
	buttons []string
	cursor  int
}

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

	conn, err := MakeSockConn()
	if err != nil {
		slog.Error("can't connect to socket", "error", err)
	}

	db, err := sqlite.NewDB()
	if err != nil {
		slog.Error("can't connect to database", "error", err)
	}

	rrTable, logTable := rrTable(&db, w, h), logTable(w, h)
	return model{
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
		logMsgChan: make(chan map[string]any, 1),
		sockConn:   conn,

		rrTable:  rrTable,
		logTable: logTable,

		popup:      popup.NewPopupModel(),
		deletePage: crud.NewDeleteModel(nil, &db, w, h),
		addPage:    crud.NewAddModel(&db, w, h),
		db: &db,
	}, nil
}

func rrTable(db *sqlite.DB, w, h int) table.Model {
	dbRRs, err := db.GetAllRRs()
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
			rr.RR.Domain,
			net.IP(rr.RR.Data).String(),
			protocol.MapKeyByValue(protocol.Types, rr.RR.Type),
			protocol.MapKeyByValue(protocol.Classes, rr.RR.Class),
			strconv.FormatInt(int64(rr.RR.TimeToLive), 10),
		})
	}
	return table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithHeight(max(3, h-20)),
		table.WithWidth(w-8))
}

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

func (m model) Close() {
	m.sockConn.Close()
}

func (m model) Init() tea.Cmd {
	go m.readSocket()
	return tea.Batch(waitForLogMsg(m.logMsgChan), popup.ListenForPopupMsg(m.popup.MsgChan))
}
