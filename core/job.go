package core

import (
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
	id    string
	cmd   *exec.Cmd
	owner int32
	state JobState
	err   error
}

type JobRunner struct {
	store JobStore
}

type JobManager interface {
	StartJob(string, []string) error
	StopJob(string) error
	StreamJob(string) ([]byte, error)
	GetJob(string) (*JobInfo, error)
}

func (jr JobRunner) StartJob(command string, arguments []string) error {
	cmd := exec.Command(command, arguments...)

	job, err := jr.store.CreateRecord(cmd, 1, JobState(CREATED), nil)

	if err != nil {
		return err
	}

	if jr.runJob(job.id, cmd); err != nil {
		jr.store.UpdateRecordState(job.id, JobState(ERROR))
		jr.store.UpdateRecordError(job.id, err)
		return err
	}

	jr.store.UpdateRecordState(job.id, JobState(COMPLETED))
	return nil
}

func (jr JobRunner) StopJob(id string) error {
	job, err := jr.store.GetRecord(id)

	if err != nil {
		return err
	}

	if job.cmd.Process.Kill(); err != nil {
		jr.store.UpdateRecordState(job.id, JobState(ERROR))
		jr.store.UpdateRecordError(job.id, err)
		return err
	}

	jr.store.UpdateRecordState(job.id, JobState(STOPPED))

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

	if err := cmd.Start(); err != nil {
		return err
	}

	jr.store.UpdateRecordState(id, JobState(RUNNING))

	if _, err := io.Copy(lb, output); err != nil {
		return err
	}

	return cmd.Wait()
}
