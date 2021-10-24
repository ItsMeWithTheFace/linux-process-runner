package core

import (
	"errors"
	"os/exec"
	"sync"

	uuid "github.com/google/uuid"
)

type JobStore interface {
	CreateRecord(*exec.Cmd, int32, JobState, error) (*JobInfo, error)
	GetRecord(string) (*JobInfo, error)
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

func (store InMemoryJobStore) CreateRecord(cmd *exec.Cmd, owner int32, state JobState, jobError error) (*JobInfo, error) {

	id, err := uuid.NewRandom()

	if err != nil {
		return nil, err
	}

	store.mu.Lock()
	defer store.mu.Unlock()
	store.jobs[id.String()] = &JobInfo{
		id.String(),
		cmd,
		owner,
		state,
		jobError,
	}
	return store.jobs[id.String()], nil
}

func (store InMemoryJobStore) GetRecord(id string) (*JobInfo, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	if jobInfo, ok := store.jobs[id]; ok {
		return jobInfo, nil
	}
	return nil, errors.New("Querying for record that does not exist")
}

func (store InMemoryJobStore) UpdateRecordState(id string, newState JobState) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.jobs[id].state = newState
}

func (store InMemoryJobStore) UpdateRecordError(id string, newError error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.jobs[id].err = newError
}
