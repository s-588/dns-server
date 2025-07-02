package userTable

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	bubbleTable "github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/prionis/dns-server/cmd/tui/model/table"
	"github.com/prionis/dns-server/cmd/tui/structs"
	"github.com/prionis/dns-server/cmd/tui/transport"
)

func NewDescriptor(w, h int) table.TableDescriptor {
	return table.TableDescriptor{
		Columns:      GetColumns(w),
		RefreshFn:    RefreshFn,
		FilterFn:     FilterFn,
		FilterFields: GetFilteredFields(w),
		SortFn:       SortFn,
		SearchFn:     SearchFn,
		DeleteFn:     DeleteFn,
		AddFn:        AddFn,
		UpdateFn:     UpdateFn,
		InputFields:  GetInputFields(w),
	}
}

func RefreshFn(t *transport.Transport) ([]bubbleTable.Row, error) {
	users, err := t.GetAllUsers()
	if err != nil {
		return []bubbleTable.Row{}, err
	}
	rows := make([]bubbleTable.Row, len(users))

	for i, u := range users {
		rows[i] = []string{strconv.FormatInt(int64(u.ID), 10), u.Login, u.FirstName, u.LastName, u.Role}
	}

	return rows, nil
}

func UpdateFn(t *transport.Transport, inputs []textinput.Model, id int32) (bubbleTable.Row, error) {
	for i := range inputs {
		if inputs[i].Validate != nil {
			err := inputs[i].Validate(inputs[i].Value())
			if err != nil {
				return bubbleTable.Row{}, err
			}
		}
	}
	login := inputs[0].Value()
	fname := inputs[1].Value()
	lname := inputs[2].Value()
	role := inputs[3].Value()
	password := inputs[4].Value()
	rpassword := inputs[5].Value()

	if password != rpassword {
		return bubbleTable.Row{}, errors.New("password is not match")
	}
	err := t.UpdateUser(structs.User{
		ID:        id,
		Login:     login,
		FirstName: fname,
		LastName:  lname,
		Role:      role,
		Password:  password,
	})
	if err != nil {
		return nil, err
	}

	return []string{strconv.FormatInt(int64(id), 10), login, fname, lname, role}, nil
}

func GetColumns(w int) []bubbleTable.Column {
	return []bubbleTable.Column{
		{
			Title: "ID",
			Width: max(4, w/10-5),
		},
		{
			Title: "Login",
			Width: max(4, w/10-5),
		},
		{
			Title: "First name",
			Width: max(8, w/10*3-5),
		},
		{
			Title: "Last name",
			Width: max(10, w/10*3-5),
		},
		{
			Title: "Role",
			Width: max(5, w/10-5),
		},
	}
}

func DeleteFn(t *transport.Transport, id int32) error {
	return t.DeleteUser(id)
}

func SearchFn(query string, rows []bubbleTable.Row) []bubbleTable.Row {
	result := make([]bubbleTable.Row, 0, len(rows))
	for _, row := range rows {
		if strings.Contains(strings.Join(row, ""), query) {
			result = append(result, row)
		}
	}
	return result
}

func AddFn(t *transport.Transport, inputs []textinput.Model) (bubbleTable.Row, error) {
	for i := range inputs {
		if inputs[i].Validate != nil {
			err := inputs[i].Validate(inputs[i].Value())
			if err != nil {
				return bubbleTable.Row{}, err
			}
		}
	}
	login := inputs[0].Value()
	fname := inputs[1].Value()
	lname := inputs[2].Value()
	role := inputs[3].Value()
	password := inputs[4].Value()
	rpassword := inputs[5].Value()
	if password != rpassword {
		return bubbleTable.Row{}, errors.New("passwords is not equal")
	}
	user, err := t.RegisterNewUser(structs.User{
		Login:     login,
		FirstName: fname,
		LastName:  lname,
		Role:      role,
		Password:  password,
	})
	return bubbleTable.Row{
		strconv.FormatInt(int64(user.ID), 10),
		user.Login, user.FirstName, user.LastName, user.Role,
	}, err
}

func GetInputFields(width int) []textinput.Model {
	inputs := make([]textinput.Model, 6)
	for i := range inputs {
		input := textinput.New()
		input.Width = width / 3
		switch i {
		case 0:
			input.Placeholder = "Login"
			input.Validate = func(s string) error {
				if len(s) < 5 {
					return errors.New("login is too short, must be at least 5 symbols")
				}
				if len(s) > 16 {
					return errors.New("login is too long, must less than 16 symbols")
				}
				if strings.ContainsAny(s, "._/\\^?!%+[{(&=)}]*") {
					return errors.New("login must not contain any special characters")
				}
				return nil
			}
			input.Focus()
		case 1:
			input.Placeholder = "First name"
			input.Validate = func(s string) error {
				if s == "" {
					return errors.New("first name must be filled")
				}
				if strings.ContainsAny(s, "1234567890+[{(&=)}]*!/-|`_?%^#@\\") {
					return errors.New("first name can't contain any numbers or special symbols")
				}
				return nil
			}
		case 2:
			input.Placeholder = "Last name"
			input.Validate = func(s string) error {
				if s == "" {
					return errors.New("last name must be filled")
				}
				if strings.ContainsAny(s, "1234567890+[{(&=)}]*!/-|`_?%^#@\\") {
					return errors.New("last name can't contain any numbers or special symbols")
				}
				return nil
			}
		case 3:
			input.Placeholder = "Role"
			input.ShowSuggestions = true
			input.SetSuggestions([]string{"admin", "user"})
			input.Validate = func(s string) error {
				if s == "" {
					return errors.New("role must be filled")
				}
				if strings.ContainsAny(s, "1234567890+[{(&=)}]*!/-|`_?%^#@\\") {
					return errors.New("role can't contain any numbers or special symbols")
				}
				return nil
			}
		case 4:
			input.Placeholder = "Password"
			input.Validate = func(s string) error {
				if s == "" {
					return errors.New("password must be filled")
				}
				if len(s) < 6 {
					return errors.New("password is too short, password must be atleast 6 characters")
				}
				if len(s) > 72 {
					return errors.New("password is too long")
				}
				if !strings.ContainsAny(s, "123456789") {
					return errors.New("password must contain atleast one number")
				}
				return nil
			}
			input.EchoMode = textinput.EchoPassword
			input.EchoCharacter = '*'
		case 5:
			input.Placeholder = "Repeat password"
			input.Validate = func(s string) error {
				if s == "" {
					return errors.New("password must be filled")
				}
				if len(s) < 6 {
					return errors.New("password is too short, password must be atleast 6 characters")
				}
				if len(s) > 72 {
					return errors.New("password is too long")
				}
				if !strings.ContainsAny(s, "123456789") {
					return errors.New("password must contain atleast one number")
				}
				return nil
			}
			input.EchoMode = textinput.EchoPassword
			input.EchoCharacter = '*'
		}
		inputs[i] = input
	}
	return inputs
}

func GetFilteredFields(width int) []textinput.Model {
	fields := make([]textinput.Model, 3)
	var input textinput.Model
	for i := range fields {
		input.Width = width / 3
		switch i {
		case 0:
			input.Placeholder = "First name"
		case 1:
			input.Placeholder = "Last name"
		case 2:
			input.Placeholder = "Role"
		}
		fields[i] = input
	}
	return fields
}

func FilterFn(inputs []textinput.Model, rows []bubbleTable.Row) ([]bubbleTable.Row, error) {
	result := make([]bubbleTable.Row, 0)
	fname, lname, role := inputs[0].Value(), inputs[1].Value(), inputs[2].Value()
	for _, row := range rows {
		if fname != "" {
			if fname != row[2] {
				continue
			}
		}
		if lname != "" {
			if fname != row[3] {
				continue
			}
		}
		if role != "" {
			if fname != row[4] {
				continue
			}
		}
		result = append(result, row)
	}
	return result, nil
}

func SortFn(index int, r []bubbleTable.Row, asc bool) []bubbleTable.Row {
	sort.Slice(r, func(i, j int) bool {
		if asc {
			return strings.ToLower(r[i][index]) > strings.ToLower(r[j][index])
		} else {
			return !(strings.ToLower(r[i][index]) > strings.ToLower(r[j][index]))
		}
	})
	return r
}

func userButtonsHandler(index int, m table.TableModel) (table.TableModel, tea.Cmd) {
	switch index {
	case 0: // View
		m.Table.Focus()
		return m, nil
	case 1: // Add
		return m, func() tea.Msg {
			return table.AddRequestMsg{}
		}
	case 2: // Update
		return m, func() tea.Msg {
			return table.UpdateRequestMsg{}
		}
	case 3: // Delete
		return m, func() tea.Msg {
			return table.DeleteRequestMsg{}
		}
	case 4: // Sort
		return m, func() tea.Msg {
			return table.SortRequestMsg{}
		}
	case 5: // Filter
		return m, func() tea.Msg {
			return table.FilterRequestMsg{}
		}
	case 6: // Reset filters
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

func New(t *transport.Transport, w, h int) (table.TableModel, error) {
	users, err := t.GetAllUsers()
	if err != nil {
		return table.TableModel{}, nil
	}
	rows := make([]bubbleTable.Row, 0, len(users))
	for _, u := range users {
		rows = append(rows, bubbleTable.Row{strconv.FormatInt(int64(u.ID), 10), u.Login, u.FirstName, u.LastName, u.Role})
	}

	buttons := []string{
		fmt.Sprintf("View %c ", '\uebb7'),
		fmt.Sprintf("Add %c ", '\uea60'),
		fmt.Sprintf("Update %c ", '\uea60'),
		fmt.Sprintf("Delete %c ", '\uf00d'),
		fmt.Sprintf("Sort %c ", '\ueaf1'),
		fmt.Sprintf("Filter %c ", '\ueaf1'),
		fmt.Sprintf("Reset filters %c ", '\ueaf1'),
		fmt.Sprintf("Refresh %c ", '\ueaf1'),
		fmt.Sprintf("Export to Word %c ", '\ue6a5'),
		fmt.Sprintf("Export to Excel %c ", '\uf1c3'),
	}
	return table.NewModel(NewDescriptor(w, h), w, h, buttons, rows, userButtonsHandler, "Users"), nil
}
