// Package export provides functionality to export data to Excel format using a Python script.
package export

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"

	"github.com/charmbracelet/bubbles/table"
)

// TableData represents the data for a single table.
type TableData struct {
	Name string      `json:"name"`
	Rows []table.Row `json:"rows"`
}

// Excel takes a slice of TableData, marshals it to JSON,
// and passes it to a Python script that generates an Excel file.
// It returns the path to the generated Excel file or an error.
func Excel(tables []TableData) (string, error) {
	data, err := json.Marshal(tables)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("python3", "cmd/tui/util/export/excel.py")
	cmd.Stdin = bytes.NewReader(data)
	var stdOut, stdErr bytes.Buffer
	cmd.Stderr = &stdErr
	cmd.Stdout = &stdOut

	if err = cmd.Run(); err != nil {
		return "", fmt.Errorf("%w: %s", err, stdErr.String())
	}

	path := stdOut.String()
	if path == "" {
		return "", errors.New("no path was returned")
	}
	return path, nil
}
