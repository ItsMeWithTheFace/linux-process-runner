package core

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type InMemoryJobStoreTestSuite struct {
	suite.Suite
	store JobStore
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
		{"/bin/ls", []string{}, 123, STOPPED, nil},
		{"/usr/bin/tail", []string{"-f", "log.txt"}, 789, CREATED, nil},
		{"/bin/cp", []string{"file1", "file2"}, 456, ERROR, errors.New("File does not exist")},
	}

	for _, tc := range cases {
		record, err := suite.store.CreateRecord(exec.Command(tc.command, tc.arguments...), tc.owner, tc.state, tc.jobError)
		assert.Nil(suite.T(), err, "it should not produce an error")
		assert.NotNil(suite.T(), record.id, "it has an ID")
		assert.Equal(suite.T(), tc.command, record.cmd.Path, "it has the same command")
		assert.Equal(suite.T(), tc.arguments, record.cmd.Args[1:], "it has the same arguments")
		assert.Equal(suite.T(), tc.owner, record.owner, "it has the same owner")
		assert.Equal(suite.T(), tc.state, record.state, "it has the same state")
		assert.Equal(suite.T(), tc.jobError, record.err, "it has the same error")
	}
}

func (suite *InMemoryJobStoreTestSuite) TestGetExistingRecord() {
	jobInfo, err := suite.store.CreateRecord(exec.Command("tail", "-f", "log.txt"), 789, CREATED, nil)
	retrievedJobInfo, err := suite.store.GetRecord(jobInfo.id)
	assert.Nil(suite.T(), err, "it should not return an error")
	assert.Equal(suite.T(), jobInfo, retrievedJobInfo, "retrieved record should be equal to created record")
}

func (suite *InMemoryJobStoreTestSuite) TestGetNonExistentRecord() {
	jobInfo, err := suite.store.GetRecord("non-existent-id")
	assert.NotNil(suite.T(), err, "it should return an error")
	assert.Nil(suite.T(), jobInfo, "it should not return any job info")
}

func (suite *InMemoryJobStoreTestSuite) TestUpdateRecordState() {
	jobInfo, err := suite.store.CreateRecord(exec.Command("tail", "-f", "log.txt"), 789, CREATED, nil)
	suite.store.UpdateRecordState(jobInfo.id, STOPPED)
	assert.Nil(suite.T(), err, "it should not return an error")
	assert.Equal(suite.T(), JobState(STOPPED), jobInfo.state)
}

func TestInMemoryJobTestSuite(t *testing.T) {
	suite.Run(t, new(InMemoryJobStoreTestSuite))
}
