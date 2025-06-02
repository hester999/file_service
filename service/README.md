# Go Service



## Запуск

### Локально

```bash
# Установка зависимостей
go mod download

# Запуск сервиса
go run cmd/main.go
```

### Docker

```bash
# Сборка образа
docker build -t go-service .

# Запуск контейнера
docker run -p 8080:8080 go-service
```

## API Endpoints

### 1. Генерация файлов

```http
POST /generate
Content-Type: application/json

{
    "iterations": 10,
    "max_workers": 2,
    "max_files": 5
}
```

Ответ:
```json
{
    "status": "success",
    "code": 200,
    "data": {
        "iterations": 10,
        "max_workers": 2,
        "max_files": 5
    }
}
```

### 2. Получение файла

```http
GET /file/{id}
```

Ответ:
```json
{
    "status": "success",
    "code": 200,
    "data": {
        "id": "123",
        "name": "file.txt",
        "content": "..."
    }
}
```

## Конфигурация

Сервис не использует конфигурационный файл. Все настройки передаются через эндпоинт `/generate`:

- `iterations` - количество итераций генерации
- `max_workers` - максимальное количество параллельных воркеров
- `max_files` - максимальное количество файлов для генерации

Пример минимальной конфигурации:
```json
{
    "iterations": 1,
    "max_workers": 1,
    "max_files": 1
}
```

