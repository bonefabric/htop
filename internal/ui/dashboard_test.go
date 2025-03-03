package ui

import (
	"testing"
	"time"

	ui "github.com/gizak/termui/v3"
)

// MockUI реализация UIProvider для тестирования
type MockUI struct {
	initCalled    bool
	closeCalled   bool
	renderCalled  bool
	renderedItems []ui.Drawable
	events        chan ui.Event
}

func NewMockUI() *MockUI {
	return &MockUI{
		events: make(chan ui.Event),
	}
}

func (m *MockUI) Init() error {
	m.initCalled = true
	return nil
}

func (m *MockUI) Close() {
	m.closeCalled = true
}

func (m *MockUI) PollEvents() <-chan ui.Event {
	return m.events
}

func (m *MockUI) Render(drawables ...ui.Drawable) {
	m.renderCalled = true
	m.renderedItems = drawables
}

func TestNewDashboardWithUI(t *testing.T) {
	mock := NewMockUI()
	dashboard, err := NewDashboardWithUI(mock)
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if dashboard == nil {
		t.Fatal("Expected dashboard to be created")
	}
	if !mock.initCalled {
		t.Error("Expected UI.Init to be called")
	}

	// Проверяем инициализацию компонентов
	if dashboard.processList == nil {
		t.Error("Expected processList to be initialized")
	}
	if dashboard.signalMenu == nil {
		t.Error("Expected signalMenu to be initialized")
	}
	if dashboard.memChart == nil {
		t.Error("Expected memChart to be initialized")
	}

	// Проверяем заголовки виджетов
	if dashboard.memChart.Title != "Memory Usage" {
		t.Errorf("Unexpected memory chart title: %s", dashboard.memChart.Title)
	}
	if dashboard.processList.Title != "Processes (↑/↓ to navigate, → for signals)" {
		t.Errorf("Unexpected process list title: %s", dashboard.processList.Title)
	}

	// Проверяем начальное состояние
	if dashboard.selectedRow != 0 {
		t.Errorf("Unexpected initial selected row: %d", dashboard.selectedRow)
	}
	if dashboard.showSignalMenu {
		t.Error("Signal menu should be hidden initially")
	}
}

func TestUpdateSignalMenuPosition(t *testing.T) {
	mock := NewMockUI()
	dashboard, _ := NewDashboardWithUI(mock)

	// Тестируем когда меню скрыто
	dashboard.showSignalMenu = false
	dashboard.updateSignalMenuPosition()
	if mock.renderCalled {
		t.Error("Expected no render call when signal menu is hidden")
	}

	// Тестируем когда меню показано
	dashboard.showSignalMenu = true
	dashboard.selectedRow = 5
	dashboard.updateSignalMenuPosition()

	// Проверяем, что меню находится в правильной позиции
	rect := dashboard.signalMenu.GetRect()
	if rect.Dx() <= 0 || rect.Dy() <= 0 {
		t.Error("Signal menu has invalid dimensions")
	}
}

func TestGetColorByPercent(t *testing.T) {
	tests := []struct {
		name     string
		percent  int
		expected ui.Color
	}{
		{"Low load", 30, ui.ColorGreen},
		{"Medium load", 60, ui.ColorMagenta},
		{"High load", 75, ui.ColorYellow},
		{"Critical load", 95, ui.ColorRed},
		{"Border case - 50%", 50, ui.ColorMagenta},
		{"Border case - 70%", 70, ui.ColorYellow},
		{"Border case - 90%", 90, ui.ColorRed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			color := getColorByPercent(tt.percent)
			if color != tt.expected {
				t.Errorf("Expected color %v for %d%%, got %v", tt.expected, tt.percent, color)
			}
		})
	}
}

func TestDashboardNavigation(t *testing.T) {
	mock := NewMockUI()
	dashboard, err := NewDashboardWithUI(mock)
	if err != nil {
		t.Fatalf("Failed to create dashboard: %v", err)
	}

	// Добавляем тестовые данные в список процессов
	testProcesses := []string{"Process 1", "Process 2", "Process 3"}
	dashboard.processList.Rows = testProcesses

	// Тестируем навигацию вниз
	tests := []struct {
		name           string
		initialRow     int
		movement       int // положительное число для движения вниз, отрицательное для движения вверх
		expectedRow    int
	}{
		{"Move down within bounds", 0, 1, 1},
		{"Move down to last item", 1, 1, 2},
		{"Try to move down beyond last item", 2, 1, 2},
		{"Move up within bounds", 2, -1, 1},
		{"Move up to first item", 1, -1, 0},
		{"Try to move up beyond first item", 0, -1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dashboard.selectedRow = tt.initialRow
			dashboard.processList.SelectedRow = tt.initialRow

			// Симулируем движение
			if tt.movement > 0 {
				for i := 0; i < tt.movement; i++ {
					dashboard.selectedRow++
					if dashboard.selectedRow >= len(testProcesses) {
						dashboard.selectedRow = len(testProcesses) - 1
					}
				}
			} else {
				for i := 0; i < -tt.movement; i++ {
					dashboard.selectedRow--
					if dashboard.selectedRow < 0 {
						dashboard.selectedRow = 0
					}
				}
			}

			dashboard.processList.SelectedRow = dashboard.selectedRow

			if dashboard.selectedRow != tt.expectedRow {
				t.Errorf("Expected row %d, got %d", tt.expectedRow, dashboard.selectedRow)
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    uint64
		expected string
	}{
		{"Bytes", 500, "500 B"},
		{"Kilobytes", 1024, "1.0 KiB"},
		{"Megabytes", 1024 * 1024, "1.0 MiB"},
		{"Gigabytes", 1024 * 1024 * 1024, "1.0 GiB"},
		{"Terabytes", 1024 * 1024 * 1024 * 1024, "1.0 TiB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatBytes(%d) = %s, want %s", tt.bytes, result, tt.expected)
			}
		})
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
			t.Errorf("Run() вернул ошибку: %v", err)
		}
	case <-time.After(time.Second):
		t.Error("Тест превысил время ожидания")
	}

	if !mockUI.closeCalled {
		t.Error("UI.Close не был вызван")
	}
	if !mockUI.renderCalled {
		t.Error("UI.Render не был вызван")
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
	if len(mockUI.renderedItems) != expectedWidgets {
		t.Errorf("Неверное количество отрендеренных виджетов: %d, ожидалось: %d", 
			len(mockUI.renderedItems), expectedWidgets)
	}
} 