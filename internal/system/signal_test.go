// +build !windows

package system

import (
	"syscall"
	"testing"
)

func TestAvailableSignals(t *testing.T) {
	signals := getAvailableSignals()

	// Проверяем сигналы для Linux/Unix
	expectedSignals := map[syscall.Signal]string{
		syscall.SIGTERM: "SIGTERM",
		syscall.SIGKILL: "SIGKILL",
		syscall.SIGINT:  "SIGINT",
		syscall.SIGHUP:  "SIGHUP",
		syscall.SIGQUIT: "SIGQUIT",
		syscall.SIGABRT: "SIGABRT",
		syscall.SIGUSR1: "SIGUSR1",
		syscall.SIGUSR2: "SIGUSR2",
		syscall.SIGSTOP: "SIGSTOP",
		syscall.SIGCONT: "SIGCONT",
	}

	if len(signals) != len(expectedSignals) {
		t.Errorf("Expected %d signals for Unix, got %d", len(expectedSignals), len(signals))
	}

	for _, sig := range signals {
		if _, ok := expectedSignals[sig.Signal]; !ok {
			t.Errorf("Unexpected signal %s in Unix signals list", sig.Name)
		}
	}

	// Проверяем, что все сигналы имеют описание
	for _, sig := range signals {
		if sig.Description == "" {
			t.Errorf("Signal %s has no description", sig.Name)
		}
	}
}

func TestSendSignal(t *testing.T) {
	// Тест на отправку сигнала несуществующему процессу
	err := SendSignal(-1, syscall.SIGTERM)
	if err == nil {
		t.Error("Expected error when sending signal to non-existent process, got nil")
	}
} 