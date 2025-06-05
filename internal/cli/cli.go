package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/prionis/dns-server/cmd/tui"
	"github.com/prionis/dns-server/internal/database"
	"github.com/prionis/dns-server/internal/server"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	errorHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#e78284")).
				Bold(true)

	successHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#a6d189")).
				Bold(true)

	textStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#c6d0f5"))

	messageStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Width(50).
			Align(lipgloss.Center).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#ca9ee6"))
)

// func AddRR(arg string) {
// 	// Multi-arg validation for --add and --del
// 	args := strings.Split(arg, ",")
// 	if len(args) < 4 {
// 		printError("--add requires 4 comma-separated arguments (type,class,domain,data,TimeToLive)")
// 		return
// 	}
// 	switch args[0]{
// 	case dns.A:
//
// 	}
// 	rrType := args[0]
// 	rrClass := args[1]
// 	rrDomain := args[2]
// 	rrData := args[3]
// 	rrTTL := int64(0)
// 	if len(args) == 5 {
// 		ttl, err := strconv.ParseInt(args[4], 10, 64)
// 		if err != nil {
// 			printError("can't parse TimeToLive parameter\n" + err.Error())
// 			return
// 		}
// 		rrTTL = ttl
// 	}
// 	db, err := database.NewPostgres()
// 	if err != nil {
// 		printError("can't connect to database\n" + err.Error())
// 		return
// 	}
// 	db.AddRR(database.ResourceRecord{
// 		Class: rrClass,
// 		Domain: rrDomain,
// 		Data: string(rrData),
//
// 	}, rrClass, rrDomain, rrData, int64(rrTTL))
// 	printSuccess(fmt.Sprintf("new resource record was added\n%s %s\nClass: %s\nType: %s\nTimeToLive: %d", rrDomain, rrData.String(), rrClass, rrType, rrTTL))
// }

// func DelRR(arg int64) {
// 	db, err := sqlite.NewDB()
// 	if err != nil {
// 		printError("can't connect to database\n" + err.Error())
// 		return
// 	}
// 	rr, err := db.DelRR(arg)
// 	if err != nil {
// 		printError("can't retrive resource record from database\n" + err.Error())
// 		return
// 	}
// 	printSuccess(fmt.Sprintf("record was deleted\n%s %s\nClass: %s\nType: %s\nTimeToLive: %d",
// 		rr.RR.Domain, net.IP(rr.RR.Data).String(),
// 		protocol.MapKeyByValue(protocol.Classes, rr.RR.Class),
// 		protocol.MapKeyByValue(protocol.Types, rr.RR.Type), rr.RR.TimeToLive))
// }

func StartServer() {
	// sock, err := tui.NewUISocket()
	// if err != nil {
	// 	printError("can't create socket\n" + err.Error())
	// 	return
	// }
	//
	logFile, err := os.OpenFile("DNSServer.log", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		printError("can't open log file\n" + err.Error())
		return
	}

	writer := io.MultiWriter(logFile, os.Stdout)
	logger := slog.New(slog.NewJSONHandler(writer, nil))
	slog.SetDefault(logger)

	db, err := database.NewPostgres("")
	if err != nil {
		fmt.Println("can't connect to database\n" + err.Error())
		return
	}
	logger.Info("connection with database established")

	config := []server.Option{
		server.SetDNSPort(":1053"),
		server.WithDB(db),
	}
	s, err := server.NewServer(config...)
	if err != nil {
		printError(fmt.Sprintf("can't create new server\n%s", err.Error()))
		return
	}
	logger.Info("server created")

	logger.Info("starting server")
	if err := s.Start(); err != nil {
		printError(fmt.Sprintf("can't start server\n%s", err.Error()))
	}
}

func StartTUI() {
	m, err := tui.NewModel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't start. %s", err.Error())
		return
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		printError("can't open TUI\n" + err.Error())
	}
}

func PrintLogList() {
	file, err := os.Open("DNSServer.log")
	if err != nil {
		printError("can't open log file\n" + err.Error())
		return
	}
	scanner := bufio.NewScanner(file)
	s := make([]table.Row, 0)
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
		s = append(s, table.Row{t.Format(time.DateTime), log["level"].(string), log["msg"].(string)})
	}
	columns := []table.Column{{"time", 20}, {"level", 5}, {"message", 20}}
	t := table.New(table.WithColumns(columns), table.WithRows(s))
	style := table.DefaultStyles()
	style.Selected = style.Cell
	t.SetStyles(style)
	fmt.Fprintln(os.Stdout, t.View())
}

// func PrintRRList() {
// 	db, err := sqlite.NewDB()
// 	if err != nil {
// 		printError("can't connect to database/n" + err.Error())
// 	}
// 	rrs, err := db.GetAllRRs()
// 	if err != nil {
// 		printError("can't get resource records from database\n" + err.Error())
// 		return
// 	}
// 	rows := make([]table.Row, 0)
// 	for _, rr := range rrs {
// 		rows = append(rows, table.Row{
// 			strconv.FormatInt(rr.ID, 10),
// 			protocol.MapKeyByValue(protocol.Types, rr.RR.Type),
// 			protocol.MapKeyByValue(protocol.Classes, rr.RR.Class),
// 			rr.RR.Domain,
// 			net.IP(rr.RR.Data).String(),
// 			strconv.FormatInt(int64(rr.RR.TimeToLive), 10),
// 		})
// 	}
// 	cols := []table.Column{{"ID", 5}, {"Type", 4}, {"Class", 5}, {"Domain", 15}, {"Data", 20}, {"TimeToLive", 14}}
// 	t := table.New(table.WithColumns(cols), table.WithRows(rows))
// 	style := table.DefaultStyles()
// 	style.Selected = style.Cell
// 	t.SetStyles(style)
// 	fmt.Println(t.View())
// }

func CheckArgs(args ...any) {
	count := 0
	for _, arg := range args {
		switch v := arg.(type) {
		case *string:
			if *v != "" {
				count++
			}
		case *bool:
			if *v {
				count++
			}
		case *int64:
			if *v != -1 {
				count++
			}
		}
	}
	if count > 1 {
		printError("Only one main action flag can be used at a time.")
		os.Exit(1)
	}
}

func printError(err string) {
	// Define a purple-magenta palette similar to Bubbles style
	s := errorHeaderStyle.Render("ERROR")
	s += "\n\n"
	s += textStyle.Render(err)
	fmt.Fprint(os.Stderr, messageStyle.Render(s))
}

func printSuccess(msg string) {
	// Define a purple-magenta palette similar to Bubbles style
	s := successHeaderStyle.Render("SUCCESS")
	s += "\n\n"
	s += textStyle.Render(msg)
	fmt.Fprint(os.Stderr, messageStyle.Render(s))
}
