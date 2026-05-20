package logger_test

import (
	"testing"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

type testWriter struct {
	Output []byte
}

func (w *testWriter) Write(p []byte) (n int, err error) {
	w.Output = p
	return len(p), nil
}

func TestLogInitWithWriters(t *testing.T) {
	w := &testWriter{}
	logger.LogInitWithWriterSyncers(zapcore.AddSync(w))

	msg := "test"

	logger.Info("test")

	assert.Contains(t, string(w.Output), msg)
}

func TestWarn(t *testing.T) {
	w := &testWriter{}
	logger.LogInitWithWriterSyncers(zapcore.AddSync(w))

	msg := "warn test"

	logger.Warn(msg)

	output := string(w.Output)
	assert.Contains(t, output, "warn")
	assert.Contains(t, output, msg)
}
