package ui

import (
	"fmt"
	"log"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	"thop/internal/system"
)

// UIProvider интерфейс для абстракции termui
type UIProvider interface {
	Init() error
	Close()
	PollEvents() <-chan ui.Event
	Render(...ui.Drawable)
}

// RealUI реализация UIProvider для реального termui
type RealUI struct{}

func (r *RealUI) Init() error {
	return ui.Init()
}

func (r *RealUI) Close() {
	ui.Close()
}

func (r *RealUI) PollEvents() <-chan ui.Event {
	return ui.PollEvents()
}

func (r *RealUI) Render(drawables ...ui.Drawable) {
	ui.Render(drawables...)
}

// Dashboard представляет главный экран приложения
type Dashboard struct {
	ui          UIProvider
	cpuChart    *widgets.Gauge
	memChart    *widgets.Gauge
	processList *widgets.List
}

// NewDashboard создает новый экземпляр Dashboard
func NewDashboard() (*Dashboard, error) {
	return NewDashboardWithUI(&RealUI{})
}

// NewDashboardWithUI создает новый экземпляр Dashboard с указанным UI провайдером
func NewDashboardWithUI(provider UIProvider) (*Dashboard, error) {
	if err := provider.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize termui: %v", err)
	}

	d := &Dashboard{
		ui:          provider,
		cpuChart:    widgets.NewGauge(),
		memChart:    widgets.NewGauge(),
		processList: widgets.NewList(),
	}

	// Настройка CPU виджета
	d.cpuChart.Title = "CPU Usage"
	d.cpuChart.SetRect(0, 0, 50, 3)
	d.cpuChart.BarColor = ui.ColorGreen
	d.cpuChart.BorderStyle.Fg = ui.ColorCyan
	d.cpuChart.TitleStyle.Fg = ui.ColorWhite

	// Настройка Memory виджета
	d.memChart.Title = "Memory Usage"
	d.memChart.SetRect(0, 3, 50, 6)
	d.memChart.BarColor = ui.ColorGreen
	d.memChart.BorderStyle.Fg = ui.ColorCyan
	d.memChart.TitleStyle.Fg = ui.ColorWhite

	// Настройка списка процессов
	d.processList.Title = "Processes"
	d.processList.SetRect(0, 6, 100, 20)
	d.processList.BorderStyle.Fg = ui.ColorCyan
	d.processList.TitleStyle.Fg = ui.ColorWhite

	return d, nil
}

// getColorByPercent возвращает цвет в зависимости от процента загрузки
func getColorByPercent(percent int) ui.Color {
	switch {
	case percent >= 90:
		return ui.ColorRed
	case percent >= 70:
		return ui.ColorYellow
	case percent >= 50:
		return ui.ColorMagenta
	default:
		return ui.ColorGreen
	}
}

// Run запускает основной цикл обновления Dashboard
func (d *Dashboard) Run() error {
	defer d.ui.Close()

	// Создаем тикер для обновления данных
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// Обработка выхода по клавише 'q'
	uiEvents := d.ui.PollEvents()

	for {
		select {
		case e := <-uiEvents:
			if e.Type == ui.KeyboardEvent {
				if e.ID == "q" || e.ID == "<C-c>" {
					return nil
				}
			}
		case <-ticker.C:
			if err := d.update(); err != nil {
				return err
			}
		}
	}
}

// update обновляет все виджеты Dashboard
func (d *Dashboard) update() error {
	// Обновляем CPU
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return fmt.Errorf("failed to get CPU percent: %v", err)
	}
	if len(cpuPercent) > 0 {
		percent := int(cpuPercent[0])
		d.cpuChart.Percent = percent
		d.cpuChart.BarColor = getColorByPercent(percent)
	}

	// Обновляем память
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Errorf("failed to get memory info: %v", err)
	}
	percent := int(memInfo.UsedPercent)
	d.memChart.Percent = percent
	d.memChart.BarColor = getColorByPercent(percent)

	// Обновляем список процессов
	processes, err := system.GetProcessList()
	if err != nil {
		log.Printf("failed to get process list: %v", err)
	} else {
		processTexts := make([]string, 0, len(processes))
		for _, p := range processes {
			processTexts = append(processTexts,
				fmt.Sprintf("[%d] %s (CPU: %.1f%%, Mem: %.1f%%, Status: %s)",
					p.PID, p.Name, p.CPU, p.Memory, p.Status))
		}
		d.processList.Rows = processTexts
	}

	d.ui.Render(d.cpuChart, d.memChart, d.processList)
	return nil
} 