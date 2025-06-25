package beater

import (
	"github.com/Axway/agent-sdk/pkg/traceability"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"

	"github.com/vivekschauhan/sample_traceability_agent/pkg/config"
	"github.com/vivekschauhan/sample_traceability_agent/pkg/gateway"
)

// customLogBeater configuration.
type customLogBeater struct {
	done           chan struct{}
	logReader      *gateway.LogReader
	eventProcessor *gateway.EventProcessor
	client         beat.Client
	eventChannel   chan string
}

var bt *customLogBeater
var gatewayConfig *config.GatewayConfig

// New creates an instance of aws_apigw_traceability_agent.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	bt := &customLogBeater{
		done:         make(chan struct{}),
		eventChannel: make(chan string),
	}

	var err error
	bt.logReader, err = gateway.NewLogReader(gatewayConfig, bt.eventChannel)
	bt.eventProcessor = gateway.NewEventProcessor(gatewayConfig)
	if err != nil {
		return nil, err
	}

	traceability.SetOutputEventProcessor(bt.eventProcessor)

	return bt, nil
}

// SetGatewayConfig - set parsed gateway config
func SetGatewayConfig(gatewayCfg *config.GatewayConfig) {
	gatewayConfig = gatewayCfg
}

// Run starts awsApigwTraceabilityAgent.
func (bt *customLogBeater) Run(b *beat.Beat) error {
	logp.Info("sample_traceability_agent is running! Hit CTRL-C to stop it.")

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	bt.logReader.Start()

	for {
		select {
		case <-bt.done:
			return nil
		case eventData := <-bt.eventChannel:
			eventToPublish := beat.Event{
				Fields: common.MapStr{
					"message": eventData,
				},
			}
			bt.client.Publish(eventToPublish)
		}
	}
}

// Stop stops customLogTraceabilityAgent.
func (bt *customLogBeater) Stop() {
	bt.client.Close()
	close(bt.done)
}
