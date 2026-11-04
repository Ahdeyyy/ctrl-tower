package main

import (
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"strings"
	"time"
)

// Kind tells an event, which has a length and a place, from a plain
// reminder, which is a nudge at a time.
type Kind string

const (
	KindEvent    Kind = "event"
	KindReminder Kind = "reminder"
)

// Freq is how often an item repeats.
type Freq string

const (
	FreqNone     Freq = ""
	FreqDaily    Freq = "daily"
	FreqWeekdays Freq = "weekdays"
	FreqWeekly   Freq = "weekly"
	FreqMonthly  Freq = "monthly"
)

// Recurrence repeats an item from its start. Weekly items repeat on Days,
// or on the start's weekday when Days is empty; monthly ones on the
// start's day of the month, or the month's last day when it is shorter.
type Recurrence struct {
	Freq  Freq           `json:"freq,omitempty"`
	Days  []time.Weekday `json:"days,omitempty"`
	Until time.Time      `json:"until,omitzero"` // last day included; zero repeats forever
}

// Override changes one occurrence of a repeating item, keyed by the date
// it would have fallen on.
type Override struct {
	Skip     bool      `json:"skip,omitempty"`
	Start    time.Time `json:"start,omitzero"`
	Minutes  int       `json:"minutes,omitempty"`
	Title    string    `json:"title,omitempty"`
	Location string    `json:"location,omitempty"`
	URL      string    `json:"url,omitempty"`
	Notes    string    `json:"notes,omitempty"`
}

// Item is an event or a reminder, repeating or not.
type Item struct {
	ID       string    `json:"id"`
	Kind     Kind      `json:"kind"`
	Title    string    `json:"title"`
	Start    time.Time `json:"start"`
	Minutes  int       `json:"minutes,omitempty"` // events only
	Location string    `json:"location,omitempty"`
	URL      string    `json:"url,omitempty"`
	Notes    string    `json:"notes,omitempty"`
	// Lead is how many minutes before the start the first alert comes;
	// negative for none.
	Lead      int                 `json:"lead"`
	Repeat    Recurrence          `json:"repeat,omitzero"`
	Overrides map[string]Override `json:"overrides,omitempty"`
	Created   time.Time           `json:"created"`
}

// Limits of what the editor and quick-add accept. Imports and the JSON
// file are clamped to them on load.
const (
	maxTitle    = 200
	maxLocation = 300
	maxURL      = 2048
	maxNotes    = 5000
	maxTag      = 60
	maxMinutes  = 24 * 60
)

func newID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// Occurrence is an item on one day: the item itself, or one repetition of
// it with its override applied.
type Occurrence struct {
	Item     *Item
	Key      string // the date the series put it on, 2006-01-02
	Start    time.Time
	End      time.Time
	Title    string
	Location string
	URL      string
	Notes    string
}

// ID names the occurrence across runs, for alerts already shown.
func (o Occurrence) ID() string { return o.Item.ID + "@" + o.Key }

func (o Occurrence) IsEvent() bool { return o.Item.Kind == KindEvent }

// Online reports whether the occurrence has a link to join.
func (o Occurrence) Online() bool { return o.URL != "" }

// LeadTime is when the first alert for the occurrence is due, and whether
// it has one.
func (o Occurrence) LeadTime() (time.Time, bool) {
	if o.Item.Lead <= 0 {
		return time.Time{}, false
	}
	return o.Start.Add(-time.Duration(o.Item.Lead) * time.Minute), true
}

func dateKey(t time.Time) string { return t.Format("2006-01-02") }

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// at puts the time of day of clock on the date of day, in day's location,
// so a 9:00 standup stays at 9:00 across daylight saving changes.
func at(day, clock time.Time) time.Time {
	y, m, d := day.Date()
	clock = clock.In(day.Location())
	return time.Date(y, m, d, clock.Hour(), clock.Minute(), 0, 0, day.Location())
}

func daysIn(y int, m time.Month) int {
	return time.Date(y, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// matches reports whether a repeating item falls on day (a midnight).
func (it *Item) matches(day time.Time) bool {
	first := startOfDay(it.Start.In(day.Location()))
	if day.Before(first) {
		return false
	}
	if !it.Repeat.Until.IsZero() && day.After(startOfDay(it.Repeat.Until.In(day.Location()))) {
		return false
	}
	switch it.Repeat.Freq {
	case FreqDaily:
		return true
	case FreqWeekdays:
		wd := day.Weekday()
		return wd != time.Saturday && wd != time.Sunday
	case FreqWeekly:
		days := it.Repeat.Days
		if len(days) == 0 {
			return day.Weekday() == first.Weekday()
		}
		for _, d := range days {
			if d == day.Weekday() {
				return true
			}
		}
		return false
	case FreqMonthly:
		want := first.Day()
		if last := daysIn(day.Year(), day.Month()); want > last {
			want = last
		}
		return day.Day() == want
	}
	return day.Equal(first)
}

func (it *Item) length() time.Duration {
	if it.Kind != KindEvent {
		return 0
	}
	return time.Duration(it.Minutes) * time.Minute
}

func (it *Item) occurrence(key string, start time.Time) (Occurrence, bool) {
	o := Occurrence{
		Item: it, Key: key, Start: start, End: start.Add(it.length()),
		Title: it.Title, Location: it.Location, URL: it.URL, Notes: it.Notes,
	}
	ov, ok := it.Overrides[key]
	if !ok {
		return o, true
	}
	if ov.Skip {
		return o, false
	}
	if !ov.Start.IsZero() {
		o.Start = ov.Start
		o.End = ov.Start.Add(it.length())
	}
	if ov.Minutes > 0 && it.Kind == KindEvent {
		o.End = o.Start.Add(time.Duration(ov.Minutes) * time.Minute)
	}
	if ov.Title != "" {
		o.Title = ov.Title
	}
	if ov.Location != "" {
		o.Location = ov.Location
	}
	if ov.URL != "" {
		o.URL = ov.URL
	}
	if ov.Notes != "" {
		o.Notes = ov.Notes
	}
	return o, true
}

// Occurrences returns the occurrences of items that overlap [from, to),
// sorted by start, with reminders and zero-length items included when
// they start inside it.
func Occurrences(items []*Item, from, to time.Time) []Occurrence {
	var out []Occurrence
	loc := from.Location()
	for _, it := range items {
		if it.Repeat.Freq == FreqNone {
			o, ok := it.occurrence(dateKey(it.Start.In(loc)), it.Start)
			if ok && overlaps(o, from, to) {
				out = append(out, o)
			}
			continue
		}
		// Occurrences moved by an override are found by their overrides,
		// wherever they moved to; the others on their days.
		day := startOfDay(from).AddDate(0, 0, -1)
		if d := it.length(); d > 24*time.Hour {
			day = day.Add(-d)
		}
		for ; day.Before(to); day = day.AddDate(0, 0, 1) {
			key := dateKey(day)
			if !it.matches(day) || !it.Overrides[key].Start.IsZero() {
				continue
			}
			o, ok := it.occurrence(key, at(day, it.Start))
			if ok && overlaps(o, from, to) {
				out = append(out, o)
			}
		}
		for key, ov := range it.Overrides {
			if ov.Start.IsZero() || ov.Skip {
				continue
			}
			day, err := time.ParseInLocation("2006-01-02", key, loc)
			if err != nil || !it.matches(day) {
				continue
			}
			o, _ := it.occurrence(key, at(day, it.Start))
			if overlaps(o, from, to) {
				out = append(out, o)
			}
		}
	}
	sortOccurrences(out)
	return out
}

func overlaps(o Occurrence, from, to time.Time) bool {
	if !o.Start.Before(to) {
		return false
	}
	if o.End.After(o.Start) {
		return o.End.After(from)
	}
	return !o.Start.Before(from)
}

func sortOccurrences(os []Occurrence) {
	// Insertion sort keeps equal starts in item order; lists are short.
	for i := 1; i < len(os); i++ {
		for j := i; j > 0 && less(os[j], os[j-1]); j-- {
			os[j], os[j-1] = os[j-1], os[j]
		}
	}
}

func less(a, b Occurrence) bool {
	if !a.Start.Equal(b.Start) {
		return a.Start.Before(b.Start)
	}
	if a.Item.Kind != b.Item.Kind {
		return a.Item.Kind == KindEvent
	}
	return a.Title < b.Title
}

// meetingHosts are the services whose links quick-add and the editor take
// without a scheme.
var meetingHosts = []string{
	"zoom.us", "meet.google.com", "teams.microsoft.com", "teams.live.com",
	"webex.com", "whereby.com", "meet.jit.si", "gather.town", "discord.gg",
	"discord.com", "slack.com", "around.co", "chime.aws",
}

// normalizeURL makes a link the user typed openable: it adds https:// to a
// bare host, and returns "" for anything that is not an http(s) link.
func normalizeURL(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if !strings.Contains(s, "://") {
		s = "https://" + s
	}
	u, err := url.Parse(s)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || !strings.Contains(u.Host, ".") {
		return ""
	}
	return u.String()
}

// meetingService names the service of a meeting link, for its button.
func meetingService(link string) string {
	u, err := url.Parse(link)
	if err != nil {
		return ""
	}
	h := strings.ToLower(u.Hostname())
	switch {
	case strings.HasSuffix(h, "zoom.us"):
		return "Zoom"
	case h == "meet.google.com":
		return "Meet"
	case strings.Contains(h, "teams."):
		return "Teams"
	case strings.HasSuffix(h, "webex.com"):
		return "Webex"
	case strings.Contains(h, "discord"):
		return "Discord"
	case strings.HasSuffix(h, "slack.com"):
		return "Slack"
	}
	return ""
}

// clampText cuts s to n runes, for values loaded from disk.
func clampText(s string, n int) string {
	if len(s) <= n {
		return s
	}
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

func (it *Item) clamp() {
	it.Title = clampText(strings.TrimSpace(it.Title), maxTitle)
	it.Location = clampText(it.Location, maxLocation)
	it.URL = clampText(it.URL, maxURL)
	it.Notes = clampText(it.Notes, maxNotes)
	if it.Minutes < 0 {
		it.Minutes = 0
	}
	if it.Minutes > maxMinutes {
		it.Minutes = maxMinutes
	}
	if it.Kind != KindReminder {
		it.Kind = KindEvent
	}
}
