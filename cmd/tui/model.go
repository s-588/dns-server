package tui

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"

	"github.com/prionis/dns-server/cmd/tui/model/account"
	"github.com/prionis/dns-server/cmd/tui/model/crud"
	"github.com/prionis/dns-server/cmd/tui/model/export"
	"github.com/prionis/dns-server/cmd/tui/model/filter"
	"github.com/prionis/dns-server/cmd/tui/model/popup"
	"github.com/prionis/dns-server/cmd/tui/model/sort"
	"github.com/prionis/dns-server/cmd/tui/model/table"
	"github.com/prionis/dns-server/cmd/tui/structs"
	"github.com/prionis/dns-server/cmd/tui/transport"
)

const (
	// Minimum width and height of the screen to fit atleast one row in the tables
	minWidth  = 52
	minHeight = 22
)

const (
	// Focus layers represent selected by user element
	focusTabs = iota
	focusButtons
	focusTable
	focusAddModel
	focusDeleteModel
	focusUpdateModel
	focusFilterModel
	focusSortModel
	focusLoginModel
	focusExportModel
	focusSearch
)

var focusNames map[int]string = map[int]string{
	focusTabs:        "tabs",
	focusSearch:      "search",
	focusButtons:     "buttons",
	focusTable:       "table",
	focusAddModel:    "add page",
	focusDeleteModel: "delete page",
	focusUpdateModel: "update page",
	focusFilterModel: "filter page",
	focusSortModel:   "sort page",
	focusLoginModel:  "login page",
	focusExportModel: "export page",
}

type keyMap struct {
	Enter   key.Binding
	Up      key.Binding
	Down    key.Binding
	Right   key.Binding
	Left    key.Binding
	Search  key.Binding
	Sort    key.Binding
	Filter  key.Binding
	Refresh key.Binding
	Reset   key.Binding
	Add     key.Binding
	Delete  key.Binding
	Update  key.Binding
	Help    key.Binding
	Unfocus key.Binding
	Quit    key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Search, k.Sort, k.Filter},
		{k.Reset, k.Refresh},
		{k.Add, k.Delete, k.Update},
		{k.Help, k.Enter, k.Unfocus, k.Quit},
	}
}

var keys = keyMap{
	Enter: key.NewBinding(
		key.WithKeys("enter", "space"),
		key.WithHelp("enter/space", "select item"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("←/h", "move right"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("→/l", "move left"),
	),
	Search: key.NewBinding(
		key.WithKeys("/", "ctrl+f"),
		key.WithHelp("ctrl+f or /", "search"),
	),
	Sort: key.NewBinding(
		key.WithKeys("S"),
		key.WithHelp("S", "sort"),
	),
	Filter: key.NewBinding(
		key.WithKeys("F"),
		key.WithHelp("F", "filter"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("F5", "ctrl+r"),
		key.WithHelp("F5/ctrl+r", "refresh"),
	),
	Reset: key.NewBinding(
		key.WithKeys("ctrl+z", "R"),
		key.WithHelp("R/ctrl+z", "reset filtering"),
	),
	Add: key.NewBinding(
		key.WithKeys("ctrl+n", "insert", "A"),
		key.WithHelp("A/insert/ctrl+n", "add new"),
	),
	Delete: key.NewBinding(
		key.WithKeys("delete", "backspace", "D", "ctrl+d"),
		key.WithHelp("delete/backspace/D/ctrld+d", "delete"),
	),
	Update: key.NewBinding(
		key.WithKeys("F2", "ctrl+s", "U"),
		key.WithHelp("U/F2/ctrl+s", "update"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Unfocus: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "unfocus"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "Q"),
		key.WithHelp("Q/ctrl+c", "quit"),
	),
}

// This is the main model of the user interface.
// It render all other buttons, tables and other models.
type model struct {
	// Width of the screen
	width int
	// Height of the screen
	height int
	user   *structs.User

	// What element user use right now
	focusLayer int
	// Tabs for select table
	tabs        []string
	selectedTab int

	tables  []table.TableModel
	logChan chan transport.LogMsg

	// Model for popup notifications.
	popup popup.PopupModel

	loginPage account.LoginModel
	// Model for adding resource records to the database.
	addModel crud.AddModel
	// Model for updating resource records of the database.
	updatePage crud.UpdateModel
	// Model for deleting resource records from the database.
	deletePage  crud.DeleteModel
	searchInput textinput.Model
	sortPage    sort.SortModel
	filterPage  filter.FilterModel
	exportModel export.ExportModel

	keys keyMap
	help help.Model

	transport *transport.Transport
}

// Creating new model of the user interface.
func NewModel() (model, error) {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		slog.Error("can't get term size")
	}
	if w < minWidth || h < minHeight {
		return model{},
			fmt.Errorf("Minimum size of the screen is %dx%d. Current is %dx%d",
				minWidth, minHeight, w, h)
	}
	w, h = w-8, max(3, h-20)

	t, err := transport.New("172.31.155.196:8083")
	if err != nil {
		return model{}, fmt.Errorf("can't create http transport: %w", err)
	}

	searchInput := textinput.New()
	searchInput.Prompt = fmt.Sprintf("%c  > ", '\uea6d')
	searchInput.Placeholder = "Type here to search"
	searchInput.ShowSuggestions = true
	searchInput.Width = w / 3

	return model{
		loginPage:   account.NewLoginModel(t, w, h),
		focusLayer:  focusLoginModel,
		exportModel: export.NewExportModel(t, w, h),
		searchInput: searchInput,
		help:        help.New(),

		keys:      keys,
		width:     w,
		height:    h,
		logChan:   make(chan transport.LogMsg, 1),
		transport: t,

		popup: popup.NewPopupModel(),
	}, nil
}

// Exit point of the app.
func (m model) Close() tea.Cmd {
	return tea.Quit
}

// Initialize esential things.
func (m model) Init() tea.Cmd {
	return tea.Batch(popup.ListenForPopupMsg(m.popup.MsgChan))
}
