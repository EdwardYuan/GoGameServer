package service_common

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestServerCommonInitCommonInitializesRuntimeFields(t *testing.T) {
	viper.Reset()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")
	if err := writeTestConfig(configPath); err != nil {
		t.Fatal(err)
	}

	s := NewServerCommon("test_1", 1)
	if err := s.InitCommon(dir); err != nil {
		t.Fatalf("InitCommon failed: %v", err)
	}
	if s.CloseChan == nil {
		t.Fatal("CloseChan is nil")
	}
	if s.SvrTick == nil {
		t.Fatal("SvrTick is nil")
	}
	s.Stop()
}

func TestServerCommonLoadConfigReturnsError(t *testing.T) {
	viper.Reset()
	s := NewServerCommon("test_1", 1)
	if err := s.LoadConfig(t.TempDir()); err == nil {
		t.Fatal("LoadConfig returned nil for missing config")
	}
}

func writeTestConfig(path string) error {
	return os.WriteFile(path, []byte(`
[rabbitmq]
url="amqp://guest:guest@localhost:5672/"
[gameserver]
addr="127.0.0.1"
port="8889"
[logingate]
addr="127.0.0.1"
port="8888"
[gamegate]
addr="127.0.0.1"
port="8890"
[dbserver]
addr="127.0.0.1"
port="8891"
[mysql]
addr="127.0.0.1:3306"
[etcd]
endpoints="127.0.0.1:2379"
[proxy]
addr="127.0.0.1:9100"
`), 0600)
}
