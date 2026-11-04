package main

import (
	"sort"
	"time"
)

// Phase is what the pomodoro timer is timing.
type Phase int

const (
	PhaseWork Phase = iota
	PhaseBreak
	PhaseLongBreak
)

func (p Phase) String() string {
	switch p {
	case PhaseBreak:
		return "Short break"
	case PhaseLongBreak:
		return "Long break"
	}
	return "Focus"
}

// Timer is the pomodoro timer. It keeps no clock of its own: every method
// takes the time, so tests and the UI agree on it.
type Timer struct {
	Phase   Phase
	Running bool
	EndsAt  time.Time     // while running
	Left    time.Duration // while not running
	Length  time.Duration // of the current phase, which the ring fills
	Done    int           // work sessions finished in this cycle
	Tag     string

	// Work time of the current session, for the session recorded when it
	// ends early.
	started  time.Time
	worked   time.Duration
	runSince time.Time
}

func minutes(n int) time.Duration { return time.Duration(n) * time.Minute }

func NewTimer(p Preset) Timer {
	return Timer{Phase: PhaseWork, Length: minutes(p.Work), Left: minutes(p.Work)}
}

// Fresh reports whether the current phase has not run yet.
func (t *Timer) Fresh() bool { return !t.Running && t.Left == t.Length }

func (t *Timer) Remaining(now time.Time) time.Duration {
	if !t.Running {
		return t.Left
	}
	return max(0, t.EndsAt.Sub(now))
}

// Progress is how much of the phase has gone, from 0 to 1.
func (t *Timer) Progress(now time.Time) float64 {
	if t.Length <= 0 {
		return 0
	}
	return 1 - float64(t.Remaining(now))/float64(t.Length)
}

func (t *Timer) Start(now time.Time) {
	if t.Running {
		return
	}
	if t.Phase == PhaseWork && t.started.IsZero() {
		t.started = now
	}
	t.Running = true
	t.runSince = now
	t.EndsAt = now.Add(t.Left)
}

func (t *Timer) Pause(now time.Time) {
	if !t.Running {
		return
	}
	t.Left = t.Remaining(now)
	t.Running = false
	if t.Phase == PhaseWork {
		t.worked += now.Sub(t.runSince)
	}
}

// Reset puts the timer back at the start of a work session, forgetting the
// cycle.
func (t *Timer) Reset(p Preset) {
	tag := t.Tag
	*t = NewTimer(p)
	t.Tag = tag
}

// SetLength changes the length of the phase under way, as when a session
// is shortened to end before a meeting. The time already gone stays gone.
func (t *Timer) SetLength(now time.Time, d time.Duration) {
	gone := t.Length - t.Remaining(now)
	t.Length = max(d, gone)
	if t.Running {
		t.EndsAt = now.Add(t.Length - gone)
	} else {
		t.Left = t.Length - gone
	}
}

// Transition is what Tick and Skip report as a phase ends.
type Transition struct {
	From, To Phase
	Session  *Session // the work session that ended, if any
}

// Tick ends the phase if its time is up, and returns what happened.
func (t *Timer) Tick(now time.Time, p Preset, auto bool) (Transition, bool) {
	if !t.Running || now.Before(t.EndsAt) {
		return Transition{}, false
	}
	return t.advance(t.EndsAt, now, p, auto, true), true
}

// Skip ends the phase now. A work session ended early is recorded when it
// ran at least a minute.
func (t *Timer) Skip(now time.Time, p Preset, auto bool) Transition {
	return t.advance(now, now, p, auto, false)
}

func (t *Timer) advance(ended, now time.Time, p Preset, auto, full bool) Transition {
	tr := Transition{From: t.Phase}
	if t.Phase == PhaseWork {
		worked := t.worked
		if t.Running {
			worked += ended.Sub(t.runSince)
		}
		if worked >= time.Minute {
			start := t.started
			if start.IsZero() {
				start = ended.Add(-worked)
			}
			tr.Session = &Session{Start: start, Minutes: worked.Minutes(), Tag: t.Tag, Preset: p.Name, Full: full}
		}
		if full || worked >= time.Minute {
			t.Done++
		}
		if t.Done >= p.LongEvery {
			t.Phase = PhaseLongBreak
		} else {
			t.Phase = PhaseBreak
		}
	} else {
		if t.Phase == PhaseLongBreak {
			t.Done = 0
		}
		t.Phase = PhaseWork
	}
	t.started, t.worked = time.Time{}, 0
	switch t.Phase {
	case PhaseWork:
		t.Length = minutes(p.Work)
	case PhaseBreak:
		t.Length = minutes(p.Break)
	case PhaseLongBreak:
		t.Length = minutes(p.LongBreak)
	}
	tr.To = t.Phase
	wasRunning := t.Running
	t.Running, t.Left = false, t.Length
	if auto && wasRunning {
		// Chained from when the phase ended, so a late tick costs no time,
		// unless the computer slept through the whole next phase.
		start := ended
		if now.Sub(ended) > t.Length {
			start = now
		}
		t.Start(start)
	}
	return tr
}

// meetingConflict finds the first event that starts before a session
// ending at ends would, with two minutes to spare, and how long the
// session can run to end that much before it.
func meetingConflict(now, ends time.Time, occs []Occurrence) (Occurrence, time.Duration, bool) {
	const spare = 2 * time.Minute
	for _, o := range occs {
		if !o.IsEvent() || !o.Start.After(now) {
			continue
		}
		if o.Start.Before(ends.Add(spare)) {
			return o, max(0, o.Start.Sub(now)-spare).Truncate(time.Minute), true
		}
	}
	return Occurrence{}, 0, false
}

// Focus totals for the stats.

// minutesOnDays returns the minutes of focus on each of the n days ending
// with the day of now, oldest first.
func minutesOnDays(sessions []Session, now time.Time, n int) []float64 {
	out := make([]float64, n)
	first := startOfDay(now).AddDate(0, 0, -(n - 1))
	for _, s := range sessions {
		start := s.Start.In(now.Location())
		if start.Before(first) {
			continue
		}
		i := int(startOfDay(start).Sub(first).Hours()/24 + 0.5)
		if i >= 0 && i < n {
			out[i] += s.Minutes
		}
	}
	return out
}

type TagTotal struct {
	Tag     string
	Minutes float64
	Count   int
}

// minutesByTag totals sessions since from by tag, largest first.
func minutesByTag(sessions []Session, from time.Time) []TagTotal {
	byTag := map[string]*TagTotal{}
	var out []*TagTotal
	for _, s := range sessions {
		if s.Start.Before(from) {
			continue
		}
		tt := byTag[s.Tag]
		if tt == nil {
			tt = &TagTotal{Tag: s.Tag}
			byTag[s.Tag] = tt
			out = append(out, tt)
		}
		tt.Minutes += s.Minutes
		tt.Count++
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Minutes > out[j].Minutes })
	res := make([]TagTotal, len(out))
	for i, t := range out {
		res[i] = *t
	}
	return res
}

// streak counts the days in a row, up to today, with any focus.
func streak(sessions []Session, now time.Time) int {
	days := minutesOnDays(sessions, now, 366)
	n := 0
	for i := len(days) - 1; i >= 0; i-- {
		if days[i] == 0 {
			if i == len(days)-1 {
				continue // today may not have started yet
			}
			break
		}
		n++
	}
	return n
}
