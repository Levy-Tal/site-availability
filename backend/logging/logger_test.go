package logging

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestInitLogger tests the Init function from the logging package
func TestInitLogger(t *testing.T) {
	// Reset the environment before each test
	defer os.Unsetenv("LOG_LEVEL")

	// Test default behavior (when LOG_LEVEL is not set)
	err := Init()
	assert.NoError(t, err, "Expected no error when LOG_LEVEL is not set")
	assert.Equal(t, logrus.InfoLevel, Logger.GetLevel(), "Logger level should default to 'info'")

	// Test setting LOG_LEVEL to "debug"
	os.Setenv("LOG_LEVEL", "debug")
	err = Init()
	assert.NoError(t, err, "Expected no error when LOG_LEVEL is set to 'debug'")
	assert.Equal(t, logrus.DebugLevel, Logger.GetLevel(), "Logger level should be 'debug' when LOG_LEVEL is set to 'debug'")

	// Test setting an invalid LOG_LEVEL
	os.Setenv("LOG_LEVEL", "invalidlevel")
	err = Init()
	assert.Error(t, err, "Expected error when LOG_LEVEL is set to an invalid level")
	assert.Contains(t, err.Error(), "not a valid logrus Level", "Expected error message for invalid LOG_LEVEL")
}

// TestLoggerFormatter tests the default log formatter
func TestLoggerFormatter(t *testing.T) {
	err := Init()
	assert.NoError(t, err, "Expected no error during logger initialization")

	// Check that the logger uses the TextFormatter with timestamps enabled
	formatter, ok := Logger.Formatter.(*logrus.TextFormatter)
	assert.True(t, ok, "Logger formatter should be of type TextFormatter")
	assert.True(t, formatter.FullTimestamp, "Logger formatter should have FullTimestamp enabled")
}
