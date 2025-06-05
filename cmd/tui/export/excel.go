package export

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	excelize "github.com/xuri/excelize/v2"
)

// NewExcelFile creates an Excel file with the given rows and charts based on the table type.
// For logTable, it includes a pie chart for Level counts and a column chart for messages by day.
// For rrTable, it includes pie charts for Type and Class counts.
// Returns the path to the saved Excel file or an error.
func NewExcelFile(rows []table.Row, isLogTable bool) (string, error) {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	// Create Data sheet
	dataSheet := "Data"
	idx, err := f.NewSheet(dataSheet)
	if err != nil {
		return "", err
	}
	f.SetActiveSheet(idx)

	// Write headers
	headers := []string{"ID", "Domain", "Data", "Type", "Class", "TTL"}
	if isLogTable {
		headers = []string{"Time", "Level", "Message"}
	}
	for col, header := range headers {
		cell, err := excelize.CoordinatesToCellName(col+1, 1)
		if err != nil {
			return "", fmt.Errorf("failed to set header: %w", err)
		}
		if err := f.SetCellValue(dataSheet, cell, header); err != nil {
			return "", fmt.Errorf("failed to set header value: %w", err)
		}
	}

	// Write rows
	for rowIdx, row := range rows {
		for colIdx, value := range row {
			cell, err := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			if err != nil {
				return "", fmt.Errorf("failed to set cell: %w", err)
			}
			if err := f.SetCellValue(dataSheet, cell, value); err != nil {
				return "", fmt.Errorf("failed to set cell value: %w", err)
			}
		}
	}

	// Create Charts sheet
	chartSheet := "Charts"
	f.NewSheet(chartSheet)

	if isLogTable {
		// Log Table Charts
		// Pie Chart: Count by Level
		levelCounts := map[string]int{"ERROR": 0, "INFO": 0, "WARN": 0}
		for _, row := range rows {
			if len(row) > 1 {
				level := strings.ToUpper(row[1])
				if _, ok := levelCounts[level]; ok {
					levelCounts[level]++
				}
			}
		}

		// Write level counts to Charts sheet
		f.SetCellValue(chartSheet, "A1", "Level")
		f.SetCellValue(chartSheet, "B1", "Count")
		f.SetCellValue(chartSheet, "A2", "ERROR")
		f.SetCellValue(chartSheet, "B2", levelCounts["ERROR"])
		f.SetCellValue(chartSheet, "A3", "INFO")
		f.SetCellValue(chartSheet, "B3", levelCounts["INFO"])
		f.SetCellValue(chartSheet, "A4", "WARN")
		f.SetCellValue(chartSheet, "B4", levelCounts["WARN"])

		pieChart := &excelize.Chart{
			Type: excelize.Pie,
			Series: []excelize.ChartSeries{
				{
					Name:       "Level Counts",
					Categories: fmt.Sprintf("%s!A2:A4", chartSheet),
					Values:     fmt.Sprintf("%s!B2:B4", chartSheet),
				},
			},
			Format: excelize.GraphicOptions{
				OffsetX: 10,
				OffsetY: 10,
			},
			Title: []excelize.RichTextRun{
				{Text: "Messages by Level"},
			},
			PlotArea: excelize.ChartPlotArea{
				ShowPercent: true,
			},
		}
		if err := f.AddChart(chartSheet, "A5", pieChart); err != nil {
			return "", fmt.Errorf("failed to add pie chart: %w", err)
		}

		// Column Chart: Messages by Day
		dayCounts := make(map[string]int)
		for _, row := range rows {
			if len(row) > 0 {
				t, err := time.Parse(time.DateTime, row[0])
				if err == nil {
					day := t.Format("2006-01-02")
					dayCounts[day]++
				}
			}
		}

		// Sort days
		var days []string
		for day := range dayCounts {
			days = append(days, day)
		}
		sort.Strings(days)

		// Write day counts to Charts sheet
		f.SetCellValue(chartSheet, "D1", "Day")
		f.SetCellValue(chartSheet, "E1", "Count")
		for i, day := range days {
			f.SetCellValue(chartSheet, fmt.Sprintf("D%d", i+2), day)
			f.SetCellValue(chartSheet, fmt.Sprintf("E%d", i+2), dayCounts[day])
		}

		columnChart := &excelize.Chart{
			Type: excelize.Col,
			Series: []excelize.ChartSeries{
				{
					Name:       "Messages by Day",
					Categories: fmt.Sprintf("%s!D2:D%d", chartSheet, len(days)+1),
					Values:     fmt.Sprintf("%s!E2:E%d", chartSheet, len(days)+1),
				},
			},
			Format: excelize.GraphicOptions{
				OffsetX: 350,
				OffsetY: 10,
			},
			Title: []excelize.RichTextRun{
				{Text: "Messages by Day"},
			},
			PlotArea: excelize.ChartPlotArea{
				ShowVal: true,
			},
		}
		if err := f.AddChart(chartSheet, "I5", columnChart); err != nil {
			return "", fmt.Errorf("failed to add column chart: %w", err)
		}
	} else {
		// Record Table Charts
		// Pie Chart: Count by Type
		typeCounts := make(map[string]int)
		for _, row := range rows {
			if len(row) > 3 {
				typ := strings.ToUpper(row[3])
				typeCounts[typ]++
			}
		}

		// Write type counts to Charts sheet
		f.SetCellValue(chartSheet, "A1", "Type")
		f.SetCellValue(chartSheet, "B1", "Count")
		rowIdx := 2
		var types []string
		for typ := range typeCounts {
			types = append(types, typ)
		}
		sort.Strings(types)
		for _, typ := range types {
			f.SetCellValue(chartSheet, fmt.Sprintf("A%d", rowIdx), typ)
			f.SetCellValue(chartSheet, fmt.Sprintf("B%d", rowIdx), typeCounts[typ])
			rowIdx++
		}

		pieChartType := &excelize.Chart{
			Type: excelize.Pie,
			Series: []excelize.ChartSeries{
				{
					Name:       "Type Counts",
					Categories: fmt.Sprintf("%s!A2:A%d", chartSheet, rowIdx-1),
					Values:     fmt.Sprintf("%s!B2:B%d", chartSheet, rowIdx-1),
				},
			},
			Format: excelize.GraphicOptions{
				OffsetX: 10,
				OffsetY: 10,
			},
			Title: []excelize.RichTextRun{
				{Text: "Records by Type"},
			},
			PlotArea: excelize.ChartPlotArea{
				ShowPercent: true,
			},
		}
		if err := f.AddChart(chartSheet, "A5", pieChartType); err != nil {
			return "", fmt.Errorf("failed to add type pie chart: %w", err)
		}

		// Pie Chart: Count by Class
		classCounts := make(map[string]int)
		for _, row := range rows {
			if len(row) > 4 {
				cls := strings.ToUpper(row[4])
				classCounts[cls]++
			}
		}

		// Write class counts to Charts sheet
		f.SetCellValue(chartSheet, "D1", "Class")
		f.SetCellValue(chartSheet, "E1", "Count")
		rowIdx = 2
		var classes []string
		for cls := range classCounts {
			classes = append(classes, cls)
		}
		sort.Strings(classes)
		for _, cls := range classes {
			f.SetCellValue(chartSheet, fmt.Sprintf("D%d", rowIdx), cls)
			f.SetCellValue(chartSheet, fmt.Sprintf("E%d", rowIdx), classCounts[cls])
			rowIdx++
		}

		pieChartClass := &excelize.Chart{
			Type: excelize.Pie,
			Series: []excelize.ChartSeries{
				{
					Name:       "Class Counts",
					Categories: fmt.Sprintf("%s!D2:D%d", chartSheet, rowIdx-1),
					Values:     fmt.Sprintf("%s!E2:E%d", chartSheet, rowIdx-1),
				},
			},
			Format: excelize.GraphicOptions{
				OffsetX: 350,
				OffsetY: 10,
			},
			Title: []excelize.RichTextRun{
				{Text: "Records by Class"},
			},
			PlotArea: excelize.ChartPlotArea{
				ShowPercent: true,
			},
		}
		if err := f.AddChart(chartSheet, "I5", pieChartClass); err != nil {
			return "", fmt.Errorf("failed to add class pie chart: %w", err)
		}
	}

	// Delete default Sheet1
	f.DeleteSheet("Sheet1")

	// Save file to temporary directory
	tempDir := os.TempDir()
	fileName := fmt.Sprintf("export_%s.xlsx", time.Now().Format("20060102150405"))
	filePath := filepath.Join(tempDir, fileName)
	if err := f.SaveAs(filePath); err != nil {
		return "", fmt.Errorf("failed to save Excel file: %w", err)
	}

	return filePath, nil
}
