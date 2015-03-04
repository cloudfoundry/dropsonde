package log_sender

import (
	"bufio"
	"github.com/cloudfoundry/dropsonde/emitter"
	"github.com/cloudfoundry/dropsonde/events"
	"github.com/cloudfoundry/gosteno"
	"github.com/gogo/protobuf/proto"
	"io"
	"strings"
	"sync"
	"time"
)

// A LogSender emits log events.
type LogSender interface {
	SendAppLog(appId, message, sourceType, sourceInstance string) error
	SendAppErrorLog(appId, message, sourceType, sourceInstance string) error

	ScanLogStream(appId, sourceType, sourceInstance string, reader io.Reader)
	ScanErrorLogStream(appId, sourceType, sourceInstance string, reader io.Reader)
}

type logSender struct {
	eventEmitter            emitter.EventEmitter
	logger                  *gosteno.Logger
	logMessageReceiveCounts map[string]float64
	logMessageTotalCount    float64
	sync.RWMutex
}

// NewLogSender instantiates a logSender with the given EventEmitter.
func NewLogSender(eventEmitter emitter.EventEmitter, counterEmissionInterval time.Duration, logger *gosteno.Logger) LogSender {
	l := logSender{
		eventEmitter:            eventEmitter,
		logger:                  logger,
		logMessageReceiveCounts: make(map[string]float64),
	}

	go func() {
		ticker := time.NewTicker(counterEmissionInterval)
		for {
			<-ticker.C
			l.emitCounters()
		}
	}()

	return &l
}

// SendAppLog sends a log message with the given appid and log message
// with a message type of std out.
// Returns an error if one occurs while sending the event.
func (l *logSender) SendAppLog(appId, message, sourceType, sourceInstance string) error {
	l.Lock()
	l.logMessageTotalCount += 1
	l.logMessageReceiveCounts[appId] = l.logMessageReceiveCounts[appId] + 1
	l.Unlock()

	return l.eventEmitter.Emit(makeLogMessage(appId, message, sourceType, sourceInstance, events.LogMessage_OUT))
}

// SendAppErrorLog sends a log error message with the given appid and log message
// with a message type of std err.
// Returns an error if one occurs while sending the event.
func (l *logSender) SendAppErrorLog(appId, message, sourceType, sourceInstance string) error {
	l.Lock()
	l.logMessageTotalCount += 1
	l.logMessageReceiveCounts[appId] += 1
	l.Unlock()

	return l.eventEmitter.Emit(makeLogMessage(appId, message, sourceType, sourceInstance, events.LogMessage_ERR))
}

// ScanLogStream sends a log message with the given meta-data for each line from reader.
// Restarts on read errors and continues until EOF.
func (l *logSender) ScanLogStream(appId, sourceType, sourceInstance string, reader io.Reader) {
	l.scanLogStream(appId, sourceType, sourceInstance, l.SendAppLog, reader)
}

// ScanErrorLogStream sends a log error message with the given meta-data for each line from reader.
// Restarts on read errors and continues until EOF.
func (l *logSender) ScanErrorLogStream(appId, sourceType, sourceInstance string, reader io.Reader) {
	l.scanLogStream(appId, sourceType, sourceInstance, l.SendAppErrorLog, reader)
}

func (l *logSender) scanLogStream(appId, sourceType, sourceInstance string, sender func(string, string, string, string) error, reader io.Reader) {
	for {
		err := sendScannedLines(appId, sourceType, sourceInstance, bufio.NewScanner(reader), sender)
		if err == bufio.ErrTooLong {
			l.SendAppErrorLog(appId, "Dropped log message: message too long (>64K without a newline)", sourceType, sourceInstance)
			continue
		}
		if err == nil {
			l.logger.Debugf("EOF on log stream for app %s/%s", appId, sourceInstance)
		} else {
			l.logger.Infof("ScanLogStream: Error while reading STDOUT/STDERR for app %s/%s: %s", appId, sourceInstance, err.Error())
		}
		return
	}
}

func (l *logSender) emitCounters() {
	l.Lock()
	defer l.Unlock()

	l.eventEmitter.Emit(&events.ValueMetric{
		Name:  proto.String("logSenderTotalMessagesRead"),
		Value: proto.Float64(l.logMessageTotalCount),
	})

	for appID, count := range l.logMessageReceiveCounts {
		l.eventEmitter.Emit(&events.ValueMetric{
			Name:  proto.String("logSenderTotalMessagesRead." + appID),
			Value: proto.Float64(count),
		})
	}
}

func makeLogMessage(appId, message, sourceType, sourceInstance string, messageType events.LogMessage_MessageType) *events.LogMessage {
	return &events.LogMessage{
		Message:        []byte(message),
		AppId:          proto.String(appId),
		MessageType:    &messageType,
		SourceType:     &sourceType,
		SourceInstance: &sourceInstance,
		Timestamp:      proto.Int64(time.Now().UnixNano()),
	}
}

func sendScannedLines(appId, sourceType, sourceInstance string, scanner *bufio.Scanner, send func(string, string, string, string) error) error {
	for scanner.Scan() {
		line := scanner.Text()

		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		send(appId, line, sourceType, sourceInstance)
	}
	return scanner.Err()
}
