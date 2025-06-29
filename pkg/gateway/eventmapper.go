package gateway

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Axway/agent-sdk/pkg/agent"
	"github.com/Axway/agent-sdk/pkg/transaction"
	transUtil "github.com/Axway/agent-sdk/pkg/transaction/util"
	"github.com/Axway/agent-sdk/pkg/util/log"
)

// EventMapper -
type EventMapper struct {
}

func (m *EventMapper) processMapping(gatewayTrafficLogEntry GwTrafficLogEntry) (*transaction.LogEvent, []transaction.LogEvent, error) {
	centralCfg := agent.GetCentralConfig()

	eventTime := time.Now().UnixNano() / int64(time.Millisecond)
	txID := gatewayTrafficLogEntry.TraceID
	apiName := gatewayTrafficLogEntry.APIName
	txEventID := gatewayTrafficLogEntry.InboundTransaction.ID
	txDetails := gatewayTrafficLogEntry.InboundTransaction
	transInboundLogEventLeg, err := m.createTransactionEvent(eventTime, txID, txDetails, txEventID, "", "Inbound")
	if err != nil {
		return nil, nil, err
	}

	txEventID = gatewayTrafficLogEntry.OutboundTransaction.ID
	txParentEventID := gatewayTrafficLogEntry.InboundTransaction.ID
	txDetails = gatewayTrafficLogEntry.OutboundTransaction
	transOutboundLogEventLeg, err := m.createTransactionEvent(eventTime, txID, txDetails, txEventID, txParentEventID, "Outbound")
	if err != nil {
		return nil, nil, err
	}

	transSummaryLogEvent, err := m.createSummaryEvent(eventTime, apiName, txID, gatewayTrafficLogEntry, centralCfg.GetTeamID())
	if err != nil {
		return nil, nil, err
	}

	return transSummaryLogEvent, []transaction.LogEvent{
		*transInboundLogEventLeg,
		*transOutboundLogEventLeg,
	}, nil
}

func (m *EventMapper) getTransactionEventStatus(code int) transaction.TxEventStatus {
	if code >= 400 {
		return transaction.TxEventStatusFail
	}
	return transaction.TxEventStatusFail
}

func (m *EventMapper) getTransactionSummaryStatus(statusCode int) transaction.TxSummaryStatus {
	transSummaryStatus := transaction.TxSummaryStatusUnknown
	if statusCode >= http.StatusOK && statusCode < http.StatusBadRequest {
		transSummaryStatus = transaction.TxSummaryStatusSuccess
	} else if statusCode >= http.StatusBadRequest && statusCode < http.StatusInternalServerError {
		transSummaryStatus = transaction.TxSummaryStatusFailure
	} else if statusCode >= http.StatusInternalServerError && statusCode < http.StatusNetworkAuthenticationRequired {
		transSummaryStatus = transaction.TxSummaryStatusException
	}
	return transSummaryStatus
}

func (m *EventMapper) buildHeaders(headers map[string]string) string {
	jsonHeader, err := json.Marshal(headers)
	if err != nil {
		log.Error(err.Error())
	}
	return string(jsonHeader)
}

func (m *EventMapper) createTransactionEvent(eventTime int64, txID string, txDetails GwTransaction, eventID, parentEventID, direction string) (*transaction.LogEvent, error) {

	httpProtocolDetails, err := transaction.NewHTTPProtocolBuilder().
		SetURI(txDetails.URI).
		SetMethod(txDetails.Method).
		SetStatus(txDetails.StatusCode, http.StatusText(txDetails.StatusCode)).
		SetHost(txDetails.SourceHost).
		SetHeaders(m.buildHeaders(txDetails.RequestHeaders), m.buildHeaders(txDetails.ResponseHeaders)).
		SetByteLength(txDetails.RequestBytes, txDetails.ResponseBytes).
		SetRemoteAddress("", txDetails.DesHost, txDetails.DestPort).
		SetLocalAddress(txDetails.SourceHost, txDetails.SourcePort).
		Build()
	if err != nil {
		return nil, err
	}

	return transaction.NewTransactionEventBuilder().
		SetTimestamp(eventTime).
		SetTransactionID(txID).
		SetID(eventID).
		SetParentID(parentEventID).
		SetSource(txDetails.SourceHost + ":" + strconv.Itoa(txDetails.SourcePort)).
		SetDestination(txDetails.DesHost + ":" + strconv.Itoa(txDetails.DestPort)).
		SetDirection(direction).
		SetStatus(m.getTransactionEventStatus(txDetails.StatusCode)).
		SetProtocolDetail(httpProtocolDetails).
		Build()
}

func (m *EventMapper) createSummaryEvent(eventTime int64, apiName, txID string, gatewayTrafficLogEntry GwTrafficLogEntry, teamID string) (*transaction.LogEvent, error) {
	statusCode := gatewayTrafficLogEntry.InboundTransaction.StatusCode
	method := gatewayTrafficLogEntry.InboundTransaction.Method
	uri := gatewayTrafficLogEntry.InboundTransaction.URI
	host := gatewayTrafficLogEntry.InboundTransaction.SourceHost

	return transaction.NewTransactionSummaryBuilder().
		SetTimestamp(eventTime).
		SetTransactionID(txID).
		SetStatus(m.getTransactionSummaryStatus(statusCode), strconv.Itoa(statusCode)).
		SetTeam(teamID).
		SetEntryPoint("http", method, uri, host).
		SetProxy(transUtil.FormatProxyID(apiName), apiName, 0).
		Build()
}
