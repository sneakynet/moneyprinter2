package tui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sneakynet/moneyprinter2/pkg/types"
)

// runKey feeds a key to the model and reports the returned command.
// Commands returned by the model must never be invoked in tests: they
// perform network fetches or start print jobs.
func runKey(t *testing.T, m *model, keyType tea.KeyType, runes ...rune) tea.Cmd {
	t.Helper()
	msg := tea.KeyMsg{Type: keyType}
	if keyType == tea.KeyRunes {
		msg.Runes = runes
	}
	next, cmd := m.Update(msg)
	*m = next.(model)
	return cmd
}

func mustAccountItems(n int) []list.Item {
	items := make([]list.Item, n)
	for i := 0; i < n; i++ {
		items[i] = accountItem{account: types.Account{ID: uint(i + 1), Name: "Alpha"}}
	}
	return items
}

// isQuitCmd reports whether the model returned the quit command. It must
// never invoke other commands: they perform network fetches or printing.
func isQuitCmd(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	return reflect.ValueOf(cmd).Pointer() == reflect.ValueOf(tea.Quit).Pointer()
}

// TestAccountSearchKeyRouting verifies that the search box is usable from the
// accounts phase: "/" (or "f") opens the filter, typed keys reach the filter
// input instead of the global q/Escape/Enter handlers, and Escape/Enter
// behave natively while typing.
func TestAccountSearchKeyRouting(t *testing.T) {
	m := NewBillViewer(nil, "", nil)
	next, cmd := m.Update(tea.WindowSizeMsg{Width: 60, Height: 20})
	if cmd != nil {
		t.Fatalf("WindowSizeMsg: unexpected cmd %v", cmd)
	}
	m = next.(model)
	// Greeting -> accounts. The returned cmd performs a network fetch;
	// it must never be invoked in tests.
	if cmd := runKey(t, &m, tea.KeyEnter); cmd == nil {
		t.Fatal("Enter on greeting: expected loadAccounts cmd")
	}
	if m.phase != "accounts" {
		t.Fatalf("phase = %q, want accounts", m.phase)
	}
	next, cmd = m.Update(accountsLoadedMsg{items: mustAccountItems(3)})
	if cmd != nil {
		t.Fatalf("accountsLoadedMsg: unexpected cmd %v", cmd)
	}
	m = next.(model)

	// The search box must be visible even before the filter is active.
	if !strings.Contains(m.accountsView(), "search accounts") {
		t.Errorf("accountsView does not show the search box:\n%s", m.accountsView())
	}

	// Pressing "/" starts filter mode without quitting.
	if cmd := runKey(t, &m, tea.KeyRunes, '/'); isQuitCmd(cmd) {
		t.Fatal("pressing '/' quit the app")
	}
	if !m.accountList.SettingFilter() {
		t.Fatal("filter mode did not start after '/'")
	}

	// While typing, a letter must reach the filter input, not the app.
	if cmd := runKey(t, &m, tea.KeyRunes, 'a'); isQuitCmd(cmd) {
		t.Fatal("typing a letter in the filter quit the app")
	}
	if got := m.accountList.FilterValue(); got != "a" {
		t.Fatalf("FilterValue() = %q, want %q", got, "a")
	}
	if !m.accountList.SettingFilter() {
		t.Fatal("filter mode ended while typing")
	}

	// "q" is the app's quit key, but it must be swallowed by the filter.
	if cmd := runKey(t, &m, tea.KeyRunes, 'q'); isQuitCmd(cmd) {
		t.Fatal("typing 'q' in the filter quit the app")
	}
	if got := m.accountList.FilterValue(); got != "aq" {
		t.Fatalf("FilterValue() = %q, want %q", got, "aq")
	}
	if !m.accountList.SettingFilter() {
		t.Fatal("filter mode ended after typing 'q'")
	}

	// Escape cancels the filter; the app must stay on the accounts phase.
	if cmd := runKey(t, &m, tea.KeyEscape); isQuitCmd(cmd) {
		t.Fatal("Escape in the filter quit the app")
	}
	if m.accountList.SettingFilter() {
		t.Fatal("filter was still active after Escape")
	}
	if m.accountList.IsFiltered() {
		t.Fatal("filter was still applied after cancelling with Escape")
	}
	if m.phase != "accounts" {
		t.Fatalf("phase = %q after Escape, want accounts", m.phase)
	}

	// The "f" shortcut also starts the filter.
	if cmd := runKey(t, &m, tea.KeyRunes, 'f'); isQuitCmd(cmd) {
		t.Fatal("pressing 'f' quit the app")
	}
	if !m.accountList.SettingFilter() {
		t.Fatal("filter mode did not start after 'f'")
	}
	if cmd := runKey(t, &m, tea.KeyEscape); isQuitCmd(cmd) {
		t.Fatal("Escape in the filter quit the app")
	}
	if m.accountList.SettingFilter() {
		t.Fatal("filter was still active after cancelling with 'f'")
	}

	// Enter applies the filter, the second Enter selects the item and
	// advances to the LEC phase.
	if cmd := runKey(t, &m, tea.KeyRunes, '/'); isQuitCmd(cmd) {
		t.Fatal("pressing '/' quit the app")
	}
	if cmd := runKey(t, &m, tea.KeyRunes, 'a'); isQuitCmd(cmd) {
		t.Fatal("pressing 'a' quit the app")
	}
	if cmd := runKey(t, &m, tea.KeyEnter); isQuitCmd(cmd) {
		t.Fatal("Enter applied the filter and quit the app")
	}
	if m.accountList.SettingFilter() {
		t.Fatal("filter was still active after Enter")
	}
	if !m.accountList.IsFiltered() {
		t.Fatal("filter was not applied after Enter")
	}
	if m.phase != "accounts" {
		t.Fatalf("phase = %q after applying the filter, want accounts", m.phase)
	}
	// The filter must be reusable after it was applied.
	if cmd := runKey(t, &m, tea.KeyRunes, '/'); isQuitCmd(cmd) {
		t.Fatal("pressing '/' again quit the app")
	}
	if !m.accountList.SettingFilter() {
		t.Fatal("filter could not be reopened after applying")
	}
	if cmd := runKey(t, &m, tea.KeyEscape); isQuitCmd(cmd) {
		t.Fatal("Escape in the filter quit the app")
	}
	// The second Enter selects the highlighted account and moves on.
	if cmd := runKey(t, &m, tea.KeyEnter); isQuitCmd(cmd) {
		t.Fatal("second Enter quit the app")
	}
	if m.phase != "lecs" {
		t.Fatalf("phase = %q after selecting an account, want lecs", m.phase)
	}
	if m.selectedAccountID == 0 {
		t.Fatal("selectedAccountID was not set on selection")
	}
}
