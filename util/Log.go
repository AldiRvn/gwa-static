package util

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	waLog "go.mau.fi/whatsmeow/util/log"
	"resty.dev/v3"
)

type Logger struct {
	lokiUri     string
	serviceName string
	minLevel    slog.Leveler
	logger      *slog.Logger
}

func NewLogger(lokiUri, serviceName string, minLevel slog.Leveler) *Logger {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: minLevel,
	}))
	return &Logger{
		lokiUri:     lokiUri,
		serviceName: serviceName,
		minLevel:    minLevel,
		logger:      logger,
	}
}

func (l *Logger) sendToLoki(level slog.Leveler, msg string, args ...any) {
	logger := l.logger.With("method", "send to loki")

	client := resty.New()
	defer client.Close()

	res, err := client.R().
		EnableTrace().SetBody(
		map[string]any{
			"streams": []map[string]any{
				{
					"stream": map[string]string{
						"service_name": l.serviceName,
						"level":        level.Level().String(),
					},
					"values": [][]string{{fmt.Sprint(time.Now().UnixNano()), fmt.Sprintf(msg, args...)}},
				},
			},
		}).
		Post(fmt.Sprintf("%s/loki/api/v1/push", l.lokiUri))
	logger.Debug("request", "trace info", res.Request.TraceInfo())
	if err != nil {
		logger.Error("send to loki", "err", err.Error())
		return
	}
	respBody := res.Bytes()
	if len(respBody) > 0 {
		logger.Debug("resp", "body", string(respBody))
	}
}

func (l *Logger) Warnf(msg string, args ...interface{}) {
	l.sendToLoki(slog.LevelWarn, msg, args...)
}

func (l *Logger) Errorf(msg string, args ...interface{}) {
	l.sendToLoki(slog.LevelError, msg, args...)
}

func (l *Logger) Infof(msg string, args ...interface{}) {
	l.sendToLoki(slog.LevelInfo, msg, args...)
}

func (l *Logger) Debugf(msg string, args ...interface{}) {
	l.sendToLoki(slog.LevelDebug, msg, args...)
}

func (l *Logger) Sub(module string) (_ waLog.Logger) {
	return l
}
