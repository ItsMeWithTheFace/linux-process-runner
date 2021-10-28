package core

import (
	"fmt"
	"io"
	"math/big"
	"os"
	"os/exec"
	"syscall"
)

type JobState int32

const (
	// Created is the initial status when a job is spawned.
	Created JobState = iota
	// Running indicates that a job is in progress.
	Running
	// Stopped means a job has been manually terminated.
	Stopped
	// Completed is assigned when the job has successfully finished running.
	Completed
	// Error is set if a command returned with a non-zero exit code.
	Error
)

// JobInfo represents a job within the server's context.
type JobInfo struct {
	Id     string
	Cmd    *exec.Cmd
	Output LogBuffer
	Owner  *big.Int
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

// CreateJob creates a new job and stores it in memory.
func (jr *JobRunner) CreateJob(id string, owner *big.Int, cmd *exec.Cmd) JobInfo {
	return jr.store.CreateRecord(id, cmd, owner, JobState(Created), nil)
}

// StartJob runs a job.
func (jr *JobRunner) StartJob(job JobInfo) error {
	err := jr.runJob(job.Id, job.Cmd)

	if err != nil && !isKilled(err) {
		jr.store.UpdateRecordError(job.Id, err)
		return err
	}

	if job.Cmd.ProcessState.Success() {
		jr.store.UpdateRecordState(job.Id, JobState(Completed))
	}

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

	if job.State > Running {
		return fmt.Errorf("cannot stop a job in a terminal state")
	}

	err = job.Cmd.Process.Kill()

	if err != nil && err != os.ErrProcessDone {
		jr.store.UpdateRecordError(job.Id, err)
		return err
	}

	err = jr.store.UpdateRecordState(job.Id, JobState(Stopped))

	if err != nil {
		return err
	}
	return nil
}

// GetJob retrieves an existing job from storage.
func (jr *JobRunner) GetJob(id string) (JobInfo, error) {
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

	defer lb.Close()

	jr.store.UpdateRecordOutput(id, lb)

	err = cmd.Start()

	if err != nil {
		return err
	}

	jr.store.UpdateRecordState(id, JobState(Running))

	if _, err := io.Copy(lb, output); err != nil {
		return err
	}

	return cmd.Wait()
}

// isKilled checks if a command exited via a SIGKILL signal by
// checking its Wait() status.
func isKilled(err error) bool {
	if exitErr, ok := err.(*exec.ExitError); ok {
		if waitStatus, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			return waitStatus.Signal() == syscall.SIGKILL
		}
	}
	return false
}
