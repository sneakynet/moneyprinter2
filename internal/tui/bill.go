package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sneakynet/moneyprinter2/pkg/types"
)

const defaultGreeting = "MoneyPrinter Bill Viewer"

// -- List items -----------------------------------------------------------

type accountItem struct {
	account types.Account
}

func (i accountItem) Title() string {
	s := fmt.Sprintf("#%d %s", i.account.ID, i.account.Name)
	if i.account.Alias != "" {
		s += fmt.Sprintf(" (%s)", i.account.Alias)
	}
	return s
}

func (i accountItem) Description() string { return "" }

func (i accountItem) FilterValue() string {
	return fmt.Sprintf("%d %s %s", i.account.ID, i.account.Name, i.account.Alias)
}

type lecItem struct {
	lec types.LEC
}

func (i lecItem) Title() string {
	s := i.lec.Name
	if i.lec.Byline != "" {
		s += fmt.Sprintf(" - %s", i.lec.Byline)
	}
	return s
}

func (i lecItem) Description() string { return "" }

func (i lecItem) FilterValue() string { return i.lec.Name + " " + i.lec.Byline }

// -- Messages -------------------------------------------------------------

type errMsg error

type accountsLoadedMsg struct {
	items []list.Item
}

type lecsLoadedMsg struct {
	items []list.Item
}

type billLoadedMsg struct {
	text string
}

// -- Model ---------------------------------------------------------------

type model struct {
	width    int
	height   int
	greeting string
	client   *Client

	// Phase: "greeting" | "accounts" | "lecs" | "bill"
	phase string

	// Account list
	accountList list.Model

	// LEC list
	lecList list.Model

	// Bill display
	billText  string
	billError string
	viewport  viewport.Model

	// Selected IDs
	selectedAccountID uint
	selectedLECID     uint

	// Loading state
	loading   bool
	loadMsg   string
	loadError string
}

// NewBillViewer creates a bill viewer model with the given greeting.
// If greeting is empty, the default greeting is used.
func NewBillViewer(client *Client, greeting string) model {
	if greeting == "" {
		greeting = defaultGreeting
	}

	// Account list
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = selectedItemStyle
	delegate.Styles.NormalTitle = itemStyle
	accountList := list.New(nil, delegate, 0, 0)
	accountList.Title = ""
	accountList.Styles.Title = titleStyle
	accountList.Styles.StatusBar = statusStyle
	accountList.Styles.StatusBarActiveFilter = statusStyle
	accountList.Styles.PaginationStyle = statusStyle
	accountList.SetShowHelp(false)

	// LEC list
	lecDelegate := list.NewDefaultDelegate()
	lecDelegate.Styles.SelectedTitle = selectedItemStyle
	lecDelegate.Styles.NormalTitle = itemStyle
	lecList := list.New(nil, lecDelegate, 0, 0)
	lecList.Title = ""
	lecList.Styles.Title = titleStyle
	lecList.Styles.StatusBar = statusStyle
	lecList.Styles.StatusBarActiveFilter = statusStyle
	lecList.Styles.PaginationStyle = statusStyle
	lecList.SetShowHelp(false)

	return model{
		greeting:    greeting,
		client:      client,
		phase:       "greeting",
		accountList: accountList,
		lecList:     lecList,
		viewport:    viewport.New(0, 0),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

// -- View -----------------------------------------------------------------

func (m model) View() string {
	var v string
	switch m.phase {
	case "greeting":
		v = m.greetingView()
	case "accounts":
		v = m.accountsView()
	case "lecs":
		v = m.lecsView()
	case "bill":
		v = m.billView()
	}
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, v)
	}
	return v
}

func (m model) greetingView() string {
	s := titleStyle.Render(m.greeting)
	msg := statusStyle.Render("\nPress Enter to continue, 'q' to quit.")
	return s + "\n" + msg
}

func (m model) accountsView() string {
	s := titleStyle.Render("Accounts")
	if m.loading {
		return s + "\n" + statusStyle.Render(m.loadMsg)
	}
	if m.loadError != "" {
		return s + "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(m.loadError) +
			"\n\n" + statusStyle.Render("Press Enter to retry, 'q' to quit")
	}
	return s + "\n" + m.accountList.View()
}

func (m model) lecsView() string {
	s := titleStyle.Render("LECs")
	if m.loading {
		return s + "\n" + statusStyle.Render(m.loadMsg)
	}
	if m.loadError != "" {
		return s + "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(m.loadError) +
			"\n\n" + statusStyle.Render("Press Enter to retry, 'esc' to go back")
	}
	info := statusStyle.Render(fmt.Sprintf("Account #%d - Press 'esc' to go back\n", m.selectedAccountID))
	return s + "\n" + info + m.lecList.View()
}

func (m model) billView() string {
	s := titleStyle.Render("Bill")
	if m.loading {
		return s + "\n" + statusStyle.Render(m.loadMsg)
	}

	if m.billError != "" || m.loadError != "" {
		err := m.billError
		if err == "" {
			err = m.loadError
		}
		return s + "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(err) +
			"\n\n" + statusStyle.Render("Press Enter to retry, 'esc' to go back")
	}

	info := statusStyle.Render("Press 'esc' to go back, 'q' for new bill")
	return s + "\n" + info + "\n" + m.viewport.View()
}

// -- Update ---------------------------------------------------------------

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		w := m.width - 2
		h := m.height - titleStyle.GetPaddingTop() - titleStyle.GetPaddingBottom() - titleStyle.GetMarginBottom() - 5
		if w > 0 && h > 0 {
			m.accountList.SetWidth(w)
			m.accountList.SetHeight(h)
			m.lecList.SetWidth(w)
			m.lecList.SetHeight(h)
		}
		m.viewport.Width = w
		m.viewport.Height = h - 2

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			if m.phase == "bill" {
				m.phase = "accounts"
				m.billText = ""
				m.billError = ""
				m.selectedAccountID = 0
				m.selectedLECID = 0
				return m, nil
			}
			return m, tea.Quit

		case tea.KeyEscape:
			if m.phase == "bill" {
				m.phase = "lecs"
				m.billText = ""
				m.billError = ""
				m.loadError = ""
				m.selectedLECID = 0
				return m, nil
			}
			if m.phase == "lecs" {
				m.phase = "accounts"
				m.loadError = ""
				m.selectedAccountID = 0
				m.selectedLECID = 0
				return m, nil
			}

		default:
			if msg.String() == "q" {
				return m, tea.Quit
			}

		case tea.KeyEnter:
			if m.phase == "greeting" {
				m.phase = "accounts"
				m.loading = true
				m.loadMsg = "Loading accounts..."
				return m, m.loadAccounts
			}
			if m.phase == "accounts" {
				if m.loadError != "" {
					// Retry loading accounts
					m.loading = true
					m.loadMsg = "Loading accounts..."
					m.loadError = ""
					return m, m.loadAccounts
				}
				item := m.accountList.SelectedItem().(accountItem)
				m.selectedAccountID = item.account.ID
				m.phase = "lecs"
				m.loading = true
				m.loadMsg = "Loading LECs..."
				m.loadError = ""
				return m, m.loadLECs
			}
			if m.phase == "lecs" {
				if m.loadError != "" {
					// Retry loading LECs
					m.loading = true
					m.loadMsg = "Loading LECs..."
					m.loadError = ""
					return m, m.loadLECs
				}
				item := m.lecList.SelectedItem().(lecItem)
				m.selectedLECID = item.lec.ID
				m.phase = "bill"
				m.loading = true
				m.loadMsg = "Loading bill..."
				m.loadError = ""
				return m, m.loadBill
			}
			if m.phase == "bill" && m.loadError != "" {
				// Retry loading bill
				m.loading = true
				m.loadMsg = "Loading bill..."
				m.loadError = ""
				m.billError = ""
				return m, m.loadBill
			}
		}

	// Handle async loading messages
	case accountsLoadedMsg:
		m.loading = false
		m.accountList.SetItems(msg.items)
		return m, nil
	case lecsLoadedMsg:
		m.loading = false
		m.lecList.SetItems(msg.items)
		return m, nil
	case billLoadedMsg:
		m.loading = false
		m.billText = msg.text
		m.viewport.SetContent(msg.text)
		return m, nil
	case errMsg:
		m.loading = false
		m.loadError = msg.Error()
		if m.phase == "bill" {
			m.billError = msg.Error()
		}
		return m, nil
	}

	switch m.phase {
	case "accounts":
		m.accountList, cmd = m.accountList.Update(msg)
		return m, cmd
	case "lecs":
		m.lecList, cmd = m.lecList.Update(msg)
		return m, cmd
	case "bill":
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	return m, cmd
}

// -- Commands --------------------------------------------------------------

func (m model) loadAccounts() tea.Msg {
	accounts, err := m.client.FetchAccounts()
	if err != nil {
		return errMsg(err)
	}

	items := make([]list.Item, len(accounts))
	for i, a := range accounts {
		items[i] = accountItem{account: a}
	}
	return accountsLoadedMsg{items: items}
}

func (m model) loadLECs() tea.Msg {
	lecs, err := m.client.FetchLECs(m.selectedAccountID)
	if err != nil {
		return errMsg(err)
	}

	items := make([]list.Item, len(lecs))
	for i, lec := range lecs {
		items[i] = lecItem{lec: lec}
	}
	return lecsLoadedMsg{items: items}
}

func (m model) loadBill() tea.Msg {
	text, err := m.client.FetchBill(m.selectedAccountID, m.selectedLECID)
	if err != nil {
		return errMsg(err)
	}
	return billLoadedMsg{text: text}
}
