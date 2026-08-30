package tui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

// Client handles HTTP requests to the MoneyPrinter API.
type Client struct {
	addr     string
	username string
	password string
	http     *http.Client
}

// NewClient creates an API client from environment variables.
func NewClient() *Client {
	addr := os.Getenv("MP_ADDR")
	if addr == "" {
		addr = "localhost:8000"
	}

	return &Client{
		addr:     addr,
		username: os.Getenv("MP_USERNAME"),
		password: os.Getenv("MP_PASSWORD"),
		http:     &http.Client{Timeout: 30 * time.Second},
	}
}

// newRequest builds an authenticated GET request for the given path and query.
func (c *Client) newRequest(path string, query url.Values) (*http.Request, error) {
	req, err := http.NewRequest("GET", "", nil)
	if err != nil {
		return nil, err
	}
	req.URL = &url.URL{
		Scheme:   "http",
		Host:     c.addr,
		Path:     path,
		RawQuery: query.Encode(),
		User:     url.UserPassword(c.username, c.password),
	}
	return req, nil
}

// FetchAccounts returns all accounts from the server.
func (c *Client) FetchAccounts() ([]types.Account, error) {
	req, err := c.newRequest("/api/admin/accounts", nil)
	if err != nil {
		return nil, fmt.Errorf("building accounts request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching accounts: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("accounts request failed (status %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var accounts []types.Account
	if err := json.NewDecoder(resp.Body).Decode(&accounts); err != nil {
		return nil, fmt.Errorf("decoding accounts: %w", err)
	}

	return accounts, nil
}

// FetchLECs returns the LECs that service the given account ID.
func (c *Client) FetchLECs(accountID uint) ([]types.LEC, error) {
	req, err := c.newRequest(fmt.Sprintf("/api/admin/accounts/%d", accountID), nil)
	if err != nil {
		return nil, fmt.Errorf("building LECs request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching LECs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("LECs request failed (status %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var result struct {
		LECs []types.LEC `json:"lecs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding LECs: %w", err)
	}

	return result.LECs, nil
}

// FetchBill returns the text bill for the given account and LEC.
func (c *Client) FetchBill(accountID, lecID uint) (string, error) {
	query := url.Values{}
	query.Set("lec", fmt.Sprintf("%d", lecID))
	query.Set("width", "80")

	req, err := c.newRequest(
		fmt.Sprintf("/api/admin/money/bills/by-account/%d", accountID),
		query,
	)
	if err != nil {
		return "", fmt.Errorf("building bill request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching bill: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("bill request failed (status %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading bill: %w", err)
	}

	return string(body), nil
}

// Printer represents a configured line printer device.
type Printer struct {
	Label string
	Path  string
}

// ParsePrinters parses MONEYPRINTER_PRINTERS into a list of printers.
// Format: "label1:/path/to/lp0,label2:/path/to/lp1,..."
func ParsePrinters(raw string) []Printer {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	var printers []Printer
	for _, entry := range strings.Split(raw, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		parts := strings.SplitN(entry, ":", 2)
		if len(parts) == 2 {
			printers = append(printers, Printer{
				Label: strings.TrimSpace(parts[0]),
				Path:  strings.TrimSpace(parts[1]),
			})
		}
	}
	return printers
}
