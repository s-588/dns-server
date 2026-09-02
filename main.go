package main

import (
	"flag"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/prionis/dns-server/cmd/cli"
	"github.com/prionis/dns-server/cmd/tui"
)

func main() {
	flagServer := flag.Bool("server", false, "start DNS server")
	flagTUI := flag.Bool("tui", false, "start TUI")
	flagListLog := flag.Bool("logs", false, "show all logs")
	flagListRR := flag.Bool("records", false, "show all resource records in list view")
	flagAddRR := flag.String("add", "", "add new resource record. Accept 5 parameters (type,class,domain,data,TimeToLive). First 4 is necessary")
	flagAddr := flag.String("addr", "127.0.0.1", "set specific addr of the server. Default is 127.0.0.1")
	flagPort := flag.String("port", ":8080", "set specific port of the server. Default is \":8080\"")
	flagLogPath := flag.String("logfile", "DNSServer.log", "set specific name(or path) of log file")
	flagDelRR := flag.Int64("del", -1, "delete resource record. Accept ID of resource record to delete")

	flag.Parse()

	switch {
	case *flagAddRR != "":
		cli.AddRR(*flagAddRR, *flagAddr, *flagPort)
	case *flagDelRR != -1:
		cli.DelRR(*flagDelRR, *flagAddr, *flagPort)
	case *flagServer:
		cli.StartServer(*flagLogPath)
	case *flagTUI:
		m, err := tui.NewModel()
		if err != nil {
			fmt.Println(err)
			return
		}
		app := tea.NewProgram(m)
		if _, err := app.Run(); err != nil {
			fmt.Println(err)
		}
	case *flagListLog:
		cli.PrintLogList(*flagLogPath)
	case *flagListRR:
		cli.PrintRRList(*flagAddr, *flagPort)
	default:
		flag.PrintDefaults()
	}
}
