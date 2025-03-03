package system

import (
	"fmt"
	"syscall"

	"github.com/shirou/gopsutil/v3/process"
)

// Signal представляет информацию о сигнале
type Signal struct {
	Name        string
	Description string
	Signal      syscall.Signal
}

// AvailableSignals возвращает список доступных сигналов
var AvailableSignals = getAvailableSignals()

// getAvailableSignals возвращает список сигналов в зависимости от ОС
func getAvailableSignals() []Signal {
	return []Signal{
		{Name: "SIGTERM", Description: "Terminate process", Signal: syscall.SIGTERM},
		{Name: "SIGKILL", Description: "Kill process", Signal: syscall.SIGKILL},
		{Name: "SIGINT", Description: "Interrupt process", Signal: syscall.SIGINT},
		{Name: "SIGHUP", Description: "Hangup", Signal: syscall.SIGHUP},
		{Name: "SIGQUIT", Description: "Quit", Signal: syscall.SIGQUIT},
		{Name: "SIGABRT", Description: "Abort", Signal: syscall.SIGABRT},
		{Name: "SIGUSR1", Description: "User-defined signal 1", Signal: syscall.SIGUSR1},
		{Name: "SIGUSR2", Description: "User-defined signal 2", Signal: syscall.SIGUSR2},
		{Name: "SIGSTOP", Description: "Stop process", Signal: syscall.SIGSTOP},
		{Name: "SIGCONT", Description: "Continue process", Signal: syscall.SIGCONT},
	}
}

// SendSignal отправляет сигнал процессу
func SendSignal(pid int32, sig syscall.Signal) error {
	proc, err := process.NewProcess(pid)
	if err != nil {
		return fmt.Errorf("process not found: %v", err)
	}

	return proc.SendSignal(sig)
} 