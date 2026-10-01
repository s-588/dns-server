// Package rrTable provides abstract logic for data in TUI tables.
package rrTable

import (
	"errors"
	"fmt"
	"math"
	"net"
	"sort"
	"strconv"
	"strings"

	bubbleTable "github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/miekg/dns"
	"github.com/prionis/dns-server/cmd/tui/model/table"
	"github.com/prionis/dns-server/cmd/tui/structs"
	"github.com/prionis/dns-server/cmd/tui/transport"
)

// NewDescriptor creates a new table descriptor for the resource records table.
func NewDescriptor(w int) table.Descriptor {
	return table.Descriptor{
		Columns: GetColumns(w),

		RefreshFn: RefreshFn,

		FilterFn:     FilterFn,
		FilterFields: GetFilterFields(w),

		SortFn:   SortFn,
		SearchFn: SearchFn,
		DeleteFn: DeleteFn,

		AddFn:       AddFn,
		UpdateFn:    UpdateFn,
		InputFields: GetInputFields(w),
	}
}

// DeleteFn deletes a resource record with the specified ID using the provided transport.
func DeleteFn(transport *transport.Transport, id int64) error {
	return transport.DeleteRR(id)
}

// SearchFn searches for a query string in the resource record rows and returns the matching rows.
func SearchFn(query string, rows []bubbleTable.Row) []bubbleTable.Row {
	result := make([]bubbleTable.Row, 0, len(rows))
	for _, row := range rows {
		for _, cell := range row {
			if strings.Contains(strings.ToLower(cell), strings.ToLower(query)) {
				result = append(result, row)
				break
			}
		}
	}
	return result
}

// UpdateFn updates a resource record with the specified ID using the provided transport and input fields.
func UpdateFn(t *transport.Transport, inputs []textinput.Model, id int64) (bubbleTable.Row, error) {
	for i := range inputs {
		if inputs[i].Validate != nil {
			err := inputs[i].Validate(inputs[i].Value())
			if err != nil {
				return bubbleTable.Row{}, err
			}
		}
	}

	domain := inputs[0].Value()
	domain = dns.CanonicalName(domain)
	dataStr := inputs[1].Value()
	rrType := inputs[2].Value()
	class := inputs[3].Value()
	ttlStr := inputs[4].Value()
	if ttlStr == "" {
		ttlStr = "3600"
	}
	ttl64, err := strconv.ParseUint(ttlStr, 10, 32)
	if err != nil {
		return bubbleTable.Row{}, errors.New("bad TTL")
	}
	ttl := uint32(ttl64)

	_, err = dns.NewRR(fmt.Sprintf("%s %d %s %s %s",
		domain,
		ttl,
		class,
		rrType,
		dataStr,
	))
	if err != nil {
		return nil, err
	}

	err = t.UpdateRR(structs.RR{
		ID:     id,
		Domain: domain,
		Data:   dataStr,
		Type:   rrType,
		Class:  class,
		TTL:    ttl,
	})
	if err != nil {
		return nil, err
	}

	return []string{strconv.FormatInt(id, 10), domain, dataStr, class, rrType, ttlStr}, nil
}

// AddFn adds a new resource record using the provided transport and input fields.
func AddFn(transport *transport.Transport, inputFields []textinput.Model) (bubbleTable.Row, error) {
	for i := range inputFields {
		if inputFields[i].Validate != nil {
			err := inputFields[i].Validate(inputFields[i].Value())
			if err != nil {
				return bubbleTable.Row{}, err
			}
		}
	}

	rr := structs.RR{
		Domain: inputFields[0].Value(),
		Data:   inputFields[1].Value(),
		Type:   inputFields[2].Value(),
		Class:  inputFields[3].Value(),
	}

	t := dns.StringToType[inputFields[2].Value()]
	switch t {
	case dns.TypeA:
		parsedData := net.ParseIP(rr.Data).To4().String()
		if parsedData == "<nil>" {
			return bubbleTable.Row{}, fmt.Errorf("%s is incorrect IPv4 address", rr.Data)
		}
	case dns.TypeAAAA:
		parsedData := net.ParseIP(rr.Data).To16().String()
		if parsedData == "<nil>" {
			return bubbleTable.Row{}, errors.New("incorrect IPv6 address")
		}
	case dns.TypeCNAME, dns.TypeNS, dns.TypeMX:
		if _, ok := dns.IsDomainName(rr.Data); !ok {
			return bubbleTable.Row{}, errors.New("incorrect domain name in data")
		}
	}

	ttlStr := inputFields[4].Value()
	if ttlStr == "" {
		ttlStr = "3600"
	}
	ttl, err := strconv.ParseUint(ttlStr, 32, 10)
	if err != nil {
		return bubbleTable.Row{}, errors.New("bad TTL")
	}
	if ttl > math.MaxUint32 {
		return bubbleTable.Row{}, errors.New("too big TTL")
	}
	rr.TTL = uint32(ttl)

	rr, err = transport.AddRR(rr)
	if err != nil {
		return bubbleTable.Row{}, err
	}
	return bubbleTable.Row{
		strconv.FormatInt(rr.ID, 10),
		rr.Domain, rr.Type, rr.Class, rr.Data,
		ttlStr,
	}, nil
}

// GetInputFields returns the input fields for adding or updating a resource record based on the provided width.
func GetInputFields(width int) []textinput.Model {
	inputs := make([]textinput.Model, 5)

	for i := range inputs {
		input := textinput.New()
		input.Width = width / 3
		input.Prompt = "* > "
		switch i {
		case 0:
			setupDomainInput(input)
		case 1:
			input.Placeholder = "Data"
		case 2:
			setuppTypeInput(input)
		case 3:
			setupClassInput(input)
		case 4:
			input = setupTTLInput(input)
		}
		inputs[i] = input
	}

	return inputs
}

func setupTTLInput(input textinput.Model) textinput.Model {
	input.Placeholder = "Time to live"
	input.Prompt = "  > "
	input.Validate = func(s string) error {
		if s != "" {
			_, err := strconv.ParseInt(s, 10, 32)
			if err != nil {
				return errors.New("incorrect TTL value, it must be a number")
			}
		}
		return nil
	}
	return input
}

func setupClassInput(input textinput.Model) {
	input.Placeholder = "Class"
	input.ShowSuggestions = true
	input.Validate = func(s string) error {
		if s == "" {
			return errors.New("class must not be empty")
		}
		_, ok := dns.StringToClass[s]
		if !ok {
			return fmt.Errorf("uknown class '%s' of resource record", s)
		}
		return nil
	}
	input.SetSuggestions([]string{"IN", "CS", "CH", "HS"})
}

func setuppTypeInput(input textinput.Model) {
	input.Placeholder = "Type"
	input.ShowSuggestions = true
	input.Validate = func(s string) error {
		if s == "" {
			return errors.New("type must not be empty")
		}
		_, ok := dns.StringToType[s]
		if !ok {
			return fmt.Errorf("uknown type '%s' of resource record", s)
		}
		return nil
	}
	input.SetSuggestions([]string{
		"A", "NS", "MD", "MF", "CNAME", "SOA", "MB", "MG", "MR",
		"NULL", "WKS", "PTR", "HINFO", "MINFO", "MX", "TXT",
	})
}

func setupDomainInput(input textinput.Model) {
	input.Placeholder = "Domain name"
	input.Validate = func(s string) error {
		if s == "" {
			return errors.New("domain name must not be empty")
		}
		domain := dns.CanonicalName(s)
		if _, ok := dns.IsDomainName(domain); ok {
			return nil
		}
		return errors.New("incorrect domain name")
	}
	input.Focus()
}

// GetFilterFields returns the input fields for filtering resource records based on the provided width.
func GetFilterFields(width int) []textinput.Model {
	inputs := make([]textinput.Model, 4)
	for i := range inputs {
		t := textinput.New()
		t.CharLimit = 32
		t.Prompt = "> "
		t.Width = width / 3

		switch i {
		case 0:
			t.Focus()
			t.Placeholder = "Enter type(e.g. A)"
		case 1:
			t.Placeholder = "Enter class(e.g. IN)"

		case 2:
			t.Placeholder = "Enter lowest TTL"
		case 3:
			t.Placeholder = "Enter highest TTL"
		}
		inputs[i] = t
	}
	return inputs
}

// GetColumns returns the column definitions for the resource records table based on the provided width.
func GetColumns(width int) []bubbleTable.Column {
	return []bubbleTable.Column{
		{
			Title: "ID",
			Width: max(4, width/10-5),
		},
		{
			Title: "Domain",
			Width: max(8, width/10*3-5),
		},
		{
			Title: "Data",
			Width: max(10, width/10*3-5),
		},
		{
			Title: "Type",
			Width: max(5, width/10-5),
		},
		{
			Title: "Class",
			Width: max(2, width/10-5),
		},
		{
			Title: "TimeToLive",
			Width: max(6, width/10-5),
		},
	}
}

// FilterFn filters the rr rows.
func FilterFn(inputs []textinput.Model, rows []bubbleTable.Row) ([]bubbleTable.Row, error) {
	result := make([]bubbleTable.Row, 0)
	for _, row := range rows {
		if inputs[0].Value() != "" {
			if row[3] != inputs[0].Value() {
				continue
			}
		}

		if inputs[1].Value() != "" {
			if row[4] != inputs[1].Value() {
				continue
			}
		}

		ttl, err := strconv.ParseInt(row[5], 10, 32)
		if err != nil {
			return nil, err
		}
		if inputs[2].Value() != "" {
			lowest, err := strconv.ParseInt(inputs[2].Value(), 10, 32)
			if err != nil {
				return nil, err
			}
			if ttl < lowest && lowest != 0 {
				continue
			}
		}

		if inputs[3].Value() != "" {
			highest, err := strconv.ParseInt(inputs[3].Value(), 10, 32)
			if err != nil {
				return nil, err
			}
			if ttl > highest && highest != 0 {
				continue
			}
		}

		result = append(result, row)
	}

	return result, nil
}

// SortFn sorts the resource record rows based on the specified column index and order (ascending or descending).
func SortFn(index int, r []bubbleTable.Row, asc bool) []bubbleTable.Row {
	switch index {
	case 0: // ID
		sort.Slice(r, func(i, j int) bool {
			id1, err1 := strconv.ParseInt(r[i][0], 10, 64)
			id2, err2 := strconv.ParseInt(r[j][0], 10, 64)
			if err1 != nil || err2 != nil {
				return false
			}
			if asc {
				return id1 < id2
			}
			return id1 > id2
		})
	case 5: // TTL
		sort.Slice(r, func(i, j int) bool {
			ttl1, err1 := strconv.ParseInt(r[i][5], 10, 64)
			ttl2, err2 := strconv.ParseInt(r[j][5], 10, 64)
			if err1 != nil || err2 != nil {
				return false
			}
			if asc {
				return ttl1 < ttl2
			}
			return ttl1 > ttl2
		})
	case 2: // Data
		sort.Slice(r, func(i, j int) bool {
			return strings.ToLower(r[i][2]) > strings.ToLower(r[j][2])
		})
	case 1, 3, 4: // Domain, Type, Class
		sort.Slice(r, func(i, j int) bool {
			if asc {
				return strings.ToLower(r[i][index]) < strings.ToLower(r[j][index])
			}
			return strings.ToLower(r[i][index]) > strings.ToLower(r[j][index])
		})
	}
	return r
}

// RRButtonsHandler handles button actions for the resource records table based on the button index and the current table model.
func RRButtonsHandler(index int, m table.Model) (table.Model, tea.Cmd) {
	switch index {
	case 0: // View
		m.Table.Focus()
		return m, nil
	case 1: // Add
		return m, func() tea.Msg {
			return table.AddRequestMsg{}
		}
	case 2: // Delete
		return m, func() tea.Msg {
			return table.DeleteRequestMsg{}
		}
	case 3: // Update
		return m, func() tea.Msg {
			return table.UpdateRequestMsg{}
		}
	case 4: // Filter
		return m, func() tea.Msg {
			return table.FilterRequestMsg{}
		}
	case 5: // Sort
		return m, func() tea.Msg {
			return table.SortRequestMsg{}
		}
	case 6: // Reset
		return m, func() tea.Msg {
			return table.ResetRequestMsg{}
		}
	case 7: // Refresh
		return m, func() tea.Msg {
			return table.RefreshRequestMsg{}
		}
	case 8: // Export to Word
		return m, func() tea.Msg {
			return table.ExportToWordRequestMsg{}
		}
	case 9: // Export to Excel
		return m, func() tea.Msg {
			return table.ExportToExcelRequestMsg{}
		}
	}
	return m, nil
}

// RefreshFn is a function that refreshes the table
func RefreshFn(t *transport.Transport) ([]bubbleTable.Row, error) {
	rrs, err := t.GetAllRRs()
	if err != nil {
		return []bubbleTable.Row{}, err
	}
	rows := make([]bubbleTable.Row, len(rrs))

	for i, rr := range rrs {
		rows[i] = []string{strconv.FormatInt(rr.ID, 10), rr.Domain, rr.Data, rr.Type, rr.Class, strconv.FormatInt(int64(rr.TTL), 10)}
	}

	return rows, nil
}

// New is a function that creates a new table model
func New(t *transport.Transport, w, h int) (table.Model, error) {
	rows := make([]bubbleTable.Row, 0)
	rrs, err := t.GetAllRRs()
	if err != nil {
		return table.Model{}, err
	}
	for _, rr := range rrs {
		rows = append(rows, bubbleTable.Row{
			strconv.FormatInt(rr.ID, 10),
			rr.Domain, rr.Data, rr.Type, rr.Class,
			strconv.FormatInt(int64(rr.TTL), 10),
		})
	}

	buttons := []string{
		fmt.Sprintf("View %c ", '\uebb7'),
		fmt.Sprintf("Add %c ", '\uea60'),
		fmt.Sprintf("Delete %c ", '\uf00d'),
		fmt.Sprintf("Update %c ", '\uea73'),
		fmt.Sprintf("Filter %c ", '\ueaf1'),
		fmt.Sprintf("Sort %c ", '\ueaf1'),
		fmt.Sprintf("Reset %c ", '\ueaf1'),
		fmt.Sprintf("Refresh %c ", '\ueaf1'),
		fmt.Sprintf("Export to Word %c ", '\ue6a5'),
		fmt.Sprintf("Export to Excel %c ", '\uf1c3'),
	}
	return table.NewModel(NewDescriptor(w), w, h, buttons, rows, RRButtonsHandler, "Resource records"), nil
}
