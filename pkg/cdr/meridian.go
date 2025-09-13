package cdr

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/cespare/xxhash/v2"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

// MeridianRecordType aliases the record types defined in NN43001
// "Call Detail Recording Fundamentals".
type MeridianRecordType rune

const (
	// MeridianRecordTypeNormal is generated when a simple call is
	// established, whether or not it is extended through the
	// Attendant Console, and when no other phone feature is
	// activated.
	MeridianRecordTypeNormal MeridianRecordType = 'N'

	// MeridianRecordTypeInternal is produced when the Internal
	// CDR criteria are satisfied. At least one L record is
	// produced when internal calls are modified, forwarded, or
	// transferred.
	MeridianRecordTypeInternal MeridianRecordType = 'L'
)

var (
	meridianRegexpIsTrunk = regexp.MustCompile(`[AT]\d+`)

	meridianRegexpNormal = regexp.MustCompile(`([A-N])\s(\d{3})\s(\d{2})\s+([\dA-Z]+)\s+([\dA-Z]+)\s+(\d+\/\d+\s\d{2}:\d{2}:\d{2})\s(\d{2}:\d{2}:\d{2})`)

	meridianRegexpNormalDigits = regexp.MustCompile(`([A-N])\s(\d{3})\s(\d{2})\s+([\dA-Z]+)\s+([\dA-Z]+)\s+(\d+\/\d+\s\d{2}:\d{2}:\d{2})\s(\d{2}:\d{2}:\d{2})\.\d\s+([A-Z\d]+)`)

	meridianRegexpInternal = regexp.MustCompile(`([A-N])\s+(\d{3})\s+(\d{2})\s+([\dA-Z]+)\s+([\dA-Z]+)\s+(\d+\/\d+\s\d{2}:\d{2}:\d{2})\s(\d{2}:\d{2}:\d{2})`)
)

// The Meridian TTY output can have NULL characters at arbitrary
// positions within the byte stream, these need to be removed prior to
// parsing.
type meridianNullRemover struct{ io.Reader }

// Read satisfies the io.Reader interface
func (m *meridianNullRemover) Read(p []byte) (int, error) {
	n, err := m.Reader.Read(p)
	if err != nil {
		return n, err
	}

	for i := 0; i < n; i++ {
		if p[i] == 0 {
			copy(p[i:], p[i+1:])
			n--
		} else if i > 0 && p[i-1] == 0 {
			i++
		}
	}

	return n, err
}

// Meridian binds all methods for the parser
type Meridian struct{}

// Parse reads in the arbitrary data from the meridian CDRs then
// returns a set of records.
func (m *Meridian) Parse(r io.Reader, clli string) ([]types.CDR, error) {
	scanner := bufio.NewScanner(bufio.NewReader(&meridianNullRemover{r}))
	out := []types.CDR{}
	for scanner.Scan() {
		line := scanner.Text()
		cdr := types.CDR{CLLI: clli}

		if len(line) < 1 {
			continue
		}
		switch MeridianRecordType(line[0]) {
		case MeridianRecordTypeNormal:
			matches := meridianRegexpNormal.FindStringSubmatch(line)
			if len(matches) < 8 {
				slog.Warn("Dropping malformed CDR", "cdr", line, "len", len(matches))
				continue
			}
			cdr.OrigID = xxhash.Sum64String(line)
			cdr.LogTime = m.parseDate(matches[7])
			cdr.Start = m.parseDate(matches[6])
			cdr.End = m.parseDate(matches[6]).Add(m.parseDur(matches[7]))
			cdr.CLID = matches[4]

			// If the call is destined for a trunk then
			// the called number is actually in the DIGITS
			// field, which results in a re-parse of the
			// entire CDR to validate that it is what we
			// expected it to be.
			if meridianRegexpIsTrunk.MatchString(matches[5]) {
				digits := meridianRegexpNormalDigits.FindStringSubmatch(line)
				if len(digits) != 9 {
					slog.Warn("TERID is trunk but DIGITS failed to parse!", "line", line)
					continue
				}
				cdr.DNIS = digits[len(digits)-1]

				// Did this call go through ARS?
				if cdr.DNIS[0] == 'A' {
					cdr.DNIS = m.doARSTranslations(cdr.DNIS[1:])
				}
			} else {
				// Not a trunk call, but somehow still
				// resulted in an N record, likely a
				// trunk origin, and the CDR for that
				// was generated somewhere else).
				cdr.DNIS = matches[5]
			}

		case MeridianRecordTypeInternal:
			matches := meridianRegexpInternal.FindStringSubmatch(line)
			if len(matches) != 8 {
				slog.Warn("Dropping malformed CDR", "cdr", line)
				continue
			}
			cdr.OrigID = xxhash.Sum64String(line)
			cdr.LogTime = m.parseDate(matches[5])
			cdr.CLID = matches[4]
			cdr.DNIS = matches[5]
			cdr.Start = m.parseDate(matches[5]).Add(m.parseDur(matches[7]) * -1)
			cdr.End = m.parseDate(matches[5])

		default:
			continue
		}
		out = append(out, cdr)
	}

	return out, nil
}

func (m *Meridian) parseDate(d string) time.Time {
	timestamp, _ := time.Parse("2006 01/02 15:04:05", fmt.Sprintf("%d %s", time.Now().Year(), d))
	return timestamp
}

func (m *Meridian) parseDur(d string) time.Duration {
	durTS, err := time.Parse("15:04:05", d)
	if err != nil {
		slog.Warn("Error parsing duration", "error", err)
	}
	h, min, s := durTS.Clock()
	dur, err := time.ParseDuration(fmt.Sprintf("%dh%dm%ds", h, min, s))
	if err != nil {
		slog.Warn("Error parsing duration", "error", err)
	}
	return dur
}

func (m *Meridian) doARSTranslations(s string) string {
	ret := s

	t := os.Getenv("M1_ARS_TRANSLATIONS")
	if t == "" {
		return s
	}
	translations := strings.Split(t, ";")
	for _, tr := range translations {
		parts := strings.Split(tr, "|")
		r := regexp.MustCompile(parts[0])
		ret = r.ReplaceAllLiteralString(s, parts[1])
		if ret != s {
			// A replacement happened
			return ret
		}
	}

	return s
}
