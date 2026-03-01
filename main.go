package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type EventConfig struct {
	Name string `yaml:"name"`
	Date string `yaml:"date"`
	Time string `yaml:"time,omitempty"`
}

type Config struct {
	Events []EventConfig `yaml:"events"`
}

type EventResponse struct {
	Name   string `json:"name"`
	Target string `json:"target"`
}

func loadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func parseTarget(e EventConfig) (time.Time, error) {
	layout := "2006-01-02"
	dateStr := e.Date
	if e.Time != "" {
		layout = "2006-01-02 15:04:05"
		dateStr = e.Date + " " + e.Time
	}
	return time.Parse(layout, dateStr)
}

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	http.HandleFunc("/api/events", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		events := make([]EventResponse, 0, len(cfg.Events))
		for _, e := range cfg.Events {
			t, err := parseTarget(e)
			if err != nil {
				log.Printf("skipping event %q: invalid date/time: %v", e.Name, err)
				continue
			}
			events = append(events, EventResponse{
				Name:   e.Name,
				Target: t.UTC().Format(time.RFC3339),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if err := json.NewEncoder(w).Encode(events); err != nil {
			log.Printf("error encoding response: %v", err)
		}
	})

	http.Handle("/", http.FileServer(http.Dir("static")))

	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	log.Printf("starting server on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
