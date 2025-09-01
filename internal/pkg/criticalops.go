/********************************************************************************
 *  Copyright (c) 2025 IOTech Ltd
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package pkg

import (
	"sync"
	"time"
)

// CriticalOperationManager provides critical operation management functionality
type CriticalOperationManager struct {
	criticalOperations map[chan struct{}]bool
	criticalOpsMutex   sync.RWMutex
	waitGroup          sync.WaitGroup
}

// NewCriticalOperationManager creates a new critical operations manager
func NewCriticalOperationManager() *CriticalOperationManager {
	return &CriticalOperationManager{
		criticalOperations: make(map[chan struct{}]bool),
		criticalOpsMutex:   sync.RWMutex{},
		waitGroup:          sync.WaitGroup{},
	}
}

// RegisterCriticalOperation registers a critical operation with a finish signal channel
func (m *CriticalOperationManager) RegisterCriticalOperation(finishSignal chan struct{}) {
	m.criticalOpsMutex.Lock()
	defer m.criticalOpsMutex.Unlock()
	m.criticalOperations[finishSignal] = true
	m.waitGroup.Add(1)
	go func() {
		<-finishSignal
		m.waitGroup.Done()
		m.criticalOpsMutex.Lock()
		defer m.criticalOpsMutex.Unlock()
		delete(m.criticalOperations, finishSignal)
	}()
}

// WaitForCriticalOperations waits for all critical operations to complete within the specified timeout
// returns true if all operations completed, false if timeout occurred
func (m *CriticalOperationManager) WaitForCriticalOperations(timeout time.Duration) bool {
	done := make(chan struct{})
	go func() {
		m.waitGroup.Wait()
		close(done)
	}()

	select {
	case <-done:
		m.criticalOpsMutex.Lock()
		defer m.criticalOpsMutex.Unlock()
		clear(m.criticalOperations)
		return true
	case <-time.After(timeout):
		return false
	}
}
