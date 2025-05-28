package export

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/gomutex/godocx"
)

// NewWordFile creates a Word document with a report of logs and DNS records between startDate and endDate.
// The report includes tables for filtered log and record entries, and summaries of counts by Level (logs)
// and Type/Class (records). Returns the path to the saved Word file or an error.
func NewWordFile(logRows, rrRows []table.Row, startDate, endDate time.Time) (string, error) {
	// Validate dates
	if startDate.After(endDate) {
		return "", fmt.Errorf("startDate must be before or equal to endDate")
	}

	// Create new document
	doc, err := godocx.NewDocument()
	if err != nil {
		return "", fmt.Errorf("failed to create document: %w", err)
	}
	defer doc.Close()

	// Add title
	para, err := doc.AddHeading("DNS Server Report", 1)
	if err != nil {
		return "", fmt.Errorf("failed to add title: %w", err)
	}
	para.AddRun().Bold(true)

	// Add introduction
	intro := fmt.Sprintf(
		"This report summarizes DNS server logs and records from %s to %s.",
		startDate.Format("2006-01-02"), endDate.Format("2006-01-02"),
	)
	para = doc.AddParagraph(intro)
	para.AddRun()

	// Logs Section
	para, err = doc.AddHeading("Logs", 2)
	if err != nil {
		return "", fmt.Errorf("failed to add logs heading: %w", err)
	}
	para.AddRun().Bold(true)

	// Filter log rows by date
	var filteredLogRows []table.Row
	for _, row := range logRows {
		if len(row) < 1 {
			continue
		}
		t, err := time.Parse(time.DateTime, row[0])
		if err != nil {
			continue // Skip invalid dates
		}
		if !t.Before(startDate) && !t.After(endDate) {
			filteredLogRows = append(filteredLogRows, row)
		}
	}

	// Create logs table
	logTable := doc.AddTable()
	headerRow := logTable.AddRow()
	for _, header := range []string{"Time", "Level", "Message"} {
		headerCell := headerRow.AddCell()
		para := headerCell.AddParagraph(header)
		para.AddRun().Bold(true)
	}

	for _, row := range filteredLogRows {
		if len(row) < 3 {
			continue
		}
		dataRow := logTable.AddRow()
		for _, value := range row[:3] {
			cell := dataRow.AddCell()
			para := cell.AddParagraph(value)
			para.AddRun()
		}
	}

	// Logs summary
	levelCounts := map[string]int{"ERROR": 0, "INFO": 0, "WARN": 0}
	for _, row := range filteredLogRows {
		if len(row) < 2 {
			continue
		}
		level := strings.ToUpper(row[1])
		if _, ok := levelCounts[level]; ok {
			levelCounts[level]++
		}
	}
	summary := fmt.Sprintf(
		"Log Summary: %d total logs. %d ERROR, %d INFO, %d WARN.",
		len(filteredLogRows), levelCounts["ERROR"], levelCounts["INFO"], levelCounts["WARN"],
	)
	para = doc.AddParagraph(summary)
	para.AddRun()

	para = doc.AddParagraph("\n")

	// Records Section
	para, err = doc.AddHeading("DNS Records", 2)
	if err != nil {
		return "", fmt.Errorf("failed to add records heading: %w", err)
	}
	para.AddRun().Bold(true)

	// Filter rrRows (assuming all records included since no date filter)
	filteredRRRows := rrRows

	// Create records table
	{
		rrTable := doc.AddTable()
		headerRow := rrTable.AddRow()
		for _, header := range []string{"ID", "Name", "Data", "Class", "Type", "TTL"} {
			headerCell := headerRow.AddCell()
			para := headerCell.AddParagraph(header)
			para.AddRun().Bold(true)
		}

		for _, row := range filteredRRRows {
			if len(row) < 6 {
				continue
			}
			dataRow := rrTable.AddRow()
			for _, value := range row[:6] {
				cell := dataRow.AddCell()
				para := cell.AddParagraph(value)
				para.AddRun()
			}
		}

	}

	// Records summary
	{
		typeCounts := make(map[string]int)
		classCounts := make(map[string]int)
		for _, row := range filteredRRRows {
			if len(row) < 5 {
				continue
			}
			typ := strings.ToUpper(row[3])
			cls := strings.ToUpper(row[4])
			typeCounts[typ]++
			classCounts[cls]++
		}

		var typeSummary, classSummary string
		if len(typeCounts) > 0 {
			var types []string
			for typ, count := range typeCounts {
				{
					types = append(types, fmt.Sprintf("%s: %d", typ, count))
				}
			}
			sort.Strings(types)
			typeSummary = strings.Join(types, ", ")
		} else {
			typeSummary = "None"
		}
		if len(classCounts) > 0 {
			var classes []string
			for cls, count := range classCounts {
				{
					classes = append(classes, fmt.Sprintf("%s: %d", cls, count))
				}
			}
			sort.Strings(classes)
			classSummary = strings.Join(classes, ", ")
		} else {
			classSummary = "None"
		}
		summary := fmt.Sprintf(
			"Records Summary: %d total records. Types: %s. Classes: %s.",
			len(filteredRRRows), typeSummary, classSummary,
		)
		para = doc.AddParagraph(summary)
		para.AddRun()

		// Save file to temporary directory
		tempDir := os.TempDir()
		fileName := fmt.Sprintf("report_%s.docx", time.Now().Format("20060102150405"))
		filePath := filepath.Join(tempDir, fileName)
		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return "", fmt.Errorf("failed to save Word file: %w", err)
		}
		if err := doc.Write(f); err != nil {
			return "", fmt.Errorf("failed to save Word file: %w", err)
		}

		return filePath, nil
	}
}
