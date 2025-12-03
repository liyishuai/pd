// Copyright 2025 TiKV Project Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scheduling

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/pingcap/failpoint"

	"github.com/tikv/pd/pkg/schedule"
	"github.com/tikv/pd/tests"
)

// Set to 20 seconds, which is 2x the actual initSchedulerTimeout in the coordinator (10 seconds),
// to provide a buffer for test execution overhead and validation.
const (
	initSchedulersTimeout = 20 * time.Second
)

type coordinatorTestSuite struct {
	suite.Suite
	ctx     context.Context
	cancel  context.CancelFunc
	cluster *tests.TestCluster
}

func TestCoordinatorTestSuite(t *testing.T) {
	suite.Run(t, new(coordinatorTestSuite))
}

func (suite *coordinatorTestSuite) SetupTest() {
	var err error
	re := suite.Require()
	suite.ctx, suite.cancel = context.WithCancel(context.Background())
	suite.cluster, err = tests.NewTestCluster(suite.ctx, 1)
	re.NoError(err)

	err = suite.cluster.RunInitialServers()
	re.NoError(err)

	leaderName := suite.cluster.WaitLeader()
	re.NotEmpty(leaderName)
	pdLeader := suite.cluster.GetServer(leaderName)
	re.NoError(pdLeader.BootstrapCluster())
}

func (suite *coordinatorTestSuite) TearDownTest() {
	suite.cluster.Destroy()
	suite.cancel()
}

// TestInitSchedulersWithSlowCommit tests that InitSchedulers returns within its 10-second timeout
// even when etcdCommitSlow failpoint is enabled (which adds 3-second delays to commits).
// The test allows 20 seconds as a buffer for test execution overhead.
func (suite *coordinatorTestSuite) TestInitSchedulersWithSlowCommit() {
	re := suite.Require()

	// Enable the etcdCommitSlow failpoint to simulate slow etcd commits
	re.NoError(failpoint.Enable("github.com/tikv/pd/pkg/storage/kv/etcdCommitSlow", "return"))
	defer func() {
		re.NoError(failpoint.Disable("github.com/tikv/pd/pkg/storage/kv/etcdCommitSlow"))
	}()

	// Get the leader server and its coordinator
	leaderName := suite.cluster.WaitLeader()
	pdLeader := suite.cluster.GetServer(leaderName)
	rc := pdLeader.GetServer().GetRaftCluster()
	re.NotNil(rc)

	// Create a new coordinator to test InitSchedulers
	coordinator := schedule.NewCoordinator(suite.ctx, rc, rc.GetCoordinator().GetHeartbeatStreams())
	re.NotNil(coordinator)

	// Test that InitSchedulers completes within timeout
	done := make(chan struct{})

	go func() {
		coordinator.InitSchedulers(false)
		close(done)
	}()

	select {
	case <-done:
		// InitSchedulers terminated within timeout
	case <-time.After(initSchedulersTimeout):
		re.Fail("InitSchedulers did not complete within timeout")
	}
}

// TestInitSchedulersWithSaveFailed tests that InitSchedulers returns within its 10-second timeout
// even when etcdSaveFailed failpoint is enabled. The test allows 20 seconds as a buffer for test execution overhead.
func (suite *coordinatorTestSuite) TestInitSchedulersWithSaveFailed() {
	re := suite.Require()

	// Enable the etcdSaveFailed failpoint to simulate failed save operations
	re.NoError(failpoint.Enable("github.com/tikv/pd/pkg/storage/kv/etcdSaveFailed", "return"))
	defer func() {
		re.NoError(failpoint.Disable("github.com/tikv/pd/pkg/storage/kv/etcdSaveFailed"))
	}()

	// Get the leader server and its coordinator
	leaderName := suite.cluster.WaitLeader()
	pdLeader := suite.cluster.GetServer(leaderName)
	rc := pdLeader.GetServer().GetRaftCluster()
	re.NotNil(rc)

	// Create a new coordinator to test InitSchedulers
	coordinator := schedule.NewCoordinator(suite.ctx, rc, rc.GetCoordinator().GetHeartbeatStreams())
	re.NotNil(coordinator)

	// Test that InitSchedulers completes within timeout
	done := make(chan struct{})

	go func() {
		coordinator.InitSchedulers(false)
		close(done)
	}()

	select {
	case <-done:
		// InitSchedulers terminated within timeout
	case <-time.After(initSchedulersTimeout):
		re.Fail("InitSchedulers did not complete within timeout")
	}
}

// TestInitSchedulersBaseline tests that InitSchedulers behaves properly under normal conditions
// without any failpoints enabled. This serves as a baseline to verify normal operation.
func (suite *coordinatorTestSuite) TestInitSchedulersBaseline() {
	re := suite.Require()

	// Get the leader server and its coordinator
	leaderName := suite.cluster.WaitLeader()
	pdLeader := suite.cluster.GetServer(leaderName)
	rc := pdLeader.GetServer().GetRaftCluster()
	re.NotNil(rc)

	// Create a new coordinator to test InitSchedulers
	coordinator := schedule.NewCoordinator(suite.ctx, rc, rc.GetCoordinator().GetHeartbeatStreams())
	re.NotNil(coordinator)

	// Test that InitSchedulers completes within timeout
	done := make(chan struct{})

	go func() {
		coordinator.InitSchedulers(false)
		close(done)
	}()

	select {
	case <-done:
		// InitSchedulers terminated within timeout
		re.True(coordinator.AreSchedulersInitialized(),
			"Schedulers should be properly initialized under normal conditions")
	case <-time.After(initSchedulersTimeout):
		re.Fail("InitSchedulers did not complete within timeout")
	}
}
