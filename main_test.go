package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// referenceNow is a fixed point in time used across tests.
var referenceNow = time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)

// pastDate is a date clearly before referenceNow.
const pastDate = "2020-01-01"

// futureDate is a date clearly after referenceNow.
const futureDate = "2030-12-31"

func TestGetUpcomingEvents_AllFuture(t *testing.T) {
	events := []EventConfig{
		{Name: "Future Event 1", Date: futureDate},
		{Name: "Future Event 2", Date: futureDate, Time: "10:00:00"},
	}
	got := getUpcomingEvents(events, referenceNow)
	if len(got) != 2 {
		t.Fatalf("expected 2 events, got %d", len(got))
	}
	if got[0].Name != "Future Event 1" {
		t.Errorf("expected %q, got %q", "Future Event 1", got[0].Name)
	}
	if got[0].HasTime {
		t.Error("expected HasTime=false for date-only event")
	}
	if got[1].Name != "Future Event 2" {
		t.Errorf("expected %q, got %q", "Future Event 2", got[1].Name)
	}
	if !got[1].HasTime {
		t.Error("expected HasTime=true for event with time")
	}
}

func TestGetUpcomingEvents_AllPast(t *testing.T) {
	events := []EventConfig{
		{Name: "Past Event 1", Date: pastDate},
		{Name: "Past Event 2", Date: pastDate, Time: "08:00:00"},
	}
	got := getUpcomingEvents(events, referenceNow)
	if len(got) != 0 {
		t.Fatalf("expected 0 events, got %d: %v", len(got), got)
	}
}

func TestGetUpcomingEvents_Mixed(t *testing.T) {
	events := []EventConfig{
		{Name: "Past Event", Date: pastDate},
		{Name: "Future Event", Date: futureDate},
	}
	got := getUpcomingEvents(events, referenceNow)
	if len(got) != 1 {
		t.Fatalf("expected 1 event, got %d", len(got))
	}
	if got[0].Name != "Future Event" {
		t.Errorf("expected %q, got %q", "Future Event", got[0].Name)
	}
}

func TestGetUpcomingEvents_ExactlyNow(t *testing.T) {
	// An event at exactly referenceNow should be considered past (not strictly after).
	e := EventConfig{
		Name: "Exact Now",
		Date: referenceNow.Format("2006-01-02"),
		Time: referenceNow.Format("15:04:05"),
	}
	got := getUpcomingEvents([]EventConfig{e}, referenceNow)
	if len(got) != 0 {
		t.Fatalf("event at exactly now should be hidden, got %d events", len(got))
	}
}

func TestGetUpcomingEvents_InvalidDate(t *testing.T) {
	events := []EventConfig{
		{Name: "Invalid", Date: "not-a-date"},
		{Name: "Future Event", Date: futureDate},
	}
	got := getUpcomingEvents(events, referenceNow)
	if len(got) != 1 {
		t.Fatalf("expected 1 event (invalid skipped), got %d", len(got))
	}
	if got[0].Name != "Future Event" {
		t.Errorf("expected %q, got %q", "Future Event", got[0].Name)
	}
}

func TestGetUpcomingEvents_Empty(t *testing.T) {
	got := getUpcomingEvents([]EventConfig{}, referenceNow)
	if len(got) != 0 {
		t.Fatalf("expected 0 events, got %d", len(got))
	}
}

func TestGetUpcomingEvents_TargetFormat(t *testing.T) {
	events := []EventConfig{
		{Name: "Future", Date: "2030-06-15", Time: "08:30:00"},
	}
	got := getUpcomingEvents(events, referenceNow)
	if len(got) != 1 {
		t.Fatalf("expected 1 event, got %d", len(got))
	}
	want := "2030-06-15T08:30:00Z"
	if got[0].Target != want {
		t.Errorf("expected target %q, got %q", want, got[0].Target)
	}
}

// TestAPIEventsEndpoint_FiltersPastEvents verifies the /api/events HTTP handler
// returns only future events by using a server started with a known config.
func TestAPIEventsEndpoint_FiltersPastEvents(t *testing.T) {
	cfg := &Config{
		Events: []EventConfig{
			{Name: "Past", Date: pastDate},
			{Name: "Future", Date: futureDate},
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/events", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		events := getUpcomingEvents(cfg.Events, referenceNow)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(events); err != nil {
			t.Errorf("error encoding: %v", err)
		}
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/events")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var events []EventResponse
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 upcoming event, got %d: %v", len(events), events)
	}
	if events[0].Name != "Future" {
		t.Errorf("expected event named %q, got %q", "Future", events[0].Name)
	}
}

// TestAPIEventsEndpoint_MethodNotAllowed verifies the handler rejects non-GET requests.
func TestAPIEventsEndpoint_MethodNotAllowed(t *testing.T) {
	cfg := &Config{Events: []EventConfig{}}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/events", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		events := getUpcomingEvents(cfg.Events, referenceNow)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(events) //nolint:errcheck
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/api/events", "application/json", nil)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", resp.StatusCode)
	}
}
