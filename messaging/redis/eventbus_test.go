// Copyright (c) 2014 - Max Ekman <max@looplab.se>
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

package redis

import (
	"os"
	"reflect"
	"testing"

	"github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/testutil"
)

func TestEventBus(t *testing.T) {
	// Support Wercker testing with MongoDB.
	host := os.Getenv("WERCKER_REDIS_HOST")
	port := os.Getenv("WERCKER_REDIS_PORT")

	url := ":6379"
	if host != "" && port != "" {
		url = host + ":" + port
	}

	bus, err := NewEventBus("test", url, "")
	if err != nil {
		t.Fatal("there should be no error:", err)
	}
	if bus == nil {
		t.Fatal("there should be a bus")
	}
	defer bus.Close()
	if err = bus.RegisterEventType(&testutil.TestEvent{}, func() eventhorizon.Event {
		return &testutil.TestEvent{}
	}); err != nil {
		t.Error("there should be no error:", err)
	}
	if err = bus.RegisterEventType(&testutil.TestEventOther{}, func() eventhorizon.Event {
		return &testutil.TestEventOther{}
	}); err != nil {
		t.Error("there should be no error:", err)
	}
	localHandler := testutil.NewMockEventHandler()
	globalHandler := testutil.NewMockEventHandler()
	bus.AddLocalHandler(localHandler)
	bus.AddGlobalHandler(globalHandler)

	// Another bus to test the global handlers.
	bus2, err := NewEventBus("test", url, "")
	if err != nil {
		t.Fatal("there should be no error:", err)
	}
	defer bus2.Close()
	if err = bus2.RegisterEventType(&testutil.TestEvent{}, func() eventhorizon.Event {
		return &testutil.TestEvent{}
	}); err != nil {
		t.Error("there should be no error:", err)
	}
	if err = bus2.RegisterEventType(&testutil.TestEventOther{}, func() eventhorizon.Event {
		return &testutil.TestEventOther{}
	}); err != nil {
		t.Error("there should be no error:", err)
	}
	globalHandler2 := testutil.NewMockEventHandler()
	bus2.AddGlobalHandler(globalHandler2)

	t.Log("publish event without handler")
	event1 := &testutil.TestEvent{eventhorizon.NewUUID(), "event1"}
	bus.PublishEvent(event1)
	if !reflect.DeepEqual(localHandler.Events, []eventhorizon.Event{event1}) {
		t.Error("the local handler events should be correct:", localHandler.Events)
	}
	<-globalHandler.Recv
	if !reflect.DeepEqual(globalHandler.Events, []eventhorizon.Event{event1}) {
		t.Error("the global handler events should be correct:", globalHandler.Events)
	}
	<-globalHandler2.Recv
	if !reflect.DeepEqual(globalHandler2.Events, []eventhorizon.Event{event1}) {
		t.Error("the second global handler events should be correct:", globalHandler2.Events)
	}

	t.Log("publish event")
	handler := testutil.NewMockEventHandler()
	bus.AddHandler(handler, &testutil.TestEvent{})
	bus.PublishEvent(event1)
	if !reflect.DeepEqual(handler.Events, []eventhorizon.Event{event1}) {
		t.Error("the handler events should be correct:", handler.Events)
	}
	if !reflect.DeepEqual(localHandler.Events, []eventhorizon.Event{event1, event1}) {
		t.Error("the local handler events should be correct:", localHandler.Events)
	}
	<-globalHandler.Recv
	if !reflect.DeepEqual(globalHandler.Events, []eventhorizon.Event{event1, event1}) {
		t.Error("the global handler events should be correct:", globalHandler.Events)
	}
	<-globalHandler2.Recv
	if !reflect.DeepEqual(globalHandler2.Events, []eventhorizon.Event{event1, event1}) {
		t.Error("the second global handler events should be correct:", globalHandler2.Events)
	}

	t.Log("publish another event")
	bus.AddHandler(handler, &testutil.TestEventOther{})
	event2 := &testutil.TestEventOther{eventhorizon.NewUUID(), "event2"}
	bus.PublishEvent(event2)
	if !reflect.DeepEqual(handler.Events, []eventhorizon.Event{event1, event2}) {
		t.Error("the handler events should be correct:", handler.Events)
	}
	if !reflect.DeepEqual(localHandler.Events, []eventhorizon.Event{event1, event1, event2}) {
		t.Error("the local handler events should be correct:", localHandler.Events)
	}
	<-globalHandler.Recv
	if !reflect.DeepEqual(globalHandler.Events, []eventhorizon.Event{event1, event1, event2}) {
		t.Error("the global handler events should be correct:", globalHandler.Events)
	}
	<-globalHandler2.Recv
	if !reflect.DeepEqual(globalHandler2.Events, []eventhorizon.Event{event1, event1, event2}) {
		t.Error("the second global handler events should be correct:", globalHandler2.Events)
	}
}
