package gateway

import (
	"encoding/json"

	"github.com/Axway/agent-sdk/pkg/transaction"
	"github.com/Axway/agent-sdk/pkg/util/log"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/publisher"

	"github.com/vivekschauhan/sample_traceability_agent/pkg/config"
)

// EventProcessor - represents the processor for received event to generate event(s) for Amplify Central
// The event processing can be done either when the beat input receives the log entry or before the beat transport
// publishes the event to transport.
// When processing the received log entry on input, the log entry is mapped to structure expected for Amplify Central Observer
// and then beat.Event is published to beat output that produces the event over the configured transport.
// When processing the log entry on output, the log entry is published to output as beat.Event. The output transport invokes
// the Process(events []publisher.Event) method which is set as output event processor. The Process() method processes the received
// log entry and performs the mapping to structure expected for Amplify Central Observer. The method returns the converted Events to
// transport publisher which then produces the events over the transport.
type EventProcessor struct {
	cfg            *config.GatewayConfig
	eventGenerator transaction.EventGenerator
	eventMapper    *EventMapper
}

// NewEventProcessor - return a new EventProcessor
func NewEventProcessor(gateway *config.GatewayConfig) *EventProcessor {
	ep := &EventProcessor{
		cfg:            gateway,
		eventGenerator: transaction.NewEventGenerator(),
		eventMapper:    &EventMapper{},
	}
	return ep
}

// Process - callback set as output event processor that gets invoked by transport publisher to process the received events
func (p *EventProcessor) Process(events []publisher.Event) []publisher.Event {
	newEvents := make([]publisher.Event, 0)
	for _, event := range events {
		// Get the message from the log file
		eventMsgFieldVal, err := event.Content.Fields.GetValue("message")
		if err != nil {
			log.Error(err.Error())
			return newEvents
		}

		eventMsg, ok := eventMsgFieldVal.(string)
		if ok {
			// Unmarshal the message into the struct representing traffic log entry in gateway logs
			beatEvents := p.ProcessRaw(event, []byte(eventMsg))
			for _, beatEvent := range beatEvents {
				publisherEvent := publisher.Event{
					Content: beatEvent,
				}
				newEvents = append(newEvents, publisherEvent)
			}
		}
	}
	return newEvents
}

// ProcessRaw - process the received log entry and returns the event to be published to Amplify ingestion service
func (p *EventProcessor) ProcessRaw(originalEvent publisher.Event, rawEventData []byte) []beat.Event {
	var gatewayTrafficLogEntry GwTrafficLogEntry
	err := json.Unmarshal(rawEventData, &gatewayTrafficLogEntry)
	if err != nil {
		log.Error(err.Error())
		return nil
	}

	// Map the log entry to log event structure expected by Amplify Central Observer
	summaryEvent, detailEvents, err := p.eventMapper.processMapping(gatewayTrafficLogEntry)
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	report, err := transaction.NewEventReportBuilder().
		SetSummaryEvent(*summaryEvent).
		SetDetailEvents(detailEvents).
		SetEventTime(originalEvent.Content.Timestamp).
		SetSkipSampleHandling().
		Build()

	if err != nil {
		log.Error(err.Error())
		return nil
	}

	beatEvents, err := p.eventGenerator.CreateFromEventReport(report)
	if err != nil {
		log.Error(err.Error())
		return nil
	}

	return beatEvents
}
