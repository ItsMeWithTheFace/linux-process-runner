package core

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type InMemoryJobStoreTestSuite struct {
	suite.Suite
	store *InMemoryJobStore
}

func (suite *InMemoryJobStoreTestSuite) SetupTest() {
	suite.store = InitializeInMemoryJobStore()
}

func (suite *InMemoryJobStoreTestSuite) TestCreateRecord() {
	cases := []struct {
		command   string
		arguments []string
		owner     int32
		state     JobState
		jobError  error
	}{
		{"/bin/ls", []string{}, 123, Stopped, nil},
		{"/usr/bin/tail", []string{"-f", "log.txt"}, 789, Created, nil},
		{"/bin/cp", []string{"file1", "file2"}, 456, Error, fmt.Errorf("file does not exist")},
	}

	for _, tc := range cases {
		record := suite.store.CreateRecord("1", exec.Command(tc.command, tc.arguments...), tc.owner, tc.state, tc.jobError)
		assert.NotEmpty(suite.T(), record.Id, "it has an ID")
		assert.Equal(suite.T(), tc.command, record.Cmd.Path, "it has the same command")
		assert.Equal(suite.T(), tc.arguments, record.Cmd.Args[1:], "it has the same arguments")
		assert.Equal(suite.T(), tc.owner, record.Owner, "it has the same owner")
		assert.Equal(suite.T(), tc.state, record.State, "it has the same state")
		assert.Equal(suite.T(), tc.jobError, record.Err, "it has the same error")
	}
}

func (suite *InMemoryJobStoreTestSuite) TestGetExistingRecord() {
	jobInfo := suite.store.CreateRecord("1", exec.Command("tail", "-f", "log.txt"), 789, Created, nil)
	retrievedJobInfo, err := suite.store.GetRecord(jobInfo.Id)
	assert.Nil(suite.T(), err, "it should not return an error")
	assert.Equal(suite.T(), jobInfo, retrievedJobInfo, "retrieved record should be equal to created record")
}

func (suite *InMemoryJobStoreTestSuite) TestGetNonExistentRecord() {
	jobInfo, err := suite.store.GetRecord("non-existent-id")
	assert.NotNil(suite.T(), err, "it should return an error")
	assert.Equal(suite.T(), JobInfo{}, jobInfo, "it should not return any job info")
}

func (suite *InMemoryJobStoreTestSuite) TestUpdateRecordState() {
	jobInfo := suite.store.CreateRecord("1", exec.Command("tail", "-f", "log.txt"), 789, Created, nil)
	suite.store.UpdateRecordState(jobInfo.Id, Stopped)
	updatedJobInfo, _ := suite.store.GetRecord(jobInfo.Id)
	assert.Equal(suite.T(), JobState(Stopped), updatedJobInfo.State)
}

func (suite *InMemoryJobStoreTestSuite) TestUpdateRecordOutput() {
	jobInfo := suite.store.CreateRecord("1", exec.Command("tail", "-f", "log.txt"), 789, Created, nil)
	lb, err := NewLogBuffer(jobInfo.Id)

	assert.NoError(suite.T(), err, "a log buffer should not produce an error")
	suite.store.UpdateRecordOutput(jobInfo.Id, lb)
	updatedJobInfo, _ := suite.store.GetRecord(jobInfo.Id)
	assert.Equal(suite.T(), lb, updatedJobInfo.Output)
}

func TestInMemoryJobTestSuite(t *testing.T) {
	suite.Run(t, new(InMemoryJobStoreTestSuite))
}
