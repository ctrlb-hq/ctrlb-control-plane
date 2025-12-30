package queue

import (
	"sync"
	"time"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
)

// StalenessCheckerInterface defines operations for checking agent staleness
type StalenessCheckerInterface interface {
	Start()
	Stop()
}

// StalenessChecker periodically checks for agents that haven't sent a heartbeat
// and marks them as disconnected
type StalenessChecker struct {
	repository      AgentQueueRepositoryInterface
	checkInterval   time.Duration
	inactiveTimeout time.Duration
	stopChan        chan struct{}
	wg              sync.WaitGroup
}

// NewStalenessChecker creates a new StalenessChecker
func NewStalenessChecker(
	repository AgentQueueRepositoryInterface,
	checkIntervalSec int,
	inactiveTimeoutSec int,
) *StalenessChecker {
	return &StalenessChecker{
		repository:      repository,
		checkInterval:   time.Duration(checkIntervalSec) * time.Second,
		inactiveTimeout: time.Duration(inactiveTimeoutSec) * time.Second,
		stopChan:        make(chan struct{}),
	}
}

// Start begins the staleness checking loop
func (s *StalenessChecker) Start() {
	s.wg.Add(1)
	go s.run()
	utils.Logger.Info("StalenessChecker started")
}

// Stop gracefully stops the staleness checker
func (s *StalenessChecker) Stop() {
	close(s.stopChan)
	s.wg.Wait()
	utils.Logger.Info("StalenessChecker stopped")
}

func (s *StalenessChecker) run() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.checkStaleAgents()
		}
	}
}

func (s *StalenessChecker) checkStaleAgents() {
	cutoffTime := time.Now().Add(-s.inactiveTimeout).Unix()

	staleAgents, err := s.repository.GetStaleAgents(cutoffTime)
	if err != nil {
		utils.Logger.Sugar().Errorf("Error fetching stale agents: %v", err)
		return
	}

	for _, agentID := range staleAgents {
		if err := s.repository.UpdateAgentStatus(agentID, "disconnected"); err != nil {
			utils.Logger.Sugar().Errorf("Error marking agent %s as disconnected: %v", agentID, err)
		} else {
			utils.Logger.Sugar().Infof("Marked agent %s as disconnected (no heartbeat received)", agentID)
		}
	}
}

