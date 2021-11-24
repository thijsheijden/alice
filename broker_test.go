package alice

import (
	"os"
	"testing"
)

// broker used during tests
var broker Broker

func TestMain(m *testing.M) {
	// Create a broker
	broker, _ = CreateBroker(DefaultConfig)
	exit := m.Run()
	os.Exit(exit)
}
