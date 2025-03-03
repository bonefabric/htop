package ui

import (
	"testing"
	"time"

	ui "github.com/gizak/termui/v3"
)

// MockUI реализация UIProvider для тестирования
type MockUI struct {
	initialized bool
	closed      bool
	events      chan ui.Event
	rendered    []ui.Drawable
}

func NewMockUI() *MockUI {
	return &MockUI{
		events: make(chan ui.Event),
	}
}

func (m *MockUI) Init() error {
	m.initialized = true
	return nil
}

func (m *MockUI) Close() {
	m.closed = true
}

func (m *MockUI) PollEvents() <-chan ui.Event {
	return m.events
}

func (m *MockUI) Render(drawables ...ui.Drawable) {
	m.rendered = drawables
}

func TestGetColorByPercent(t *testing.T) {
	testCases := []struct {
		name     string
		percent  int
		expected ui.Color
	}{
		{
			name:     "Low load",
			percent:  30,
			expected: ui.ColorGreen,
		},
		{
			name:     "Medium load",
			percent:  60,
			expected: ui.ColorMagenta,
		},
		{
			name:     "High load",
			percent:  75,
			expected: ui.ColorYellow,
		},
		{
			name:     "Critical load",
			percent:  95,
			expected: ui.ColorRed,
		},
		{
			name:     "Border case - 50%",
			percent:  50,
			expected: ui.ColorMagenta,
		},
		{
			name:     "Border case - 70%",
			percent:  70,
			expected: ui.ColorYellow,
		},
		{
			name:     "Border case - 90%",
			percent:  90,
			expected: ui.ColorRed,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			color := getColorByPercent(tc.percent)
			if color != tc.expected {
				t.Errorf("getColorByPercent(%d) = %v, want %v", 
					tc.percent, color, tc.expected)
			}
		})
	}
}

func TestNewDashboard(t *testing.T) {
	mockUI := NewMockUI()
	dashboard, err := NewDashboardWithUI(mockUI)
	if err != nil {
		t.Fatalf("NewDashboard() вернул ошибку: %v", err)
	}

	// Проверяем, что UI был инициализирован
	if !mockUI.initialized {
		t.Error("UI не был инициализирован")
	}

	// Проверяем, что все виджеты были созданы
	if dashboard.cpuChart == nil {
		t.Error("CPU chart не был инициализирован")
	}
	if dashboard.memChart == nil {
		t.Error("Memory chart не был инициализирован")
	}
	if dashboard.processList == nil {
		t.Error("Process list не был инициализирован")
	}

	// Проверяем настройки виджетов
	if dashboard.cpuChart.Title != "CPU Usage" {
		t.Errorf("Неверный заголовок CPU chart: %s", dashboard.cpuChart.Title)
	}
	if dashboard.memChart.Title != "Memory Usage" {
		t.Errorf("Неверный заголовок Memory chart: %s", dashboard.memChart.Title)
	}
	if dashboard.processList.Title != "Processes" {
		t.Errorf("Неверный заголовок Process list: %s", dashboard.processList.Title)
	}
}

func TestDashboard_Run(t *testing.T) {
	mockUI := NewMockUI()
	dashboard, err := NewDashboardWithUI(mockUI)
	if err != nil {
		t.Fatalf("NewDashboard() вернул ошибку: %v", err)
	}

	// Запускаем dashboard в отдельной горутине
	done := make(chan error)
	go func() {
		done <- dashboard.Run()
	}()

	// Даем время на инициализацию
	time.Sleep(100 * time.Millisecond)

	// Отправляем событие выхода
	mockUI.events <- ui.Event{
		Type: ui.KeyboardEvent,
		ID:   "q",
	}

	// Ждем завершения
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Dashboard.Run() вернул ошибку: %v", err)
		}
	case <-time.After(time.Second):
		t.Error("Dashboard.Run() не завершился после события выхода")
	}

	// Проверяем, что UI был закрыт
	if !mockUI.closed {
		t.Error("UI не был закрыт")
	}
}

func TestDashboard_Update(t *testing.T) {
	mockUI := NewMockUI()
	dashboard, err := NewDashboardWithUI(mockUI)
	if err != nil {
		t.Fatalf("NewDashboard() вернул ошибку: %v", err)
	}

	// Проверяем обновление данных
	err = dashboard.update()
	if err != nil {
		t.Errorf("Dashboard.update() вернул ошибку: %v", err)
	}

	// Проверяем, что данные были обновлены
	if dashboard.cpuChart.Percent < 0 || dashboard.cpuChart.Percent > 100 {
		t.Errorf("Некорректное значение CPU: %d", dashboard.cpuChart.Percent)
	}
	if dashboard.memChart.Percent < 0 || dashboard.memChart.Percent > 100 {
		t.Errorf("Некорректное значение Memory: %d", dashboard.memChart.Percent)
	}
	if len(dashboard.processList.Rows) == 0 {
		t.Error("Список процессов пуст после обновления")
	}

	// Проверяем, что виджеты были отрендерены
	if len(mockUI.rendered) != 3 {
		t.Errorf("Неверное количество отрендеренных виджетов: %d", len(mockUI.rendered))
	}
} 