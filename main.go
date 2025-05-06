package main

import (
	"flag"

	"github.com/prionis/dns-server/ui"
)

func main() {
	flagServer := flag.Bool("server", false, "start DNS server")
	flagTUI := flag.Bool("tui", false, "show simple text interface")
	flagListLog := flag.Bool("list-log", false, "show all logs in list view")
	flagListRR := flag.Bool("list-rr", false, "show all resource records in list view")
	flagAddRR := flag.String("add", "", "add new resource record. Accept 5 parameters (type,class,domain,data,TimeToLive). First 4 is necessary")
	flagDelRR := flag.Int64("del", -1, "delete resource record. Accept ID of resource record to delete")

	flag.Parse()
	ui.CheckArgs(flagServer, flagTUI, flagListLog, flagListRR, flagAddRR, flagDelRR)

	switch {
	case *flagAddRR != "":
		ui.AddRR(*flagAddRR)
	case *flagDelRR != -1:
		ui.DelRR(*flagDelRR)
	case *flagServer:
		ui.StartServer()
	case *flagTUI:
		ui.StartTUI()
	case *flagListLog:
		ui.PrintLogList()
	case *flagListRR:
		ui.PrintRRList()
	default:
		flag.Usage()
	}
}
