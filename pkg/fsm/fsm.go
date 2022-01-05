package fsm

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/hashicorp/raft"
)

type state = map[string]string

type FSM struct {
	mu    sync.Mutex
	state state
}

var _ raft.FSM = &FSM{}

func New() *FSM {
	return &FSM{
		state: make(map[string]string),
	}
}

type event struct {
	Type  string `json:"type"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

const (
	eventSet = "set"
)

func (fsm *FSM) Get(k string) (string, bool) {
	fsm.mu.Lock()
	defer fsm.mu.Unlock()
	v, ok := fsm.state[k]
	return v, ok
}

func SetEvent(k, v string) ([]byte, error) {
	data, err := json.Marshal(event{
		Type:  eventSet,
		Key:   k,
		Value: v,
	})
	if err != nil {
		return nil, err
	}
	return data, err
}

// Apply applies a Raft log entry to the key-value store.
func (fsm *FSM) Apply(logEntry *raft.Log) interface{} {
	var e event
	if err := json.Unmarshal(logEntry.Data, &e); err != nil {
		panic("Failed unmarshaling Raft log entry. This is a bug.")
	}

	switch e.Type {
	case eventSet:
		fsm.mu.Lock()
		defer fsm.mu.Unlock()
		fsm.state[e.Key] = e.Value

		return nil
	default:
		panic(fmt.Sprintf("Unrecognized event type in Raft log entry: %s. This is a bug.", e.Type))
	}
}

func (fsm *FSM) Snapshot() (raft.FSMSnapshot, error) {
	fsm.mu.Lock()
	defer fsm.mu.Unlock()

	// NOTE:
	// We need to do the marshalling here since we might handle more Apply's
	// in the time between the snapshot being taken and it being persisted
	// (so it's not safe to pass the underlying map).
	data, err := json.Marshal(fsm.state)
	if err != nil {
		return nil, err
	}
	return &snapshot{data}, nil
}

// Restore stores the key-value store to a previous state.
func (fsm *FSM) Restore(serialized io.ReadCloser) error {
	var state state
	if err := json.NewDecoder(serialized).Decode(&state); err != nil {
		return err
	}
	fsm.state = state
	return nil
}
