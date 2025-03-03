package system

import (
	"testing"
)

func TestGetProcessList(t *testing.T) {
	processes, err := GetProcessList()
	
	// Проверяем, что функция не возвращает ошибку
	if err != nil {
		t.Errorf("GetProcessList() вернула ошибку: %v", err)
	}

	// Проверяем, что список процессов не пустой
	if len(processes) == 0 {
		t.Error("GetProcessList() вернула пустой список процессов")
	}

	// Проверяем каждый процесс на валидность данных
	for _, p := range processes {
		// PID должен быть неотрицательным (в Windows может быть 0)
		if p.PID < 0 {
			t.Errorf("Процесс имеет некорректный PID: %d", p.PID)
		}

		// Имя процесса не должно быть пустым
		if p.Name == "" {
			t.Errorf("Процесс с PID %d имеет пустое имя", p.PID)
		}

		// CPU не может быть отрицательным
		if p.CPU < 0 {
			t.Errorf("Процесс %s (PID: %d) имеет отрицательное значение CPU: %f", 
				p.Name, p.PID, p.CPU)
		}

		// Memory не может быть отрицательным или больше 100%
		if p.Memory < 0 || p.Memory > 100 {
			t.Errorf("Процесс %s (PID: %d) имеет некорректное значение Memory: %f", 
				p.Name, p.PID, p.Memory)
		}
	}
}

func TestProcessInfo_Validation(t *testing.T) {
	testCases := []struct {
		name    string
		process ProcessInfo
		valid   bool
	}{
		{
			name: "Valid process",
			process: ProcessInfo{
				PID:     1,
				Name:    "test",
				CPU:     5.0,
				Memory:  10.0,
				Status:  "running",
			},
			valid: true,
		},
		{
			name: "Zero PID (Windows)",
			process: ProcessInfo{
				PID:     0,
				Name:    "test",
				CPU:     5.0,
				Memory:  10.0,
				Status:  "running",
			},
			valid: true,
		},
		{
			name: "Invalid PID",
			process: ProcessInfo{
				PID:     -1,
				Name:    "test",
				CPU:     5.0,
				Memory:  10.0,
				Status:  "running",
			},
			valid: false,
		},
		{
			name: "Empty name",
			process: ProcessInfo{
				PID:     1,
				Name:    "",
				CPU:     5.0,
				Memory:  10.0,
				Status:  "running",
			},
			valid: false,
		},
		{
			name: "Invalid CPU",
			process: ProcessInfo{
				PID:     1,
				Name:    "test",
				CPU:     -1.0,
				Memory:  10.0,
				Status:  "running",
			},
			valid: false,
		},
		{
			name: "Invalid Memory",
			process: ProcessInfo{
				PID:     1,
				Name:    "test",
				CPU:     5.0,
				Memory:  101.0,
				Status:  "running",
			},
			valid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isValid := tc.process.PID >= 0 &&
				tc.process.Name != "" &&
				tc.process.CPU >= 0 &&
				tc.process.Memory >= 0 &&
				tc.process.Memory <= 100

			if isValid != tc.valid {
				t.Errorf("ProcessInfo validation test failed for %s", tc.name)
			}
		})
	}
} 