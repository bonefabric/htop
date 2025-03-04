# github.com/bonefabric/htop

Простая реализация системного монитора в стиле htop на Go.

## Возможности

- Отображение использования CPU в реальном времени с цветовой индикацией
- Отображение использования памяти с цветовой индикацией
- Список запущенных процессов с информацией о CPU и памяти
- Цветовая индикация нагрузки:
  - Зеленый: < 50%
  - Пурпурный: 50-69%
  - Желтый: 70-89%
  - Красный: ≥ 90%

## Требования

- Go 1.22 или выше
- Зависимости:
  - github.com/gizak/termui/v3
  - github.com/shirou/gopsutil/v3

## Установка

```bash
go mod download
```

## Запуск

```bash
go run github.com/bonefabric/htop/cmd/thop/main.go
```

## Управление

- `q` или `Ctrl+C` для выхода
- Обновление данных происходит каждую секунду

## Структура проекта

```
.
├── cmd/
│   └── thop/
│       ├── main.go          # Точка входа в приложение
├── internal/
│   ├── ui/
│   │   ├── dashboard.go     # Логика пользовательского интерфейса
│   │   └── dashboard_test.go # Тесты UI компонентов
│   └── system/
│       ├── process.go       # Работа с системными процессами
│       └── process_test.go  # Тесты обработки процессов
├── go.mod                   # Управление зависимостями
└── README.md                # Документация проекта
```