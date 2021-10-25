package core

import (
	"fmt"
	"io"
	"os/exec"
)

type JobState int32

const (
	// CREATED is the initial status when a job is spawned.
	CREATED JobState = 0
	// RUNNING indicates that a job is in progress.
	RUNNING = 1
	// STOPPED means a job has been manually terminated.
	STOPPED = 2
	// COMPLETED is assigned when the job has successfully finished running.
	COMPLETED = 3
	// ERROR is set if a command returned with a non-zero exit code.
	ERROR = 4
)

// JobInfo represents a job within the server's context.
type JobInfo struct {
	Id     string
	Cmd    *exec.Cmd
	Output LogBuffer
	Owner  int32
	State  JobState
	Err    error
}

// JobRunner handles starting, stopping and getting jobs.
type JobRunner struct {
	store *InMemoryJobStore
}

// InitializeJobRunner creates a pointer to an instantiated JobRunner.
func InitializeJobRunner(store *InMemoryJobStore) *JobRunner {
	return &JobRunner{store: store}
}

// StartJob creates and runs a new job.
func (jr *JobRunner) StartJob(id string, cmd *exec.Cmd) error {
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

// StopJob terminates a running job.
func (jr *JobRunner) StopJob(id string) error {
	job, err := jr.store.GetRecord(id)

	if err != nil {
		return err
	}

	if job.Cmd.Process == nil {
		return fmt.Errorf("cannot stop a nil process")
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

// GetJob retrieves an existing job from storage.
func (jr *JobRunner) GetJob(id string) (*JobInfo, error) {
	return jr.store.GetRecord(id)
}

// runJob handles the output of the job. It combines stdout and stderr
// into a single output that gets fed into a file on the system.
func (jr *JobRunner) runJob(id string, cmd *exec.Cmd) error {
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
