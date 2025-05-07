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
	"golang.org/x/term"
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

	msgPopup
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

		msgPopup: msgPopup{
			msgChan: make(chan string),
		},
	}
}

func (m model) Close() {
	close(m.msg)
	m.sockConn.Close()
}

func (m model) Init() tea.Cmd {
	go m.readSocket()
	return tea.Batch(waitForLogMsg(m.msg),
		listenForPopupMsg(m.msgPopup.msgChan))
}
