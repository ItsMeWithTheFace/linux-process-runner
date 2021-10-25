package core

import (
	"fmt"
	"io"
	"os/exec"
)

type JobState int32

const (
	CREATED   JobState = 0
	RUNNING            = 1
	STOPPED            = 2
	COMPLETED          = 3
	ERROR              = 4
)

type JobInfo struct {
	Id     string
	Cmd    *exec.Cmd
	Output LogBuffer
	Owner  int32
	State  JobState
	Err    error
}

type JobRunner struct {
	store *InMemoryJobStore
}

type JobManager interface {
	StartJob(string, *exec.Cmd) error
	StopJob(string) error
	GetJob(string) (*JobInfo, error)
}

func InitializeJobRunner(store *InMemoryJobStore) *JobRunner {
	return &JobRunner{store: store}
}

func (jr JobRunner) StartJob(id string, cmd *exec.Cmd) error {
	// TODO: replace with user's cert serial number
	var user int32 = 1

	job := jr.store.CreateRecord(id, cmd, user, JobState(CREATED), nil)

	err := jr.runJob(job.Id, cmd)

	if err != nil {
		jr.store.UpdateRecordState(job.Id, JobState(ERROR))
		jr.store.UpdateRecordError(job.Id, err)
		return err
	}

	jr.store.UpdateRecordState(job.Id, JobState(COMPLETED))

	return nil
}

func (jr JobRunner) StopJob(id string) error {
	job, err := jr.store.GetRecord(id)

	if err != nil {
		return err
	}

	if job.Cmd.Process == nil {
		return fmt.Errorf("Cannot stop a nil process")
	}

	err = job.Cmd.Process.Kill()

	if err != nil {
		jr.store.UpdateRecordState(job.Id, JobState(ERROR))
		jr.store.UpdateRecordError(job.Id, err)
		return err
	}

	jr.store.UpdateRecordState(job.Id, JobState(STOPPED))

	return nil
}

func (jr JobRunner) GetJob(id string) (*JobInfo, error) {
	return jr.store.GetRecord(id)
}

func (jr JobRunner) runJob(id string, cmd *exec.Cmd) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	output := io.MultiReader(stdout, stderr)

	lb, err := NewLogBuffer(id)

	if err != nil {
		return err
	}

	jr.store.UpdateRecordOutput(id, lb)

	err = cmd.Start()

	if err != nil {
		return err
	}

	jr.store.UpdateRecordState(id, JobState(RUNNING))

	if _, err := io.Copy(lb, output); err != nil {
		return err
	}

	return cmd.Wait()
}
