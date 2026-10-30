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

// defaultLogPath is used when no LOG_FILE environment variable is set.
const defaultLogPath = "DNSServer.log"

// apiBaseURL returns the HTTP(S) base URL for the server API.
// It uses environment variables (loaded via server.LoadEnvs) or sensible defaults.
// Defaults: http://127.0.0.1:8083
func apiBaseURL() string {
	server.LoadEnvs()

	host := strings.TrimSpace(os.Getenv("HTTP_HOST"))
	if host == "" {
		host = "127.0.0.1"
	}

	port := strings.TrimSpace(os.Getenv("HTTP_PORT"))
	if port == "" {
		port = ":8083" // matches NewServer built-in default
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	scheme := "http"
	if v := strings.TrimSpace(os.Getenv("HTTPS")); v != "" {
		enable, err := strconv.ParseBool(v)
		if err != nil {
			lower := strings.ToLower(v)
			enable = lower == "1" || lower == "yes" || lower == "on" || lower == "true"
		}
		if enable {
			scheme = "https"
		}
	}

	return scheme + "://" + host + port
}

// logPath returns the path to the log file (from LOG_FILE env or default).
func logPath() string {
	server.LoadEnvs()
	if p := strings.TrimSpace(os.Getenv("LOG_FILE")); p != "" {
		return p
	}
	return defaultLogPath
}

// AddRR takes a DNS record in RFC 1035 format, marshals it into a protobuf message,
// and sends it to the server's HTTP API to add the record to the database.
func AddRR(arg string) {
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

	req, err := http.NewRequest(http.MethodPost, apiBaseURL()+"/api/rr", bytes.NewReader(body))
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
func DelRR(id int64) {
	req, err := http.NewRequest(http.MethodDelete, apiBaseURL()+"/api/rr/"+strconv.FormatInt(id, 10), http.NoBody)
	if err != nil {
		printError("Can't create request. Error: " + err.Error())
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		printError("Can't make request to the server by address: " + apiBaseURL() + ". Error: " + err.Error())
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
// Configuration comes from environment variables (DNS_PORT, HTTP_PORT, HTTPS, LOG_FILE)
// or built-in defaults.
func StartServer() {
	server.LoadEnvs()
	lp := logPath()
	logFile, err := os.OpenFile(filepath.Clean(lp), os.O_RDWR|os.O_CREATE, 0o600)
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

	config := append(server.OptionsFromEnv(), server.WithDB(db))
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
func PrintLogList() {
	lp := logPath()
	file, err := os.Open(filepath.Clean(lp))
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
func PrintRRList() {
	req, err := http.NewRequest(http.MethodGet, apiBaseURL()+"/api/rr/all", http.NoBody)
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
		printSuccess("Records retrieved")
		// TODO: actually print the records (original code had a bug here)
		for _, rr := range rrs.Records {
			fmt.Printf("%d %s %d %s %s %s\n", rr.Id, rr.Domain, rr.Ttl, rr.Class, rr.Type, rr.Data)
		}
	} else {
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
