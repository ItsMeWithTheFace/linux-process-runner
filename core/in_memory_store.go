package core

import (
	"fmt"
	"os/exec"
	"sync"
)

// InMemoryJobStore represents an in-memory concurrent database to store job info.
type InMemoryJobStore struct {
	jobs map[string]*JobInfo
	mu   *sync.RWMutex
}

// InitializeInMemoryJobStore creates an empty InMemoryJobStore.
func InitializeInMemoryJobStore() *InMemoryJobStore {
	return &InMemoryJobStore{
		jobs: make(map[string]*JobInfo),
		mu:   &sync.RWMutex{},
	}
}

// CreateRecord inserts a new record of a job instance.
func (store *InMemoryJobStore) CreateRecord(id string, cmd *exec.Cmd, owner int32, state JobState, jobError error) JobInfo {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.jobs[id] = &JobInfo{
		Id:    id,
		Cmd:   cmd,
		Owner: owner,
		State: state,
		Err:   jobError,
	}
	return *store.jobs[id]
}

// GetRecord returns info on a job if it exists.
func (store *InMemoryJobStore) GetRecord(id string) (JobInfo, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	if jobInfo, ok := store.jobs[id]; ok {
		return *jobInfo, nil
	}
	return JobInfo{}, fmt.Errorf("querying for record that does not exist")
}

// UpdateRecordOutput updates a job with an input/output stream to allow easy
// retrieval of job output.
func (store *InMemoryJobStore) UpdateRecordOutput(id string, logBuffer LogBuffer) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.jobs[id].Output = logBuffer
}

// UpdateRecordState updates a job's current state.
func (store *InMemoryJobStore) UpdateRecordState(id string, newState JobState) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.jobs[id].State = newState
}

// UpdateRecordError populates a job's error field if it encountered an error
// during execution.
func (store *InMemoryJobStore) UpdateRecordError(id string, newError error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.jobs[id].Err = newError
}
