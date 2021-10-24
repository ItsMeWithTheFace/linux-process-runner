package core

import (
	"errors"
	"os/exec"
	"sync"

	uuid "github.com/google/uuid"
)

type JobStore interface {
	CreateRecord(*exec.Cmd, int32, JobState, error) *JobInfo
	GetRecord(string) (*JobInfo, error)
	UpdateRecordOutput(string, LogBuffer)
	UpdateRecordState(string, JobState)
	UpdateRecordError(string, error)
}

type InMemoryJobStore struct {
	jobs map[string]*JobInfo
	mu   *sync.RWMutex
}

func InitializeInMemoryJobStore() InMemoryJobStore {
	return InMemoryJobStore{
		jobs: make(map[string]*JobInfo),
		mu:   &sync.RWMutex{},
	}
}

func (store InMemoryJobStore) CreateRecord(cmd *exec.Cmd, owner int32, state JobState, jobError error) *JobInfo {

	id := uuid.New()

	store.mu.Lock()
	defer store.mu.Unlock()
	store.jobs[id.String()] = &JobInfo{
		Id:    id.String(),
		Cmd:   cmd,
		Owner: owner,
		State: state,
		Err:   jobError,
	}
	return store.jobs[id.String()]
}

func (store InMemoryJobStore) GetRecord(id string) (*JobInfo, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	if jobInfo, ok := store.jobs[id]; ok {
		return jobInfo, nil
	}
	return nil, errors.New("Querying for record that does not exist")
}

func (store InMemoryJobStore) UpdateRecordOutput(id string, logBuffer LogBuffer) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.jobs[id].Output = logBuffer
}

func (store InMemoryJobStore) UpdateRecordState(id string, newState JobState) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.jobs[id].State = newState
}

func (store InMemoryJobStore) UpdateRecordError(id string, newError error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.jobs[id].Err = newError
}
