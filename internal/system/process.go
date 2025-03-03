package system

import (
	"strings"

	"github.com/shirou/gopsutil/v3/process"
)

// ProcessInfo содержит информацию о процессе
type ProcessInfo struct {
	PID     int32
	Name    string
	CPU     float64
	Memory  float32
	Status  string
}

// GetProcessList возвращает список процессов с их характеристиками
func GetProcessList() ([]ProcessInfo, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var processList []ProcessInfo
	for _, p := range processes {
		name, err := p.Name()
		if err != nil {
			continue
		}

		cpu, _ := p.CPUPercent()
		mem, _ := p.MemoryPercent()
		status, _ := p.Status()
		
		processInfo := ProcessInfo{
			PID:     p.Pid,
			Name:    name,
			CPU:     cpu,
			Memory:  mem,
			Status:  strings.Join(status, ","),
		}
		processList = append(processList, processInfo)
	}

	return processList, nil
} 