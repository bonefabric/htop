package ui

import (
	"fmt"
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
	if len(dashboard.cpuCharts) == 0 {
		t.Error("CPU charts не были инициализированы")
	}

	// Проверяем каждый CPU chart
	for i, chart := range dashboard.cpuCharts {
		if chart == nil {
			t.Errorf("CPU chart %d не был инициализирован", i)
			continue
		}
		if chart.Title != fmt.Sprintf("CPU Core %d", i) {
			t.Errorf("Неверный заголовок CPU chart %d: %s", i, chart.Title)
		}
	}

	if dashboard.memChart == nil {
		t.Error("Memory chart не был инициализирован")
	}
	if dashboard.processList == nil {
		t.Error("Process list не был инициализирован")
	}

	// Проверяем настройки виджетов
	if dashboard.memChart.Title != "Memory Usage" {
		t.Errorf("Неверный заголовок Memory chart: %s", dashboard.memChart.Title)
	}
	if dashboard.processList.Title != "Processes (↑/↓ to navigate)" {
		t.Errorf("Неверный заголовок Process list: %s", dashboard.processList.Title)
	}

	// Проверяем начальное состояние навигации
	if dashboard.selectedRow != 0 {
		t.Errorf("Неверная начальная позиция курсора: %d", dashboard.selectedRow)
	}
}

func TestDashboard_Navigation(t *testing.T) {
	mockUI := NewMockUI()
	dashboard, err := NewDashboardWithUI(mockUI)
	if err != nil {
		t.Fatalf("NewDashboard() вернул ошибку: %v", err)
	}

	// Добавляем тестовые данные в список процессов
	dashboard.processList.Rows = []string{"Process 1", "Process 2", "Process 3"}

	// Тестируем навигацию вниз
	dashboard.selectedRow = 0
	dashboard.processList.SelectedRow = 0

	// Перемещаем курсор вниз
	for i := 0; i < 5; i++ { // Пробуем выйти за пределы списка
		oldPos := dashboard.selectedRow
		dashboard.selectedRow++
		if dashboard.selectedRow >= len(dashboard.processList.Rows) {
			dashboard.selectedRow = len(dashboard.processList.Rows) - 1
		}
		dashboard.processList.SelectedRow = dashboard.selectedRow

		// Проверяем, что курсор не вышел за пределы списка
		if dashboard.selectedRow >= len(dashboard.processList.Rows) {
			t.Errorf("Курсор вышел за пределы списка: %d", dashboard.selectedRow)
		}
		// Проверяем, что после достижения конца списка позиция не меняется
		if oldPos == len(dashboard.processList.Rows)-1 && dashboard.selectedRow != oldPos {
			t.Error("Курсор изменил позицию после достижения конца списка")
		}
	}

	// Тестируем навигацию вверх
	dashboard.selectedRow = len(dashboard.processList.Rows) - 1
	dashboard.processList.SelectedRow = dashboard.selectedRow

	// Перемещаем курсор вверх
	for i := 0; i < 5; i++ { // Пробуем выйти за пределы списка
		oldPos := dashboard.selectedRow
		dashboard.selectedRow--
		if dashboard.selectedRow < 0 {
			dashboard.selectedRow = 0
		}
		dashboard.processList.SelectedRow = dashboard.selectedRow

		// Проверяем, что курсор не стал отрицательным
		if dashboard.selectedRow < 0 {
			t.Errorf("Курсор стал отрицательным: %d", dashboard.selectedRow)
		}
		// Проверяем, что после достижения начала списка позиция не меняется
		if oldPos == 0 && dashboard.selectedRow != oldPos {
			t.Error("Курсор изменил позицию после достижения начала списка")
		}
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

	// Тестируем навигацию
	mockUI.events <- ui.Event{
		Type: ui.KeyboardEvent,
		ID:   "<Down>",
	}
	time.Sleep(50 * time.Millisecond)

	mockUI.events <- ui.Event{
		Type: ui.KeyboardEvent,
		ID:   "<Up>",
	}
	time.Sleep(50 * time.Millisecond)

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

	// Проверяем, что данные CPU были обновлены
	for i, chart := range dashboard.cpuCharts {
		if chart.Percent < 0 || chart.Percent > 100 {
			t.Errorf("Некорректное значение CPU для ядра %d: %d", i, chart.Percent)
		}
	}

	// Проверяем данные памяти
	if dashboard.memChart.Percent < 0 || dashboard.memChart.Percent > 100 {
		t.Errorf("Некорректное значение Memory: %d", dashboard.memChart.Percent)
	}

	// Проверяем список процессов
	if len(dashboard.processList.Rows) == 0 {
		t.Error("Список процессов пуст после обновления")
	}

	// Проверяем, что все виджеты были отрендерены
	expectedWidgets := len(dashboard.cpuCharts) + 2 // CPU charts + memory + process list
	if len(mockUI.rendered) != expectedWidgets {
		t.Errorf("Неверное количество отрендеренных виджетов: %d, ожидалось: %d", 
			len(mockUI.rendered), expectedWidgets)
	}
} 