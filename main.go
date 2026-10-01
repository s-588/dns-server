package main

import (
	"flag"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/prionis/dns-server/cmd/cli"
	"github.com/prionis/dns-server/cmd/tui"
	"github.com/prionis/dns-server/internal/server"
)

func main() {
	flagServer := flag.Bool("server", false, "start DNS server")
	flagTUI := flag.Bool("tui", false, "start TUI")
	flagListLog := flag.Bool("logs", false, "show all logs")
	flagListRR := flag.Bool("records", false, "show all resource records in list view")
	flagAddRR := flag.String("add", "", "add new resource record. Accept 5 parameters (type,class,domain,data,TimeToLive). First 4 is necessary")
	flagAddr := flag.String("addr", "127.0.0.1", "set specific addr of the server. Default is 127.0.0.1")
	flagHTTPPort := flag.String("port", ":8080", "set specific port of the server. Default is \":8080\"")
	flagDNSPort := flag.String("dns-port", ":53", "set specific port of the DNS server. Default is \":53\"")
	flagLogPath := flag.String("logfile", "DNSServer.log", "set specific name(or path) of log file")
	flagDelRR := flag.Int64("del", -1, "delete resource record. Accept ID of resource record to delete")
	flagHTTPS := flag.Bool("https", false, "enable HTTPS")

	flag.Parse()

	switch {
	case *flagAddRR != "":
		cli.AddRR(*flagAddRR, *flagAddr, *flagHTTPPort)
	case *flagDelRR != -1:
		cli.DelRR(*flagDelRR, *flagAddr, *flagHTTPPort)
	case *flagServer:
		var overrides []server.Option
		if *flagDNSPort != "" {
			p := *flagDNSPort
			if !strings.HasPrefix(p, ":") {
				p = ":" + p
			}
			overrides = append(overrides, server.WithDNSPort(p))
		}
		if *flagHTTPPort != "" {
			p := *flagHTTPPort
			if !strings.HasPrefix(p, ":") {
				p = ":" + p
			}
			overrides = append(overrides, server.WithHTTPPort(p))
		}
		if *flagHTTPS {
			overrides = append(overrides, server.WithHTTPS(true))
		}

		cli.StartServer(*flagLogPath, overrides...)
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
		cli.PrintRRList(*flagAddr, *flagHTTPPort)
	default:
		flag.PrintDefaults()
	}
}
