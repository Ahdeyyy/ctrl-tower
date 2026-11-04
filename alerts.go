package main

import "time"

// AlertKind tells the alert before an occurrence from the one as it
// starts.
type AlertKind int

const (
	AlertLead  AlertKind = iota // minutes before: a notification
	AlertStart                  // at the start: the overlay for events, a notification for reminders
)

type Alert struct {
	Kind AlertKind
	Occ  Occurrence
	At   time.Time
}

// ID names the alert in Data.Fired.
func (a Alert) ID() string {
	if a.Kind == AlertLead {
		return a.Occ.ID() + "#lead"
	}
	return a.Occ.ID() + "#start"
}

// missedWindow is how late an alert still shows, after the computer slept
// or the app was closed: a meeting that started ten minutes ago is still
// worth joining; one that ended this morning is not.
const missedWindow = 15 * time.Minute

// dueAlerts returns the alerts due in (since, now] that have not fired, in
// order. A lead alert whose start alert is also due is dropped: the
// overlay says it all.
func dueAlerts(items []*Item, since, now time.Time, fired func(string) bool) []Alert {
	if floor := now.Add(-missedWindow); since.Before(floor) {
		since = floor
	}
	occs := Occurrences(items, since.Add(-time.Hour), now.Add(25*time.Hour))
	var out []Alert
	for _, o := range occs {
		start := Alert{Kind: AlertStart, Occ: o, At: o.Start}
		startDue := o.Start.After(since) && !o.Start.After(now)
		if lt, ok := o.LeadTime(); ok && lt.After(since) && !lt.After(now) && !startDue {
			a := Alert{Kind: AlertLead, Occ: o, At: lt}
			if !fired(a.ID()) && !fired(start.ID()) {
				out = append(out, a)
			}
		}
		if startDue && !fired(start.ID()) {
			out = append(out, start)
		}
	}
	return out
}
