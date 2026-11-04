package main

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// hour24 writes times as 15:04 rather than 3:04 PM, following the
// system's regional formats (set as the app starts).
var hour24 = false

func clock(t time.Time) string {
	if hour24 {
		return t.Format("15:04")
	}
	return t.Format("3:04 PM")
}

// span writes the times of an occurrence: "3:00 – 3:45 PM", or the start
// alone for a reminder.
func span(o Occurrence) string {
	if !o.End.After(o.Start) {
		return clock(o.Start)
	}
	if !hour24 && o.Start.Format("PM") == o.End.Format("PM") && sameDay(o.Start, o.End) {
		return o.Start.Format("3:04") + " – " + clock(o.End)
	}
	end := clock(o.End)
	if !sameDay(o.Start, o.End) {
		end = o.End.Format("Mon ") + end
	}
	return clock(o.Start) + " – " + end
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

// dayLabel names a day relative to now: Today, Tomorrow, Yesterday, a
// weekday within the week ahead, or a date.
func dayLabel(t, now time.Time) string {
	d := int(math.Round(startOfDay(t).Sub(startOfDay(now)).Hours() / 24))
	switch {
	case d == 0:
		return "Today"
	case d == 1:
		return "Tomorrow"
	case d == -1:
		return "Yesterday"
	case d > 1 && d < 7:
		return t.Format("Monday")
	case t.Year() == now.Year():
		return t.Format("Mon 2 Jan")
	}
	return t.Format("Mon 2 Jan 2006")
}

// minutesLabel writes a length: "45 min", "1 h", "1 h 30 min".
func minutesLabel(m int) string {
	if m < 60 {
		return fmt.Sprintf("%d min", m)
	}
	if m%60 == 0 {
		return fmt.Sprintf("%d h", m/60)
	}
	return fmt.Sprintf("%d h %d min", m/60, m%60)
}

// until writes how far away t is: "now", "in 5 min", "in 2 h 10 min",
// "5 min ago".
func until(t, now time.Time) string {
	d := t.Sub(now)
	if d > -time.Minute && d < time.Minute {
		if d >= 0 {
			return "in <1 min"
		}
		return "now"
	}
	// Rounded toward the event, so "in 1 min" never shows once it started.
	m := int(math.Ceil(d.Minutes()))
	if d < 0 {
		m = int(-d.Minutes())
	}
	var s string
	switch {
	case m >= 48*60:
		s = fmt.Sprintf("%d days", m/(24*60))
	case m >= 60:
		s = minutesLabel(m)
	default:
		s = fmt.Sprintf("%d min", m)
	}
	if d < 0 {
		return s + " ago"
	}
	return "in " + s
}

// countdown writes a timer's time left as 24:59, or 1:04:59 past an hour.
func countdown(d time.Duration) string {
	s := int(math.Ceil(d.Seconds()))
	if s < 0 {
		s = 0
	}
	if s >= 3600 {
		return fmt.Sprintf("%d:%02d:%02d", s/3600, s/60%60, s%60)
	}
	return fmt.Sprintf("%02d:%02d", s/60, s%60)
}

// plural writes a count with its noun, "1 session", "3 sessions".
func plural(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return fmt.Sprintf("%s %s", thousands(n), many)
}

// thousands writes 1284 as 1,284.
func thousands(n int) string {
	s := fmt.Sprint(n)
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	var b strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(r)
	}
	if neg {
		return "-" + b.String()
	}
	return b.String()
}

// hoursLabel writes focus time: "0 min", "45 min", "3 h 20 min".
func hoursLabel(minutes float64) string {
	return minutesLabel(int(minutes + 0.5))
}

func repeatLabel(r Recurrence, start time.Time) string {
	switch r.Freq {
	case FreqDaily:
		return "Every day"
	case FreqWeekdays:
		return "Every weekday"
	case FreqWeekly:
		days := r.Days
		if len(days) == 0 {
			return "Every " + start.Weekday().String()
		}
		if len(days) == 7 {
			return "Every day"
		}
		names := make([]string, len(days))
		for i, d := range sortedDays(days) {
			names[i] = d.String()[:3]
		}
		return "Every " + strings.Join(names, ", ")
	case FreqMonthly:
		return fmt.Sprintf("Monthly on the %s", ordinal(start.Day()))
	}
	return ""
}

func sortedDays(days []time.Weekday) []time.Weekday {
	// Monday first, as most calendars outside the US and as weekdays read.
	out := make([]time.Weekday, 0, len(days))
	for _, d := range []time.Weekday{1, 2, 3, 4, 5, 6, 0} {
		for _, x := range days {
			if x == d {
				out = append(out, d)
				break
			}
		}
	}
	return out
}

func ordinal(n int) string {
	suffix := "th"
	if n%100 < 11 || n%100 > 13 {
		switch n % 10 {
		case 1:
			suffix = "st"
		case 2:
			suffix = "nd"
		case 3:
			suffix = "rd"
		}
	}
	return fmt.Sprintf("%d%s", n, suffix)
}

// hostOf shortens a link for display: "zoom.us/j/123…".
func hostOf(link string) string {
	s := strings.TrimPrefix(strings.TrimPrefix(link, "https://"), "http://")
	return strings.TrimSuffix(s, "/")
}
