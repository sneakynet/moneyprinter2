package cdr

import (
	"encoding/csv"
	"io"
	"log/slog"
	"time"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

const (
	// CiscoLegTypeUnknown is used when the underlying type cannot
	// be parsed from the CDR as supplied by the switch.
	CiscoLegTypeUnknown CiscoLegType = iota

	// CiscoLegTypeTelephony covers most kinds of telephony that
	// may pass through the system.
	CiscoLegTypeTelephony

	// CiscoLegTypeVoIP is used when a call is being transfered
	// via SIP or H.323.
	CiscoLegTypeVoIP

	// CiscoLegTypeMMoIP is used when transporting Fax or other
	// similar kinds of static multimedia mail over an IP channel.
	CiscoLegTypeMMoIP

	// CiscoLegTypeFrameRelay is set when the call channel
	// transits a frame relay connection.
	CiscoLegTypeFrameRelay

	// CiscoLegTypeATM is used when the call makes use of an
	// Asynchronous Transfer Mode channel.
	CiscoLegTypeATM

	// CiscoTime is specified as "NTP Time" with the format "hour,
	// minutes, seconds, microseconds, time_zone, day, month,
	// day_of_month, year."
	//
	// Example: "*19:16:54.886 UTC Fri Jun 14 2024"
	CiscoTime = "*15:04:05.000 MST Mon Jan _2 2006"
)

// CiscoLegType identifies the record type that's been provided.
type CiscoLegType int

// CiscoCDR parses the Compact Format as defined in
// https://www.cisco.com/c/en/us/td/docs/ios/voice/cdr/developer/manual/cdrdev/cdrcsv.html.
type CiscoCDR struct {
	// System time stamp when CDR is captured.
	UnixTime time.Time

	// Value of the Call-ID header.
	CallID uint

	// Template used:
	// 0=None
	// 1=Call history detail
	// 2=Custom template
	Type uint

	// Call leg type:
	// 1=Telephony
	// 2=VoIP
	// 3=MMOIP
	// 4=Frame Relay
	// 5=ATM
	LegType CiscoLegType

	// Unique call identifier generated by the gateway. Used to
	// identify the separate billable events (calls) within a
	// single calling session.
	H323ConfID string

	// Number that this call was connected to in E.164 format.
	PeerAddress string

	// Subaddress configured under a dial peer.
	PeerSubAddress string

	// Setup time, provided in CDR in NTP format.
	H323SetupTime time.Time

	// Time at which call is alerting.
	AlertTime time.Time

	// Connect time, provided in CDR in NTP format.
	H323ConnectTime time.Time

	// Disconnect time, provided in CDR in NTP format.
	H323DisconnectTime time.Time

	// Q.931 disconnect cause code retrieved from Cisco IOS
	// call-control application programming interface (Cisco IOS
	// CCAPI).
	H323DisconnectCause string

	// ASCII text describing the reason for call termination.
	DisconnectText string

	// Gateway’s behavior in relation to the connection that is active for this leg.
	//
	// answer = Legs 1 and 3
	// originate = Legs 2 and 4
	// callback = Legs 1 and 3
	H323CallOrigin string

	// Number of charged units for this connection. For incoming
	// calls or if charging information is not supplied by the
	// switch, the value is zero.
	ChargedUnits uint

	// Type of information carried by media.
	//
	// 1=Other 9 not described
	// 2=Speech
	// 3=UnrestrictedDigital
	// 4=UnrestrictedDigital56
	// 5=RestrictedDigital 6- audio31
	// 7=Audio7
	// 8=Video
	// 9=PacketSwitched
	InfoType string

	// Total number of transmitted packets.
	PacketsOut uint

	// Total number of received packets.
	PacketsIn uint

	// Total number of transmitted bytes.
	BytesOut uint

	// Total number of received bytes.
	BytesIn uint

	// Username for authentication. Usually this is the same as
	// the calling number.
	Username string

	// Calling number.
	CLID string

	// Called number.
	DNIS string
}

// Cisco implements the Parser interface and resolves the Cisco
// Compact format.
type Cisco struct{}

// Parse will read records from the provided reader and will
// return a slice of strongly typed Cisco-style CDRs.
func (c *Cisco) Parse(r io.Reader, clli string) ([]types.CDR, error) {
	cReader := csv.NewReader(r)
	out := []types.CDR{}

	for {
		record, err := cReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.Warn("Error reading CSV", "error", err)
			if len(record) == 1 {
				continue
			}
			return nil, err
		}

		// Discard CDRs that don't have at least one end
		// connected.
		if record[21] == "" || record[22] == "" {
			continue
		}

		cdr := CiscoCDR{
			UnixTime:            time.Unix(int64(strToUint(record[0])), 0),
			CallID:              strToUint(record[1]),
			Type:                strToUint(record[2]),
			LegType:             CiscoLegType(strToUint(record[3])),
			H323ConfID:          record[4],
			PeerAddress:         record[5],
			PeerSubAddress:      record[6],
			H323SetupTime:       c.strCiscoTimeToTime(record[7]),
			AlertTime:           c.strCiscoTimeToTime(record[8]),
			H323ConnectTime:     c.strCiscoTimeToTime(record[9]),
			H323DisconnectTime:  c.strCiscoTimeToTime(record[10]),
			H323DisconnectCause: record[11],
			DisconnectText:      record[12],
			H323CallOrigin:      record[13],
			ChargedUnits:        strToUint(record[14]),
			InfoType:            record[15],
			PacketsOut:          strToUint(record[16]),
			BytesOut:            strToUint(record[17]),
			PacketsIn:           strToUint(record[18]),
			BytesIn:             strToUint(record[19]),
			Username:            record[20],
			CLID:                record[21],
			DNIS:                record[22],
		}

		if cdr.H323DisconnectTime.Sub(cdr.H323ConnectTime) == 0 {
			continue
		}

		slog.Debug("Original Cisco CDR", "data", cdr)

		out = append(out, types.CDR{
			CLLI:    clli,
			OrigID:  cdr.CallID,
			LogTime: cdr.UnixTime,
			CLID:    cdr.CLID,
			DNIS:    cdr.DNIS,
			Start:   cdr.H323SetupTime,
			End:     cdr.H323DisconnectTime,
		})
	}

	return out, nil
}

func (c *Cisco) strCiscoTimeToTime(s string) time.Time {
	if s == "" {
		slog.Warn("Tried to parse empty time!")
		return time.Time{}
	}

	t, err := time.Parse(CiscoTime, s)
	if err != nil {
		slog.Warn("Failed to parse time as CiscoTime", "time", s, "error", err)
		return time.Time{}
	}
	return t
}
