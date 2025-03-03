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

// formatBytes форматирует байты в человекочитаемый формат
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Dashboard представляет главный экран приложения
type Dashboard struct {
	ui          UIProvider
	cpuCharts   []*widgets.Gauge
	memChart    *widgets.Gauge
	processList *widgets.List
	selectedRow int // Индекс выбранного процесса
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

	// Получаем количество ядер процессора
	counts, err := cpu.Counts(true)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU count: %v", err)
	}

	d := &Dashboard{
		ui:          provider,
		cpuCharts:   make([]*widgets.Gauge, counts),
		memChart:    widgets.NewGauge(),
		processList: widgets.NewList(),
		selectedRow: 0,
	}

	// Создаем и настраиваем индикаторы для каждого ядра
	// Располагаем их в два столбца
	gaugeWidth := 50
	gaugeHeight := 3
	columnsCount := 2
	
	for i := 0; i < counts; i++ {
		d.cpuCharts[i] = widgets.NewGauge()
		d.cpuCharts[i].Title = fmt.Sprintf("CPU Core %d", i)
		
		// Вычисляем позицию для текущего индикатора
		column := i % columnsCount
		row := i / columnsCount
		
		x1 := column * gaugeWidth
		y1 := row * gaugeHeight
		x2 := x1 + gaugeWidth
		y2 := y1 + gaugeHeight
		
		d.cpuCharts[i].SetRect(x1, y1, x2, y2)
		d.cpuCharts[i].BarColor = ui.ColorGreen
		d.cpuCharts[i].BorderStyle.Fg = ui.ColorCyan
		d.cpuCharts[i].TitleStyle.Fg = ui.ColorWhite
	}

	// Вычисляем высоту, занимаемую CPU индикаторами
	cpuRowsCount := (counts + columnsCount - 1) / columnsCount // округление вверх
	totalCPUHeight := cpuRowsCount * gaugeHeight

	// Настройка Memory виджета
	d.memChart.Title = "Memory Usage"
	d.memChart.SetRect(0, totalCPUHeight, 100, totalCPUHeight+3)
	d.memChart.BarColor = ui.ColorGreen
	d.memChart.BorderStyle.Fg = ui.ColorCyan
	d.memChart.TitleStyle.Fg = ui.ColorWhite
	d.memChart.Label = "Initializing..." // Начальное значение

	// Настройка списка процессов
	d.processList.Title = "Processes (↑/↓ to navigate)"
	d.processList.SetRect(0, totalCPUHeight+3, 100, totalCPUHeight+17)
	d.processList.BorderStyle.Fg = ui.ColorCyan
	d.processList.TitleStyle.Fg = ui.ColorWhite
	d.processList.TextStyle = ui.NewStyle(ui.ColorWhite)
	d.processList.SelectedRowStyle = ui.NewStyle(ui.ColorBlack, ui.ColorGreen)
	d.processList.WrapText = false

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
				switch e.ID {
				case "q", "<C-c>":
					return nil
				case "<Down>":
					d.selectedRow++
					if d.selectedRow >= len(d.processList.Rows) {
						d.selectedRow = len(d.processList.Rows) - 1
					}
					d.processList.ScrollDown()
				case "<Up>":
					d.selectedRow--
					if d.selectedRow < 0 {
						d.selectedRow = 0
					}
					d.processList.ScrollUp()
				}
				d.processList.SelectedRow = d.selectedRow
				d.ui.Render(d.processList)
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
	// Обновляем CPU для каждого ядра
	cpuPercents, err := cpu.Percent(0, true) // true для получения данных по каждому ядру
	if err != nil {
		return fmt.Errorf("failed to get CPU percent: %v", err)
	}

	for i, percent := range cpuPercents {
		if i < len(d.cpuCharts) {
			intPercent := int(percent)
			d.cpuCharts[i].Percent = intPercent
			d.cpuCharts[i].BarColor = getColorByPercent(intPercent)
		}
	}

	// Обновляем память
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Errorf("failed to get memory info: %v", err)
	}
	percent := int(memInfo.UsedPercent)
	d.memChart.Percent = percent
	d.memChart.BarColor = getColorByPercent(percent)
	
	// Обновляем метку с детальной информацией о памяти
	usedMem := formatBytes(memInfo.Used)
	totalMem := formatBytes(memInfo.Total)
	d.memChart.Label = fmt.Sprintf("%d%% [%s / %s]", percent, usedMem, totalMem)
	
	// Добавляем информацию о свободной памяти в заголовок
	freeMem := formatBytes(memInfo.Available)
	d.memChart.Title = fmt.Sprintf("Memory Usage (Free: %s)", freeMem)

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

		// Сохраняем текущую позицию курсора в пределах списка
		if d.selectedRow >= len(processTexts) {
			d.selectedRow = len(processTexts) - 1
		}
		d.processList.SelectedRow = d.selectedRow
	}

	// Рендерим все виджеты
	drawables := make([]ui.Drawable, 0, len(d.cpuCharts)+2)
	for _, chart := range d.cpuCharts {
		drawables = append(drawables, chart)
	}
	drawables = append(drawables, d.memChart, d.processList)
	d.ui.Render(drawables...)

	return nil
} 