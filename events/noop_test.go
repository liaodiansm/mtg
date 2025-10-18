package events_test

import (
	"context"
	"net"
	"testing"

	"github.com/liaodiansm/mtg/events"
	"github.com/liaodiansm/mtg/mtglib"
	"github.com/stretchr/testify/suite"
)

type NoopTestSuite struct {
	suite.Suite

	testData map[string]mtglib.Event
	ctx      context.Context
}

func (suite *NoopTestSuite) SetupSuite() {
	suite.testData = map[string]mtglib.Event{
		"start":               mtglib.NewEventStart("connID", net.ParseIP("127.0.0.1")),
		"connected-to-dc":     mtglib.NewEventConnectedToDC("connID", net.ParseIP("127.1.0.1"), 2),
		"domain-fronting":     mtglib.NewEventDomainFronting("connID"),
		"traffic":             mtglib.NewEventTraffic("connID", 1000, true),
		"finish":              mtglib.NewEventFinish("connID"),
		"concurrency-limited": mtglib.NewEventConcurrencyLimited(),
		"replay-attack":       mtglib.NewEventReplayAttack("connID"),
	}
	suite.ctx = context.Background()
}

func (suite *NoopTestSuite) TestStream() {
	stream := events.NewNoopStream()

	for name, v := range suite.testData {
		value := v

		suite.T().Run(name, func(t *testing.T) {
			stream.Send(suite.ctx, value)
		})
	}
}

func (suite *NoopTestSuite) TestObserver() {
	observer := events.NewNoopObserver()

	for name, v := range suite.testData {
		value := v

		suite.T().Run(name, func(t *testing.T) {
			switch typedEvt := value.(type) {
			case mtglib.EventStart:
				observer.EventStart(typedEvt)
			case mtglib.EventConnectedToDC:
				observer.EventConnectedToDC(typedEvt)
			case mtglib.EventDomainFronting:
				observer.EventDomainFronting(typedEvt)
			case mtglib.EventFinish:
				observer.EventFinish(typedEvt)
			case mtglib.EventConcurrencyLimited:
				observer.EventConcurrencyLimited(typedEvt)
			case mtglib.EventReplayAttack:
				observer.EventReplayAttack(typedEvt)
			}
		})
	}

	observer.Shutdown()
}

func TestNoop(t *testing.T) {
	t.Parallel()
	suite.Run(t, &NoopTestSuite{})
}
