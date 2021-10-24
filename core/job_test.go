package core

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type JobTestSuite struct {
	suite.Suite
	jr JobRunner
}

func (suite *JobTestSuite) SetupTest() {
	suite.jr = InitializeJobRunner(InitializeInMemoryJobStore())
}

func (suite *JobTestSuite) TestStartJob() {
	cmd := mockExecCommand("echo", "hello", "world")
	job, _ := suite.jr.StartJob(cmd)
	assert.Equal(suite.T(), JobState(COMPLETED), job.State)
}

func (suite *JobTestSuite) TestStopJob() {
	cmd := mockExecCommand("echo", "hello", "world")
	job := suite.jr.store.CreateRecord(cmd, 1, JobState(CREATED), nil)
	cmd.Start()
	suite.jr.StopJob(job.Id)

	assert.Equal(suite.T(), JobState(STOPPED), job.State)
}

func (suite *JobTestSuite) TestStopUnstartedJob() {
	cmd := mockExecCommand("echo", "hello", "world")
	job := suite.jr.store.CreateRecord(cmd, 1, JobState(CREATED), nil)

	assert.Error(suite.T(), suite.jr.StopJob(job.Id))
}

func (suite *JobTestSuite) TestRunJob() {
	jr := InitializeJobRunner(InitializeInMemoryJobStore())
	cmd := mockExecCommand("echo", "hello", "world")
	job := jr.store.CreateRecord(cmd, 1, JobState(CREATED), nil)
	jr.runJob(job.Id, cmd)
	assert.FileExists(suite.T(), fmt.Sprintf("/var/log/linux-process-runner/%s.log", job.Id))

	lb := job.Output
	r, _ := lb.NewReader()
	b := make([]byte, 128)
	n, _ := r.Read(b)
	s := string(b[:n])
	assert.Equal(suite.T(), "hello world", s)
}

func TestJobTestSuite(t *testing.T) {
	suite.Run(t, new(JobTestSuite))
}

func TestHelper(t *testing.T) {
	if os.Getenv("GO_TEST_PROCESS") != "1" {
		return
	}

	defer os.Exit(0)
	fmt.Print("hello world")
}

// based off exec_test.go from os/exec library
func mockExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelper", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_TEST_PROCESS=1"}
	return cmd
}
