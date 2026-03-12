package cmd

import "testing"

func TestExecuteExists(t *testing.T) {
	if rootCmd.Use == "" {
		t.Fatal("rootCmd not initialized")
	}
}
