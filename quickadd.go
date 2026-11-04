package main

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Parsed is what quick-add understood from a line of text.
type Parsed struct {
	Item    Item
	HasDate bool
	HasTime bool
}

var (
	reClock    = regexp.MustCompile(`^(\d{1,2})(?:[:.](\d{2}))?(am|pm|a|p)?$`)
	reDuration = regexp.MustCompile(`^(\d+(?:\.\d+)?)(m|min|mins|minute|minutes|h|hr|hrs|hour|hours)$`)
	reHourMin  = regexp.MustCompile(`^(\d+)h(\d+)(?:m|min)?$`)
	reISODate  = regexp.MustCompile(`^(\d{4})-(\d{1,2})-(\d{1,2})$`)
	reOrdinal  = regexp.MustCompile(`^(\d{1,2})(?:st|nd|rd|th)?,?$`)
)

var weekdayNames = map[string]time.Weekday{
	"sun": time.Sunday, "sunday": time.Sunday,
	"mon": time.Monday, "monday": time.Monday,
	"tue": time.Tuesday, "tues": time.Tuesday, "tuesday": time.Tuesday,
	"wed": time.Wednesday, "weds": time.Wednesday, "wednesday": time.Wednesday,
	"thu": time.Thursday, "thur": time.Thursday, "thurs": time.Thursday, "thursday": time.Thursday,
	"fri": time.Friday, "friday": time.Friday,
	"sat": time.Saturday, "saturday": time.Saturday,
}

var monthNames = map[string]time.Month{
	"jan": 1, "january": 1, "feb": 2, "february": 2, "mar": 3, "march": 3,
	"apr": 4, "april": 4, "may": 5, "jun": 6, "june": 6, "jul": 7, "july": 7,
	"aug": 8, "august": 8, "sep": 9, "sept": 9, "september": 9,
	"oct": 10, "october": 10, "nov": 11, "november": 11, "dec": 12, "december": 12,
}

func unitDuration(n float64, unit string) time.Duration {
	if strings.HasPrefix(unit, "h") {
		return time.Duration(n * float64(time.Hour))
	}
	return time.Duration(n * float64(time.Minute))
}

// parseDuration reads "45m", "1h30", "1.5h", or a number and a unit as two
// tokens, returning how many tokens it took.
func parseDuration(toks []string, i int) (time.Duration, int) {
	t := strings.TrimSuffix(toks[i], ",")
	if m := reHourMin.FindStringSubmatch(t); m != nil {
		h, _ := strconv.Atoi(m[1])
		mm, _ := strconv.Atoi(m[2])
		return time.Duration(h)*time.Hour + time.Duration(mm)*time.Minute, 1
	}
	if m := reDuration.FindStringSubmatch(t); m != nil {
		n, _ := strconv.ParseFloat(m[1], 64)
		return unitDuration(n, m[2]), 1
	}
	if i+1 < len(toks) {
		if n, err := strconv.ParseFloat(t, 64); err == nil {
			u := strings.TrimSuffix(toks[i+1], ",")
			if reDuration.MatchString("1" + u) {
				return unitDuration(n, u), 2
			}
		}
	}
	return 0, 0
}

// parseClock reads "3pm", "15:30", "3:30 pm" (two tokens), or a bare hour
// when bare is true, as after "at".
func parseClock(toks []string, i int, bare bool) (h, m, n int, ok bool) {
	t := strings.TrimSuffix(toks[i], ",")
	switch t {
	case "noon", "midday":
		return 12, 0, 1, true
	case "midnight":
		return 0, 0, 1, true
	}
	mt := reClock.FindStringSubmatch(t)
	if mt == nil {
		return 0, 0, 0, false
	}
	h, _ = strconv.Atoi(mt[1])
	if mt[2] != "" {
		m, _ = strconv.Atoi(mt[2])
	}
	n = 1
	suffix := mt[3]
	if suffix == "" && i+1 < len(toks) {
		switch strings.TrimSuffix(toks[i+1], ",") {
		case "am", "pm", "a.m.", "p.m.":
			suffix = toks[i+1][:1]
			n = 2
		}
	}
	if suffix == "" && mt[2] == "" && !bare {
		return 0, 0, 0, false
	}
	if h > 23 || m > 59 {
		return 0, 0, 0, false
	}
	switch suffix[:min(1, len(suffix))] {
	case "p":
		if h > 12 {
			return 0, 0, 0, false
		}
		if h < 12 {
			h += 12
		}
	case "a":
		if h > 12 {
			return 0, 0, 0, false
		}
		if h == 12 {
			h = 0
		}
	default:
		// "at 3" means the afternoon; nobody books a 3 a.m. sync.
		if mt[2] == "" && h >= 1 && h <= 7 {
			h += 12
		}
	}
	return h, m, n, true
}

// ParseQuick reads a line such as "Design sync tomorrow 3pm 45m
// zoom.us/j/123 @ Room 4" into an item, relative to now.
func ParseQuick(s string, now time.Time, defaultLead int) Parsed {
	var p Parsed
	it := &p.Item
	it.Kind = KindEvent
	it.Lead = defaultLead

	s = strings.TrimSpace(s)
	low := strings.ToLower(s)
	for _, prefix := range []string{"remind me to ", "remind me ", "reminder: ", "reminder ", "remind ", "todo: ", "todo ", "! "} {
		if strings.HasPrefix(low, prefix) {
			it.Kind = KindReminder
			s = strings.TrimSpace(s[len(prefix):])
			break
		}
	}

	// The place is everything after an @, and a link anywhere.
	raw := strings.Fields(s)
	var words []string
	var place []string
	inPlace := false
	for _, w := range raw {
		if it.URL == "" && looksLikeURL(w) {
			if u := normalizeURL(strings.TrimRight(w, ".,;)")); u != "" {
				it.URL = u
				continue
			}
		}
		if !inPlace && strings.HasPrefix(w, "@") {
			inPlace = true
			w = strings.TrimPrefix(w, "@")
			if w == "" {
				continue
			}
		}
		if inPlace {
			place = append(place, w)
		} else {
			words = append(words, w)
		}
	}
	it.Location = clampText(strings.Join(place, " "), maxLocation)

	toks := make([]string, len(words))
	for i, w := range words {
		toks[i] = strings.ToLower(w)
	}
	used := make([]bool, len(words))
	use := func(i, n int) {
		for j := i; j < i+n && j < len(used); j++ {
			used[j] = true
		}
	}

	var (
		date      time.Time
		hasDate   bool
		hour, min int
		hasTime   bool
		dur       time.Duration
		relative  time.Duration
		hasRel    bool
		wantNext  bool
		weekday   = time.Weekday(-1)
	)
	today := startOfDay(now)

	for i := 0; i < len(toks); i++ {
		if used[i] {
			continue
		}
		t := strings.TrimSuffix(toks[i], ",")
		next := func(k int) string {
			if i+k < len(toks) {
				return strings.TrimSuffix(toks[i+k], ",")
			}
			return ""
		}
		switch {
		case t == "today":
			date, hasDate = today, true
			use(i, 1)
		case t == "tonight":
			date, hasDate = today, true
			if !hasTime {
				hour, min, hasTime = 20, 0, true
			}
			use(i, 1)
		case t == "tomorrow" || t == "tmr" || t == "tmrw" || t == "tomorow":
			date, hasDate = today.AddDate(0, 0, 1), true
			use(i, 1)
		case t == "morning" || t == "afternoon" || t == "evening":
			if !hasTime {
				hour, min, hasTime = map[string]int{"morning": 9, "afternoon": 14, "evening": 18}[t], 0, true
				use(i, 1)
				if i > 0 && (toks[i-1] == "in" || toks[i-1] == "this") {
					use(i-1, 1)
				}
			}
		case t == "every":
			n := everyRule(toks, i+1, &it.Repeat)
			if n > 0 {
				use(i, n+1)
			}
		case t == "daily":
			it.Repeat.Freq = FreqDaily
			use(i, 1)
		case t == "weekdays":
			it.Repeat.Freq = FreqWeekdays
			use(i, 1)
		case t == "weekly":
			it.Repeat.Freq = FreqWeekly
			use(i, 1)
		case t == "monthly":
			it.Repeat.Freq = FreqMonthly
			use(i, 1)
		case t == "in" && !hasRel:
			if d, n := parseDuration(toks, i+1); n > 0 {
				relative, hasRel = d, true
				use(i, n+1)
			}
		case t == "for":
			if d, n := parseDuration(toks, i+1); n > 0 {
				dur = d
				use(i, n+1)
			}
		case t == "at" || t == "@":
			if h, m, n, ok := parseClock(toks, i+1, true); ok && !hasTime {
				hour, min, hasTime = h, m, true
				use(i, n+1)
			}
		case t == "on" || t == "next" || t == "this":
			if wd, ok := weekdayNames[next(1)]; ok && !hasDate {
				weekday, wantNext = wd, t == "next"
				use(i, 2)
			} else if t == "on" {
				if d, n, ok := parseDate(toks, i+1, today); ok && !hasDate {
					date, hasDate = d, true
					use(i, n+1)
				}
			}
		default:
			if wd, ok := weekdayNames[t]; ok && !hasDate && weekday < 0 {
				weekday = wd
				use(i, 1)
				continue
			}
			if d, n, ok := parseDate(toks, i, today); ok && !hasDate {
				date, hasDate = d, true
				use(i, n)
				continue
			}
			if h, m, n, ok := parseClock(toks, i, false); ok && !hasTime {
				hour, min, hasTime = h, m, true
				use(i, n)
				continue
			}
			if d, n := parseDuration(toks, i); n > 0 && dur == 0 {
				dur = d
				use(i, n)
			}
		}
	}

	if weekday >= 0 && !hasDate {
		ahead := (int(weekday) - int(now.Weekday()) + 7) % 7
		if wantNext && ahead == 0 {
			ahead = 7
		}
		date, hasDate = today.AddDate(0, 0, ahead), true
		if ahead == 0 && hasTime && !time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, now.Location()).After(now) {
			date = date.AddDate(0, 0, 7)
		}
	}

	var start time.Time
	switch {
	case hasRel:
		start = now.Add(relative).Truncate(time.Minute)
		p.HasDate, p.HasTime = true, true
	case hasDate && hasTime:
		start = time.Date(date.Year(), date.Month(), date.Day(), hour, min, 0, 0, now.Location())
	case hasDate:
		start = time.Date(date.Year(), date.Month(), date.Day(), 9, 0, 0, 0, now.Location())
	case hasTime:
		start = time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, now.Location())
		if !start.After(now) && it.Repeat.Freq == FreqNone {
			start = start.AddDate(0, 0, 1)
		}
	default:
		// The next half hour, at least ten minutes away.
		start = now.Add(10 * time.Minute).Truncate(30 * time.Minute).Add(30 * time.Minute)
		if start.Sub(now) > 40*time.Minute {
			start = start.Add(-30 * time.Minute)
		}
	}
	// A repeating item without a date starts on its first matching day.
	if it.Repeat.Freq != FreqNone && !hasDate && !hasRel {
		probe := Item{Start: today, Repeat: it.Repeat}
		for k := 0; k < 8; k++ {
			d := today.AddDate(0, 0, k)
			if cand := at(d, start); probe.matches(d) && cand.After(now) {
				start = cand
				break
			}
		}
	}
	it.Start = start
	p.HasDate = p.HasDate || hasDate
	p.HasTime = p.HasTime || hasTime

	if it.Kind == KindEvent {
		if dur <= 0 {
			dur = 30 * time.Minute
		}
		it.Minutes = clampInt(int(dur.Round(time.Minute)/time.Minute), 1, maxMinutes)
	}

	var title []string
	for i, w := range words {
		if !used[i] {
			title = append(title, w)
		}
	}
	it.Title = cleanTitle(strings.Join(title, " "))
	if it.Title == "" {
		if it.URL != "" {
			if svc := meetingService(it.URL); svc != "" {
				it.Title = svc + " call"
			} else {
				it.Title = "Meeting"
			}
		} else if it.Kind == KindReminder {
			it.Title = "Reminder"
		}
	}
	it.Title = clampText(it.Title, maxTitle)
	return p
}

// everyRule reads what follows "every", returning how many tokens it took.
func everyRule(toks []string, i int, r *Recurrence) int {
	if i >= len(toks) {
		return 0
	}
	t := strings.TrimSuffix(toks[i], ",")
	switch t {
	case "day", "morning", "evening", "night":
		r.Freq = FreqDaily
		return 1
	case "weekday", "weekdays", "workday", "workdays":
		r.Freq = FreqWeekdays
		return 1
	case "week":
		r.Freq = FreqWeekly
		return 1
	case "month":
		r.Freq = FreqMonthly
		return 1
	}
	n := 0
	var days []time.Weekday
	for j := i; j < len(toks); j++ {
		w := strings.TrimSuffix(toks[j], ",")
		if wd, ok := weekdayNames[strings.TrimSuffix(w, "s")]; ok {
			days = append(days, wd)
			n = j - i + 1
			continue
		}
		if wd, ok := weekdayNames[w]; ok {
			days = append(days, wd)
			n = j - i + 1
			continue
		}
		if (w == "and" || w == "&" || w == "/") && len(days) > 0 {
			continue
		}
		break
	}
	if len(days) > 0 {
		r.Freq = FreqWeekly
		r.Days = days
	}
	return n
}

// parseDate reads "2026-11-12", "nov 12", "12 nov" or "november 12th".
func parseDate(toks []string, i int, today time.Time) (time.Time, int, bool) {
	if i >= len(toks) {
		return time.Time{}, 0, false
	}
	t := strings.TrimSuffix(toks[i], ",")
	if m := reISODate.FindStringSubmatch(t); m != nil {
		y, _ := strconv.Atoi(m[1])
		mo, _ := strconv.Atoi(m[2])
		d, _ := strconv.Atoi(m[3])
		if mo >= 1 && mo <= 12 && d >= 1 && d <= daysIn(y, time.Month(mo)) {
			return time.Date(y, time.Month(mo), d, 0, 0, 0, 0, today.Location()), 1, true
		}
		return time.Time{}, 0, false
	}
	day := func(s string) (int, bool) {
		m := reOrdinal.FindStringSubmatch(s)
		if m == nil {
			return 0, false
		}
		d, _ := strconv.Atoi(m[1])
		return d, d >= 1 && d <= 31
	}
	build := func(mo time.Month, d int) (time.Time, bool) {
		y := today.Year()
		if d > daysIn(y, mo) {
			return time.Time{}, false
		}
		out := time.Date(y, mo, d, 0, 0, 0, 0, today.Location())
		if out.Before(today) {
			out = out.AddDate(1, 0, 0)
		}
		return out, true
	}
	if mo, ok := monthNames[strings.TrimSuffix(t, ".")]; ok && i+1 < len(toks) {
		if d, ok := day(toks[i+1]); ok {
			if out, ok := build(mo, d); ok {
				return out, 2, true
			}
		}
	}
	if d, ok := day(t); ok && i+1 < len(toks) {
		if mo, ok := monthNames[strings.TrimSuffix(strings.TrimSuffix(toks[i+1], ","), ".")]; ok {
			if out, ok := build(mo, d); ok {
				return out, 2, true
			}
		}
	}
	return time.Time{}, 0, false
}

func looksLikeURL(w string) bool {
	lw := strings.ToLower(w)
	if strings.HasPrefix(lw, "http://") || strings.HasPrefix(lw, "https://") {
		return true
	}
	for _, h := range meetingHosts {
		if strings.HasPrefix(lw, h) || strings.Contains(lw, "."+h) {
			return true
		}
	}
	return false
}

// cleanTitle drops the joining words left at the ends once the date and
// time are taken out, as "Call mum at" or "- standup".
func cleanTitle(s string) string {
	words := strings.Fields(s)
	trim := map[string]bool{"at": true, "on": true, "for": true, "in": true, "-": true, "–": true, "—": true, "with": false, "by": true, "from": true, "to": true, "and": true, ",": true, "|": true}
	for len(words) > 0 && trim[strings.ToLower(words[len(words)-1])] {
		words = words[:len(words)-1]
	}
	for len(words) > 0 && trim[strings.ToLower(words[0])] && strings.ToLower(words[0]) != "to" {
		words = words[1:]
	}
	out := strings.Join(words, " ")
	out = strings.TrimRight(out, ",;:-–— ")
	if out != "" {
		r := []rune(out)
		if r[0] >= 'a' && r[0] <= 'z' {
			r[0] -= 'a' - 'A'
		}
		out = string(r)
	}
	return out
}
