package config

import "testing"

func TestConfigVars(t *testing.T) {
	RabbitUrl = "amqp://test"
	if RabbitUrl != "amqp://test" {
		t.Fatalf("expected %s", "amqp://test")
	}
}
