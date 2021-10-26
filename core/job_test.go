package core

import (
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type JobTestSuite struct {
	suite.Suite
	jr *JobRunner
}

func (suite *JobTestSuite) SetupTest() {
	suite.jr = InitializeJobRunner(InitializeInMemoryJobStore())
}

func (suite *JobTestSuite) TestStartJob() {
	cmd := mockExecCommand("echo", "hello", "world")
	err := suite.jr.StartJob("1", big.NewInt(123), cmd)
	assert.NoError(suite.T(), err, "starting job should not throw an error")

	job, err := suite.jr.store.GetRecord("1")
	assert.NoError(suite.T(), err, "getting should not throw an error")
	assert.Equal(suite.T(), JobState(Completed), job.State, "it should have completed")
}

func (suite *JobTestSuite) TestStopJob() {
	cmd := mockExecCommand("echo", "hello", "world")
	job := suite.jr.store.CreateRecord("1", cmd, big.NewInt(123), JobState(Created), nil)
	cmd.Start()
	suite.jr.StopJob(job.Id)
	updatedJob, _ := suite.jr.store.GetRecord(job.Id)

	assert.Equal(suite.T(), JobState(Stopped), updatedJob.State, "it should have stopped")
}

func (suite *JobTestSuite) TestStopLongJob() {
	cmd := mockExecCommand("sleep")

	errChan := make(chan error, 1)
	go func() {
		errChan <- suite.jr.StartJob("1", big.NewInt(123), cmd)
	}()
	for job, _ := suite.jr.store.GetRecord("1"); job.State == JobState(Created); job, _ = suite.jr.store.GetRecord("1") {
	}
	err := suite.jr.StopJob("1")
	<-errChan
	assert.NoError(suite.T(), err)
	updatedJob, _ := suite.jr.store.GetRecord("1")
	assert.Equal(suite.T(), JobState(Stopped), updatedJob.State, "it should have stopped")
}

func (suite *JobTestSuite) TestStopUnstartedJob() {
	cmd := mockExecCommand("echo", "hello", "world")
	job := suite.jr.store.CreateRecord("1", cmd, big.NewInt(1), JobState(Created), nil)

	assert.Error(suite.T(), suite.jr.StopJob(job.Id), "it should error for unstarted job")
}

func (suite *JobTestSuite) TestRunJob() {
	cmd := mockExecCommand("echo", "hello", "world")
	job := suite.jr.store.CreateRecord("1", cmd, big.NewInt(1), JobState(Created), nil)
	suite.jr.runJob(job.Id, cmd)
	assert.FileExists(suite.T(), fmt.Sprintf("/var/log/linux-process-runner/%s.log", job.Id), "it should create an output file")

	updateJob, _ := suite.jr.store.GetRecord(job.Id)
	lb := updateJob.Output
	r, err := lb.NewReader()
	assert.NoError(suite.T(), err, "getting stream should not produce an error")

	b := make([]byte, 128)
	n, err := r.Read(b)
	assert.NoError(suite.T(), err, "reading stream should not produce an error")

	s := string(b[:n])
	assert.Equal(suite.T(), "hello world", s, "log buffer should contain the same output as command")
}

func TestJobTestSuite(t *testing.T) {
	suite.Run(t, new(JobTestSuite))
}

func TestCommand(t *testing.T) {
	if os.Getenv("GO_TEST_PROCESS") == "1" {
		fmt.Print("hello world")
		os.Exit(0)
	} else if os.Getenv("GO_TEST_PROCESS") == "2" {
		time.Sleep(10 * time.Second)
		os.Exit(0)
	}
	return
}

// mockExecCommand returns a mock command that calls a helper function.
// Based off exec_test.go from the os/exec library.
func mockExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestCommand", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	switch command {
	default:
		fallthrough
	case "echo":
		cmd.Env = []string{"GO_TEST_PROCESS=1"}
	case "sleep":
		cmd.Env = []string{"GO_TEST_PROCESS=2"}
	}

	return cmd
}
