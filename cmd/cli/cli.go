package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/prionis/dns-server/internal/database"
	"github.com/prionis/dns-server/internal/server"
	"github.com/prionis/dns-server/proto/crud/genproto/crudpb"
	"google.golang.org/protobuf/proto"

	"github.com/charmbracelet/lipgloss"
)

var (
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e78284")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a6d189")).
			Bold(true)
)

func AddRR(arg string, addr, port string) {
	rr, err := dns.NewRR(arg)
	if err != nil {
		printError("Add accept DNS records in RFC 1035 format(ex. \"example.com. 3600 IN A 127.0.0.1\" )")
		return
	}
	s := strings.Split(rr.String(), " ")

	body, err := proto.Marshal(&crudpb.ResourceRecord{
		Domain:     s[0],
		TimeToLive: int32(rr.Header().Ttl),
		Class:      s[2],
		Type:       s[3],
		Data:       s[4],
	})

	req, err := http.NewRequest(http.MethodPost, addr+port+"/api/rr", bytes.NewReader(body))
	req.Header.Add("Content-Type", "application/protobuf")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		printError("can't connect to database\n" + err.Error())
		return
	}

	if resp.StatusCode == http.StatusOK {
		printSuccess("Record was added")
	} else {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			printError("can't read response body from server")
		}
		printError(resp.Status + string(body))
	}
}

func DelRR(id int64, addr, port string) {
	req, err := http.NewRequest(http.MethodDelete, addr+port+"/api/rr/"+strconv.FormatInt(id, 10), http.NoBody)
	if err != nil {
		printError("Can't create request. Error: " + err.Error())
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		printError("Can't make request to the server by addres: " + addr + port + ". Error: " + err.Error())
		return
	}

	if resp.StatusCode == http.StatusOK {
		printSuccess("Record was deleted")
	} else {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			printError("Can't read response body: " + err.Error())
			return
		}
		printError(resp.Status + string(body))
	}
}

func StartServer(logPath string) {
	server.LoadEnvs()
	logFile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		printError("can't open log file\n" + err.Error())
		return
	}

	ws := server.NewWSWriter()
	writer := io.MultiWriter(logFile, os.Stdout, ws)
	logger := slog.New(slog.NewJSONHandler(writer, nil))
	slog.SetDefault(logger)

	slog.Info("connecting to database")
	db, err := database.NewPostgres("")
	if err != nil {
		printError("can't connect to database\n" + err.Error())
		return
	}
	logger.Info("connection with database established")

	config := []server.Option{
		server.SetDNSPort(":53"),
		server.WithDB(db),
	}
	s, err := server.NewServer(config...)
	if err != nil {
		printError(fmt.Sprintf("can't create new server\n%s", err.Error()))
		return
	}
	logger.Info("server created")

	logger.Info("starting server")
	if err := s.Start(ws); err != nil {
		printError(fmt.Sprintf("can't start server\n%s", err.Error()))
	}
}

type log struct {
	Time  time.Time `json:"time"`
	Level string    `json:"level"`
	Msg   string    `json:"msg"`
}

func PrintLogList(logPath string) {
	file, err := os.Open(logPath)
	if err != nil {
		printError("can't open log file\n" + err.Error())
		return
	}
	scanner := bufio.NewScanner(file)
	wg := sync.WaitGroup{}
	logChan := make(chan string, 1)

	wg.Add(2)

	go func(scanner *bufio.Scanner, logChan chan string) {
		for scanner.Scan() {
			line := scanner.Text()
			logChan <- line
		}
		close(logChan)
		wg.Done()
	}(scanner, logChan)

	go func(logChan <-chan string) {
		for line := range logChan {
			log := &log{}
			err := json.Unmarshal([]byte(line), log)
			if err != nil {
				continue
			}
			fmt.Fprintf(os.Stdout, "%s %s %s\n", log.Time.Format(time.DateTime), log.Level, log.Msg)
		}
		wg.Done()
	}(logChan)

	wg.Wait()
}

func PrintRRList(addr, port string) {
	req, err := http.NewRequest(http.MethodGet, "http://"+addr+port+"/api/rr/all", http.NoBody)
	if err != nil {
		printError("can't create request: " + err.Error())
		return
	}
	req.Header.Add("Content-Type", "application/protobuf")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		printError("can't connect to database\n" + err.Error())
		return
	}

	rrs := &crudpb.ResourceRecordCollection{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		printError("can't read response body: " + err.Error())
	}
	proto.Unmarshal(body, rrs)

	if resp.StatusCode == http.StatusOK {
		printSuccess("Record was added")
	} else {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			printError("can't read response body from server")
		}
		printError(resp.Status + string(body))
	}
}

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
	fmt.Fprint(os.Stderr, errorStyle.Render(err))
}

func printSuccess(msg string) {
	fmt.Fprint(os.Stderr, successStyle.Render(msg))
}
