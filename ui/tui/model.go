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

func NewModel() model {
	conn, err := MakeSockConn()
	if err != nil {
		slog.Error("can't connect to socket", "error", err)
	}

	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		slog.Error("can't get term size")
	}
	w -= 7

	db, err := sqlite.NewDB()
	if err != nil {
		slog.Error("can't connect to database", "error", err)
	}

	dbRRs, err := db.GetAllRRs()
	if err != nil {
		slog.Error("can't get records from database", "error", err)
	}

	cols := []table.Column{{"ID", 4}, {"domain", max(20, w/6)}, {"data", max(20, w/6)}, {"type", 6}, {"class", 8}, {"TimeToLive", 12}}
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
		width:  w,
		height: h,
		tabs: []tab{
			{
				name:    "Logs",
				buttons: []string{"View", "Filter", "Sort"},
			},
			{
				name:    "Records",
				buttons: []string{"View", "Add", "Delete", "Filter", "Sort"},
			},
		},
		logMsgChan: make(chan map[string]any, 1),
		sockConn:   conn,

		rrTable:  rrTable,
		logTable: logTable,

		popup:      popup.NewPopupModel(),
		deletePage: crud.NewDeleteModel(nil, &db),
		addPage:    crud.NewAddModel(&db),

		db: &db,
	}
}

func (m model) Close() {
	m.sockConn.Close()
}

func (m model) Init() tea.Cmd {
	go m.readSocket()
	return tea.Batch(waitForLogMsg(m.logMsgChan), popup.ListenForPopupMsg(m.popup.MsgChan))
}
