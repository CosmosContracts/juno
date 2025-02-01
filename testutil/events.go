package testutil

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/exp/slices"
)

// AssertEventEmitted asserts that ctx's event manager has emitted the given number of events
// of the given type.
func (s *KeeperTestHelper) AssertEventEmitted(ctx sdk.Context, eventTypeExpected string, numEventsExpected int) {
	allEvents := ctx.EventManager().Events()
	// filter out other events
	actualEvents := make([]sdk.Event, 0)
	for _, event := range allEvents {
		if event.Type == eventTypeExpected {
			actualEvents = append(actualEvents, event)
		}
	}
	s.Equal(numEventsExpected, len(actualEvents))
}

func (s *KeeperTestHelper) FindEvent(events []sdk.Event, name string) sdk.Event {
	index := slices.IndexFunc(events, func(e sdk.Event) bool { return e.Type == name })
	if index == -1 {
		return sdk.Event{}
	}

	return events[index]
}

func (s *KeeperTestHelper) ExtractAttributes(event sdk.Event) map[string]string {
	attrs := make(map[string]string)
	if event.Attributes == nil {
		return attrs
	}
	for _, a := range event.Attributes {
		attrs[a.Key] = a.Value
	}

	return attrs
}

func (s *KeeperTestHelper) AssertEventValueEmitted(eventValue, message string) {
	allEvents := s.Ctx.EventManager().Events()
	for _, event := range allEvents {
		for _, attr := range event.Attributes {
			if attr.Value == eventValue {
				return
			}
		}
	}
	s.Fail(message)
}

func (s *KeeperTestHelper) AssertNEventValuesEmitted(eventValue string, nEvents int) {
	emissions := 0
	allEvents := s.Ctx.EventManager().Events()
	for _, event := range allEvents {
		for _, attr := range event.Attributes {
			if attr.Value == eventValue {
				emissions++
			}
		}
	}
	s.Equal(nEvents, emissions, "Expected %v events, got %v", nEvents, emissions)
}

func (s *KeeperTestHelper) AssertEventValueNotEmitted(eventValue, message string) {
	allEvents := s.Ctx.EventManager().Events()
	if len(allEvents) != 0 {
		for _, attr := range allEvents[len(allEvents)-1].Attributes {
			if attr.Value == eventValue {
				s.Fail(message)
			}
		}
	}
}
