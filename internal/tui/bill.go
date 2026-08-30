package tui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
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

type printerItem struct {
	printer Printer
}

func (i printerItem) Title() string       { return i.printer.Label }
func (i printerItem) Description() string { return "" }
func (i printerItem) FilterValue() string { return i.printer.Label }

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

	// Phase: "greeting" | "accounts" | "lecs" | "bill" | "printers" | "printing"
	phase string

	// Account list
	accountList list.Model

	// LEC list
	lecList list.Model

	// Printer list
	printerList list.Model
	printers    []Printer

	// Bill display
	billText  string
	billError string
	viewport  viewport.Model

	// Selected IDs
	selectedAccountID uint
	selectedLECID     uint

	// Printing phase
	printerSpinner spinner.Model
	printingPrinter string

	// Print status
	printStatus   string
	printError    string
	buttonFocused bool

	// Loading state
	loading   bool
	loadMsg   string
	loadError string
}

// NewBillViewer creates a bill viewer model with the given greeting and printers.
// If greeting is empty, the default greeting is used.
func NewBillViewer(client *Client, greeting string, printers []Printer) model {
	if greeting == "" {
		greeting = defaultGreeting
	}

	// Account list
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetSpacing(0)
	delegate.Styles.SelectedTitle = selectedItemStyle
	delegate.Styles.NormalTitle = itemStyle
	accountList := list.New(nil, delegate, 0, 0)
	accountList.Title = ""
	accountList.Styles.Title = titleStyle
	accountList.Styles.StatusBar = statusStyle
	accountList.Styles.StatusBarActiveFilter = statusStyle
	accountList.Styles.PaginationStyle = statusStyle
	accountList.SetShowHelp(false)
	accountList.SetShowStatusBar(false)
	accountList.SetShowPagination(false)

	// LEC list
	lecDelegate := list.NewDefaultDelegate()
	lecDelegate.ShowDescription = false
	lecDelegate.SetSpacing(0)
	lecDelegate.Styles.SelectedTitle = selectedItemStyle
	lecDelegate.Styles.NormalTitle = itemStyle
	lecList := list.New(nil, lecDelegate, 0, 0)
	lecList.Title = ""
	lecList.Styles.Title = titleStyle
	lecList.Styles.StatusBar = statusStyle
	lecList.Styles.StatusBarActiveFilter = statusStyle
	lecList.Styles.PaginationStyle = statusStyle
	lecList.SetShowHelp(false)
	lecList.SetShowStatusBar(false)
	lecList.SetShowPagination(false)

	// Printer list
	printerDelegate := list.NewDefaultDelegate()
	printerDelegate.ShowDescription = false
	printerDelegate.SetSpacing(0)
	printerDelegate.Styles.SelectedTitle = selectedItemStyle
	printerDelegate.Styles.NormalTitle = itemStyle
	printerList := list.New(nil, printerDelegate, 0, 0)
	printerList.Title = ""
	printerList.Styles.Title = titleStyle
	printerList.Styles.StatusBar = statusStyle
	printerList.Styles.StatusBarActiveFilter = statusStyle
	printerList.Styles.PaginationStyle = statusStyle
	printerList.SetShowHelp(false)
	printerList.SetShowStatusBar(false)
	printerList.SetShowPagination(false)

	// Spinner for printing phase
	printerSpinner := spinner.New()
	printerSpinner.Style = statusStyle

	return model{
		greeting:       greeting,
		client:         client,
		phase:          "greeting",
		accountList:    accountList,
		lecList:        lecList,
		printerList:    printerList,
		printers:       printers,
		viewport:       viewport.New(0, 0),
		printerSpinner: printerSpinner,
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
	case "printers":
		v = m.printersView()
	case "printing":
		v = m.printingView()
	}
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, v)
	}
	return v
}

func (m model) greetingView() string {
	greetingStyleWithWidth := greetingStyle.Width(m.width)
	s := greetingStyleWithWidth.Render(m.greeting)
	var status string
	if m.printStatus != "" {
		status = m.printStatus
	} else if m.printError != "" {
		status = m.printError
	} else {
		status = "Press Enter to continue, 'q' to quit."
	}
	return strings.TrimRight(s, "\n") + "\n" + greetingStyleWithWidth.Render(statusStyle.Render(status))
}

func (m model) accountsView() string {
	// Use a reasonable dialog width so centering is visible.
	dialogWidth := 60
	if m.width-4 < dialogWidth {
		dialogWidth = m.width - 4
	}
	title := titleStyle.Render("Accounts")
	if m.loading {
		return title + "\n" + statusStyle.Render(m.loadMsg)
	}
	if m.loadError != "" {
		return title + "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(m.loadError) +
			"\n\n" + statusStyle.Render("Press Enter to retry, 'q' to quit")
	}
	m.accountList.SetWidth(dialogWidth)
	m.accountList.SetHeight(m.height - 6)
	dialog := title + "\n" + m.accountList.View()
	return dialogStyle.Render(dialog)
}

func (m model) lecsView() string {
	// Use a reasonable dialog width so centering is visible.
	dialogWidth := 60
	if m.width-4 < dialogWidth {
		dialogWidth = m.width - 4
	}
	title := titleStyle.Render("LECs")
	if m.loading {
		return title + "\n" + statusStyle.Render(m.loadMsg)
	}
	if m.loadError != "" {
		return title + "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(m.loadError) +
			"\n\n" + statusStyle.Render("Press Enter to retry, 'esc' to go back")
	}
	info := statusStyle.Render(fmt.Sprintf("Account #%d - Press 'esc' to go back\n", m.selectedAccountID))
	m.lecList.SetWidth(dialogWidth)
	m.lecList.SetHeight(m.height - 7)
	dialog := title + "\n" + info + m.lecList.View()
	return dialogStyle.Render(dialog)
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

	var hint string
	if m.printStatus != "" {
		hint = m.printStatus
	} else if m.printError != "" {
		hint = m.printError
	} else {
		hint = "Press 'esc' to go back, 'q' for new bill"
	}

	// title Render ends with \n (MarginBottom) so no extra \n needed before hint.
	// hint Render has MarginTop blank, then text, then \n — the final "\n" separates from viewport.
	content := s + statusStyle.Render(hint) + "\n" + m.viewport.View()

	if len(m.printers) > 0 {
		btnText := "[ Print ]"
		var btn lipgloss.Style
		if m.buttonFocused {
			btn = buttonFocusedStyle
		} else {
			btn = buttonStyle
		}
		// Button style has MarginTop — no extra "\n" needed.
		content += btn.Render(btnText)
	}
	return content
}

// -- Update ---------------------------------------------------------------

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		w := m.width - 2
		h := m.height - 5
		if w > 0 && h > 0 {
			m.accountList.SetWidth(w)
			m.accountList.SetHeight(h)
			m.lecList.SetWidth(w)
			m.lecList.SetHeight(h)
			m.printerList.SetWidth(w)
			m.printerList.SetHeight(h)
		}

		// Viewport: 80 chars wide (bill width), height sized to leave room for
		// title(2 lines with margin) + hint(2 lines with margin) + blank + viewport + button(2 lines with margin)
		viewWidth := 80
		if m.width-4 < viewWidth {
			viewWidth = m.width - 4
		}
		viewHeight := m.height - 5
		if len(m.printers) > 0 {
			viewHeight = m.height - 7
		}
		if viewWidth > 0 && viewHeight > 0 {
			m.viewport.Width = viewWidth
			m.viewport.Height = viewHeight
		}

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

		default:
			if msg.String() == "q" {
				return m, tea.Quit
			}

		case tea.KeyEscape:
			if m.phase == "bill" {
				m.phase = "lecs"
				m.billText = ""
				m.billError = ""
				m.loadError = ""
				m.selectedLECID = 0
				m.buttonFocused = false
				return m, nil
			}
			if m.phase == "lecs" {
				m.phase = "accounts"
				m.loadError = ""
				m.selectedAccountID = 0
				m.selectedLECID = 0
				return m, nil
			}
		if m.phase == "printers" {
				m.phase = "bill"
				m.printError = ""
				m.printStatus = ""
				return m, nil
			}

		case tea.KeyTab:
			if m.phase == "bill" && len(m.printers) > 0 {
				m.buttonFocused = true
				return m, nil
			}

		case tea.KeyShiftTab:
			if m.phase == "bill" {
				m.buttonFocused = false
				return m, nil
			}

		case tea.KeyEnter:
			// Enter on focused print button
			if m.phase == "bill" && m.buttonFocused && len(m.printers) > 0 {
				m.phase = "printers"
				m.buttonFocused = false
				m.printStatus = ""
				m.printError = ""
				items := make([]list.Item, len(m.printers))
				for i, p := range m.printers {
					items[i] = printerItem{printer: p}
				}
				m.printerList.SetItems(items)
				return m, nil
			}
		if m.phase == "printers" {
				item := m.printerList.SelectedItem().(printerItem)
				m.phase = "printing"
				m.printingPrinter = item.printer.Label
				return m, tea.Batch(
					m.printerSpinner.Tick,
					func() tea.Msg {
						return m.printBillCmd(item.printer.Path)
					},
				)
			}
		if m.phase == "greeting" {
				m.phase = "accounts"
				m.printStatus = ""
				m.printError = ""
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
	case printDoneMsg:
		if msg.err != nil {
			m.printError = msg.err.Error()
		} else {
			m.printStatus = "Print complete."
		}
		// Reset to greeting phase
		m.phase = "greeting"
		m.billText = ""
		m.billError = ""
		m.loadError = ""
		m.selectedAccountID = 0
		m.selectedLECID = 0
		m.buttonFocused = false
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
	case "printers":
		m.printerList, cmd = m.printerList.Update(msg)
		return m, cmd
	case "printing":
		var printingCmd tea.Cmd
		m.printerSpinner, printingCmd = m.printerSpinner.Update(msg)
		cmd = printingCmd
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

func (m model) printersView() string {
	// Use a reasonable dialog width so centering is visible.
	dialogWidth := 60
	if m.width-4 < dialogWidth {
		dialogWidth = m.width - 4
	}
	// Use a reasonable dialog height so vertical centering is visible.
	dialogHeight := 15
	if m.height-4 < dialogHeight {
		dialogHeight = m.height - 4
	}

	if m.printError != "" {
		errBox := overlayStyle.Render(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(m.printError) +
			"\n\n" + statusStyle.Render("Press 'esc' to go back"))
		return errBox
	}
	title := titleStyle.Render("Select Printer")
	info := statusStyle.Render("\nPress Enter to print, Esc to go back")
	m.printerList.SetWidth(dialogWidth)
	m.printerList.SetHeight(dialogHeight - 4)
	dialog := title + "\n" + info + "\n" + m.printerList.View()
	return dialogStyle.Render(dialog)
}

func (m model) printingView() string {
	if m.printError != "" {
		errBox := overlayStyle.Render(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(m.printError) +
			"\n\n" + statusStyle.Render("Press 'esc' to go back"))
		return errBox
	}
	return statusStyle.Render(m.printerSpinner.View() + " Sending to " + m.printingPrinter + "...\n\n") +
		statusStyle.Render("Press 'esc' to cancel")
}

type printDoneMsg struct {
	err error
}

func (m model) printBillCmd(path string) tea.Msg {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return printDoneMsg{err: fmt.Errorf("opening printer: %w", err)}
	}
	defer f.Close()

	if _, err := io.WriteString(f, m.billText); err != nil {
		return printDoneMsg{err: fmt.Errorf("writing to printer: %w", err)}
	}

	if err := f.Sync(); err != nil {
		return printDoneMsg{err: fmt.Errorf("flushing printer: %w", err)}
	}

	return printDoneMsg{err: nil}
}
