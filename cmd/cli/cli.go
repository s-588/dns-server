// Package cli provides a command-line interface for interacting with the DNS server.
// It communicates with the server's HTTP API to perform CRUD operations on DNS records.
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
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/prionis/dns-server/internal/database"
	"github.com/prionis/dns-server/internal/server"
	"github.com/prionis/dns-server/proto/genproto/crudpb"
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

// AddRR takes a DNS record in RFC 1035 format, marshals it into a protobuf message,
// and sends it to the server's HTTP API to add the record to the database.
func AddRR(arg string, addr, port string) {
	rr, err := dns.NewRR(arg)
	if err != nil {
		printError("Add accept DNS records in RFC 1035 format(ex. \"example.com. 3600 IN A 127.0.0.1\" )")
		return
	}
	s := strings.Split(rr.String(), " ")

	body, err := proto.Marshal(&crudpb.ResourceRecord{
		Domain: s[0],
		Ttl:    rr.Header().Ttl,
		Class:  s[2],
		Type:   s[3],
		Data:   s[4],
	})
	if err != nil {
		printError("can't marshal record: " + err.Error())
		return
	}

	req, err := http.NewRequest(http.MethodPost, addr+port+"/api/rr", bytes.NewReader(body))
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

// DelRR sends a delete request to the server's HTTP API to remove a DNS record by its ID.
func DelRR(id int64, addr, port string) {
	req, err := http.NewRequest(http.MethodDelete, addr+port+"/api/rr/"+strconv.FormatInt(id, 10), http.NoBody)
	if err != nil {
		printError("Can't create request. Error: " + err.Error())
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		printError("Can't make request to the server by address: " + addr + port + ". Error: " + err.Error())
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

// StartServer initializes and starts the DNS and HTTP servers.
func StartServer(logPath string, opts ...server.Option) {
	server.LoadEnvs()
	logFile, err := os.OpenFile(filepath.Clean(logPath), os.O_RDWR|os.O_CREATE, 0o600)
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

	config := append(server.OptionsFromEnv(), opts...)
	config = append(config, server.WithDB(db))
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

// PrintLogList reads a log file unmarshaling each line into a log struct,
// and prints the formatted log entries to stdout.
func PrintLogList(logPath string) {
	file, err := os.Open(filepath.Clean(logPath))
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
			_, err = fmt.Fprintf(os.Stdout, "%s %s %s\n", log.Time.Format(time.DateTime), log.Level, log.Msg)
			if err != nil {
				printError("can't write to stdout: " + err.Error())
			}
		}
		wg.Done()
	}(logChan)

	wg.Wait()
}

// PrintRRList sends a request to the server's HTTP API to retrieve all DNS records, and prints to stdout.
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
	err = proto.Unmarshal(body, rrs)
	if err != nil {
		printError("can't unmarshal response body: " + err.Error())
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

// CheckArgs checks the provided arguments to ensure that only one main action flag is used at a time.
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
