package main

import (
	"flag"

	"github.com/prionis/dns-server/internal/cli"
)

func main() {
	flagServer := flag.Bool("server", false, "start DNS server")
	flagListLog := flag.Bool("list-log", false, "show all logs in list view")
	flagListRR := flag.Bool("list-rr", false, "show all resource records in list view")
	flagAddRR := flag.String("add", "", "add new resource record. Accept 5 parameters (type,class,domain,data,TimeToLive). First 4 is necessary")
	flagDelRR := flag.Int64("del", -1, "delete resource record. Accept ID of resource record to delete")

	flag.Parse()
	cli.CheckArgs(flagServer, flagListLog, flagListRR, flagAddRR, flagDelRR)

	switch {
	case *flagAddRR != "":
		// ui.AddRR(*flagAddRR)
	case *flagDelRR != -1:
		// ui.DelRR(*flagDelRR)
	case *flagServer:
		cli.StartServer()
	case *flagListLog:
		cli.PrintLogList()
	case *flagListRR:
		// ui.PrintRRList()
	default:
		flag.Usage()
	}
}
