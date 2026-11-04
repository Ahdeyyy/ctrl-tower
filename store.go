package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Preset is a pomodoro rhythm: a work length, a short break after each
// session and a long one after every LongEvery sessions.
type Preset struct {
	Name      string `json:"name"`
	Work      int    `json:"work"`      // minutes
	Break     int    `json:"break"`     // minutes
	LongBreak int    `json:"longBreak"` // minutes
	LongEvery int    `json:"longEvery"`
}

// Session is a finished stretch of focus, for the stats.
type Session struct {
	Start   time.Time `json:"start"`
	Minutes float64   `json:"minutes"`
	Tag     string    `json:"tag,omitempty"`
	Preset  string    `json:"preset"`
	// Full tells a session that ran its length from one stopped early.
	Full bool `json:"full"`
}

// Settings are the user's preferences.
type Settings struct {
	Presets []Preset `json:"presets"`
	Preset  int      `json:"preset"` // index of the chosen one
	// AutoStart runs breaks and the next session without a click.
	AutoStart bool `json:"autoStart"`
	// FocusDND holds plain reminders until a work session ends.
	FocusDND bool `json:"focusDnd"`
	// MeetingAware offers to shorten a session that would run into a
	// meeting.
	MeetingAware bool   `json:"meetingAware"`
	DefaultLead  int    `json:"defaultLead"`
	Hotkey       string `json:"hotkey"`
	OpenAtLogin  bool   `json:"openAtLogin"`
	Sound        bool   `json:"sound"`
}

func defaultSettings() Settings {
	return Settings{
		Presets: []Preset{
			{Name: "Deep work", Work: 50, Break: 10, LongBreak: 20, LongEvery: 4},
			{Name: "Reading", Work: 25, Break: 5, LongBreak: 15, LongEvery: 4},
		},
		AutoStart:    true,
		FocusDND:     true,
		MeetingAware: true,
		DefaultLead:  10,
		Hotkey:       "Ctrl+Alt+Space",
		OpenAtLogin:  true,
		Sound:        true,
	}
}

// Data is everything ctrl-tower keeps.
type Data struct {
	Version  int        `json:"version"`
	Items    []*Item    `json:"items"`
	Sessions []Session  `json:"sessions"`
	Settings Settings   `json:"settings"`
	Tags     []string   `json:"tags"`  // recent tags, newest first
	Fired    []string   `json:"fired"` // alerts already shown, as occurrence IDs with a suffix
	Saved    time.Time  `json:"saved"`
}

func newData() *Data {
	return &Data{Version: 1, Settings: defaultSettings()}
}

// Store loads and saves Data as JSON in the user's app data directory.
// Saves are written off the UI thread, newest last, through a temporary
// file so that a crash never leaves half a file.
type Store struct {
	path string
	mu      sync.Mutex // orders writes
	seq     uint64     // snapshots taken, on the main thread
	written uint64     // the newest snapshot written, under mu
	// Disabled keeps fixtures shown by the dev data switch from being
	// written over the user's data.
	Disabled bool
}

func dataDir() string {
	if d, err := os.UserConfigDir(); err == nil {
		return filepath.Join(d, "ctrl-tower")
	}
	return "."
}

func NewStore(dir string) *Store {
	return &Store{path: filepath.Join(dir, "data.json")}
}

func (s *Store) Path() string { return s.path }

// Load reads the data, or returns fresh data when there is none yet. A
// file that does not parse is kept aside as data.json.bad rather than
// overwritten.
func (s *Store) Load() (*Data, error) {
	b, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return newData(), nil
	}
	if err != nil {
		return newData(), err
	}
	d := newData()
	if err := json.Unmarshal(b, d); err != nil {
		_ = os.Rename(s.path, s.path+".bad")
		return newData(), err
	}
	d.normalize()
	return d, nil
}

func (d *Data) normalize() {
	def := defaultSettings()
	if len(d.Settings.Presets) == 0 {
		d.Settings.Presets = def.Presets
	}
	for i := range d.Settings.Presets {
		p := &d.Settings.Presets[i]
		p.Work = clampInt(p.Work, 1, 180)
		p.Break = clampInt(p.Break, 1, 60)
		p.LongBreak = clampInt(p.LongBreak, 1, 90)
		p.LongEvery = clampInt(p.LongEvery, 1, 12)
		p.Name = clampText(p.Name, 40)
	}
	d.Settings.Preset = clampInt(d.Settings.Preset, 0, len(d.Settings.Presets)-1)
	if d.Settings.Hotkey == "" {
		d.Settings.Hotkey = def.Hotkey
	}
	d.Settings.DefaultLead = clampInt(d.Settings.DefaultLead, 0, 24*60)
	items := d.Items[:0]
	for _, it := range d.Items {
		if it == nil || it.Start.IsZero() {
			continue
		}
		if it.ID == "" {
			it.ID = newID()
		}
		it.clamp()
		items = append(items, it)
	}
	d.Items = items
	if len(d.Fired) > 2000 {
		d.Fired = d.Fired[len(d.Fired)-2000:]
	}
}

func clampInt(v, lo, hi int) int {
	return max(lo, min(hi, v))
}

// Save writes a snapshot of d in the background.
func (s *Store) Save(d *Data) {
	if s.Disabled {
		return
	}
	d.Saved = time.Now()
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return
	}
	s.seq++
	go s.write(b, s.seq)
}

// write writes snapshot seq, unless a newer one was written already.
func (s *Store) write(b []byte, seq uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if seq < s.written {
		return nil
	}
	s.written = seq
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// SaveNow writes synchronously, as the app quits.
func (s *Store) SaveNow(d *Data) error {
	if s.Disabled {
		return nil
	}
	d.Saved = time.Now()
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	s.seq++
	return s.write(b, s.seq)
}

func (d *Data) item(id string) *Item {
	for _, it := range d.Items {
		if it.ID == id {
			return it
		}
	}
	return nil
}

func (d *Data) removeItem(id string) *Item {
	for i, it := range d.Items {
		if it.ID == id {
			d.Items = append(d.Items[:i:i], d.Items[i+1:]...)
			return it
		}
	}
	return nil
}

// useTag moves tag to the front of the recent tags.
func (d *Data) useTag(tag string) {
	if tag == "" {
		return
	}
	tags := []string{tag}
	for _, t := range d.Tags {
		if t != tag && len(tags) < 30 {
			tags = append(tags, t)
		}
	}
	d.Tags = tags
}

func (d *Data) fired(id string) bool {
	for i := len(d.Fired) - 1; i >= 0; i-- {
		if d.Fired[i] == id {
			return true
		}
	}
	return false
}

func (d *Data) markFired(id string) {
	d.Fired = append(d.Fired, id)
	if len(d.Fired) > 2000 {
		d.Fired = d.Fired[len(d.Fired)-1500:]
	}
}
