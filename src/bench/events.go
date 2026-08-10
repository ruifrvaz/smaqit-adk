// records and delivers ordered, durable benchmark lifecycle events.
package bench

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// RunEvent is an ordered lifecycle state transition for a benchmark run.
// Events are both persisted to the experiment and optionally delivered to an
// observer so applications do not need to infer state from terminal prose.
type RunEvent struct {
	SchemaVersion     int          `json:"schemaVersion"`
	Sequence          int          `json:"sequence"`
	Type              string       `json:"type"`
	Timestamp         time.Time    `json:"timestamp"`
	ExperimentID      string       `json:"experimentId,omitempty"`
	PlanID            string       `json:"planId,omitempty"`
	Name              string       `json:"name,omitempty"`
	Directory         string       `json:"directory,omitempty"`
	Report            string       `json:"report,omitempty"`
	RunID             string       `json:"runId,omitempty"`
	CaseID            string       `json:"caseId,omitempty"`
	VariantID         string       `json:"variantId,omitempty"`
	Repetition        int          `json:"repetition,omitempty"`
	Status            string       `json:"status,omitempty"`
	RequiredPassed    *bool        `json:"requiredPassed,omitempty"`
	DurationMS        int64        `json:"durationMs,omitempty"`
	CompletedAttempts int          `json:"completedAttempts,omitempty"`
	TotalAttempts     int          `json:"totalAttempts,omitempty"`
	Complete          *bool        `json:"complete,omitempty"`
	Outcome           string       `json:"outcome,omitempty"`
	Winner            string       `json:"winner,omitempty"`
	Failure           *Failure     `json:"failure,omitempty"`
	Diagnostics       []Diagnostic `json:"diagnostics,omitempty"`
}

// RunOptions selects a run subset and receives lifecycle events. Observer is
// called synchronously only after the event has been durably appended to the
// experiment's events.jsonl journal.
type RunOptions struct {
	CaseID     string
	VariantID  string
	Repetition int
	Observer   func(RunEvent)
}

type eventJournal struct {
	mu       sync.Mutex
	file     *os.File
	encoder  *json.Encoder
	observer func(RunEvent)
	sequence int
}

func newEventJournal(directory string, observer func(RunEvent)) (*eventJournal, error) {
	path := filepath.Join(directory, "events.jsonl")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	return &eventJournal{file: file, encoder: json.NewEncoder(file), observer: observer}, nil
}

func (j *eventJournal) emit(event RunEvent) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.sequence++
	event.SchemaVersion = 1
	event.Sequence = j.sequence
	event.Timestamp = time.Now().UTC()
	if err := j.encoder.Encode(event); err != nil {
		return err
	}
	if err := j.file.Sync(); err != nil {
		return err
	}
	if j.observer != nil {
		j.observer(event)
	}
	return nil
}

func (j *eventJournal) close() error {
	return j.file.Close()
}
