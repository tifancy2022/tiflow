// Copyright 2022 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package memory

import (
	"fmt"
	"sync"

	"github.com/pingcap/tiflow/cdc/model"
	"github.com/pingcap/tiflow/cdc/sinkv2/eventsink"
	"github.com/pingcap/tiflow/cdc/sinkv2/tablesink/state"
	"github.com/pingcap/tiflow/pkg/chann"
)

// Assert EventSink[E event.TableEvent] implementation
var _ eventsink.EventSink[*model.RowChangedEvent] = (*Sink)(nil)

var (
	sinksMu sync.RWMutex
	sinks   = make(map[string]*Sink)
)

type Sink struct {
	id      string
	eventCh *chann.Chann[*model.RowChangedEvent]
}

// New create a memory DML sink.
func New(sinkID string) *Sink {
	sinksMu.Lock()
	defer sinksMu.Unlock()

	if _, ok := sinks[sinkID]; ok {
		panic(fmt.Sprintf("memory sink %s already exists", sinkID))
	}

	sinks[sinkID] = &Sink{
		id:      sinkID,
		eventCh: chann.New[*model.RowChangedEvent](),
	}
	return sinks[sinkID]
}

// WriteEvents write events to the sink.
func (s *Sink) WriteEvents(rows ...*eventsink.CallbackableEvent[*model.RowChangedEvent]) error {
	for _, row := range rows {
		if row.GetTableSinkState() != state.TableSinkSinking {
			// The table where the event comes from is in stopping, so it's safe
			// to drop the event directly.
			row.Callback()
			continue
		}
		s.eventCh.In() <- row.Event
	}
	return nil
}

// Close closes the sink and removes it from the global sink map.
func (s *Sink) Close() error {
	s.eventCh.Close()
	sinksMu.Lock()
	delete(sinks, s.id)
	sinksMu.Unlock()
	return nil
}

// MakeSinkURI creates a sink URI for the memory sink.
func MakeSinkURI(sinkID string) string {
	return fmt.Sprintf("memory://%s", sinkID)
}

// ConsumeEvents consume events from the sink with the given sink ID.
// This is hack for other components in the same process to consume events directly.
// Note that the events can only be consumed once.
func ConsumeEvents(sinkID string) (<-chan *model.RowChangedEvent, bool) {
	sinksMu.RLock()
	defer sinksMu.RUnlock()

	sink, ok := sinks[sinkID]
	if !ok {
		return nil, false
	}
	return sink.eventCh.Out(), true
}
