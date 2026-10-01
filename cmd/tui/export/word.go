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
	"github.com/gomutex/godocx/docx"
)

// NewWordFile creates a Word document with a report of logs and DNS records between startDate and endDate.
// The report includes tables for filtered log and record entries, and summaries of counts by Level (logs)
// and Type/Class (records). Returns the path to the saved Word file or an error.
func NewWordFile(logRows, rrRows []table.Row, startDate, endDate time.Time) (string, error) {
	if startDate.After(endDate) {
		return "", fmt.Errorf("startDate must be before or equal to endDate")
	}

	doc, err := godocx.NewDocument()
	if err != nil {
		return "", fmt.Errorf("failed to create document: %w", err)
	}
	defer func() {
		if err := doc.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to close document: %v\n", err)
		}
	}()

	if err := addTitleAndIntro(doc, startDate, endDate); err != nil {
		return "", err
	}

	filteredLogs := filterLogRows(logRows, startDate, endDate)
	if err := addLogsSection(doc, filteredLogs); err != nil {
		return "", err
	}

	if err := addRecordsSection(doc, rrRows); err != nil {
		return "", err
	}

	return saveDocument(doc)
}

func addTitleAndIntro(doc *docx.RootDoc, start, end time.Time) error {
	para, err := doc.AddHeading("DNS Server Report", 1)
	if err != nil {
		return fmt.Errorf("failed to add title: %w", err)
	}
	para.AddRun().Bold(true)

	intro := fmt.Sprintf(
		"This report summarizes DNS server logs and records from %s to %s.",
		start.Format("2006-01-02"), end.Format("2006-01-02"),
	)
	doc.AddParagraph(intro)
	return nil
}

func filterLogRows(rows []table.Row, start, end time.Time) []table.Row {
	var out []table.Row
	for _, row := range rows {
		if len(row) < 1 {
			continue
		}
		t, err := time.Parse(time.DateTime, row[0])
		if err != nil {
			continue
		}
		if !t.Before(start) && !t.After(end) {
			out = append(out, row)
		}
	}
	return out
}

func addLogsSection(doc *docx.RootDoc, rows []table.Row) error {
	para, err := doc.AddHeading("Logs", 2)
	if err != nil {
		return fmt.Errorf("failed to add logs heading: %w", err)
	}
	para.AddRun().Bold(true)

	table := doc.AddTable()
	addHeaderRow(table, []string{"Time", "Level", "Message"})

	for _, row := range rows {
		if len(row) < 3 {
			continue
		}
		addDataRow(table, row[:3])
	}

	summary := fmt.Sprintf(
		"Log Summary: %d total logs. %d ERROR, %d INFO, %d WARN.",
		len(rows),
		countLevel(rows, "ERROR"),
		countLevel(rows, "INFO"),
		countLevel(rows, "WARN"),
	)
	doc.AddParagraph(summary)
	doc.AddParagraph("") // blank line
	return nil
}

func addRecordsSection(doc *docx.RootDoc, rows []table.Row) error {
	para, err := doc.AddHeading("DNS Records", 2)
	if err != nil {
		return fmt.Errorf("failed to add records heading: %w", err)
	}
	para.AddRun().Bold(true)

	table := doc.AddTable()
	addHeaderRow(table, []string{"ID", "Name", "Data", "Class", "Type", "TTL"})

	for _, row := range rows {
		if len(row) < 6 {
			continue
		}
		addDataRow(table, row[:6])
	}

	typeCounts, classCounts := countTypesAndClasses(rows)
	summary := fmt.Sprintf(
		"Records Summary: %d total records. Types: %s. Classes: %s.",
		len(rows),
		formatCounts(typeCounts),
		formatCounts(classCounts),
	)
	doc.AddParagraph(summary)
	return nil
}

func addHeaderRow(table *docx.Table, headers []string) {
	row := table.AddRow()
	for _, h := range headers {
		cell := row.AddCell()
		p := cell.AddParagraph(h)
		p.AddRun().Bold(true)
	}
}

func addDataRow(table *docx.Table, values []string) {
	row := table.AddRow()
	for _, v := range values {
		cell := row.AddCell()
		cell.AddParagraph(v)
	}
}

func countLevel(rows []table.Row, level string) int {
	n := 0
	for _, row := range rows {
		if len(row) >= 2 && strings.EqualFold(row[1], level) {
			n++
		}
	}
	return n
}

func countTypesAndClasses(rows []table.Row) (map[string]int, map[string]int) {
	types := make(map[string]int)
	classes := make(map[string]int)
	for _, row := range rows {
		if len(row) < 5 {
			continue
		}
		// Headers: ID, Name, Data, Class, Type, TTL
		// → index 3 = Class, index 4 = Type
		classes[strings.ToUpper(row[3])]++
		types[strings.ToUpper(row[4])]++
	}
	return types, classes
}

func formatCounts(m map[string]int) string {
	if len(m) == 0 {
		return "None"
	}
	parts := make([]string, 0, len(m))
	for k, v := range m {
		parts = append(parts, fmt.Sprintf("%s: %d", k, v))
	}
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}

func saveDocument(doc *docx.RootDoc) (string, error) {
	fileName := fmt.Sprintf("report_%s.docx", time.Now().Format("20060102150405"))
	filePath := filepath.Join(os.TempDir(), fileName)

	f, err := os.OpenFile(filepath.Clean(filePath), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return "", fmt.Errorf("failed to create Word file: %w", err)
	}
	defer func() {
		err := f.Close()
		if err != nil {
			fmt.Printf("failed to close Word file: %v\n", err)
		}
	}()

	if err := doc.Write(f); err != nil {
		return "", fmt.Errorf("failed to write Word file: %w", err)
	}
	return filePath, nil
}
