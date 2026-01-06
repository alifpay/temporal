# temporal

Демо‑сервис “платёж” на базе Temporal.

**Что делает**
- Принимает HTTP запрос `POST /pay`.
- Стартует Temporal workflow `PaymentWorkflow` в task queue `payment`.
- В workflow:
	- Получает курс USD на дату платежа (`GetUSDRate`) и пересчитывает сумму.
	- Отправляет платёж в RabbitMQ (`SendPayment`).
	- Ждёт сигнал со статусом платежа (`StatusSignal`).
	- При `success` вызывает activity `UpdatePaymentStatus`.

**Архитектура**
- HTTP сервер: `infra/server/http.go`
- Workflow/activities: `workflows/payment/*`
- Temporal client и HTTP handler: `app/*`
- RabbitMQ publish/consume: `infra/rmq/*`

## Требования
- Temporal Server доступен на `localhost:7233` (значение по умолчанию в SDK).
- RabbitMQ (в коде подключение задано в `infra/rmq/rabbitmq.go`).

## Запуск

1) Запустить зависимости (Temporal Server и RabbitMQ).

2) Запустить сервис:

```bash
go run .
```

По умолчанию HTTP сервер слушает `:8989` (см. `main.go`).

## HTTP API

### `POST /pay`

Тело запроса — JSON:

```json
{
	"ID": "payment-123",
	"Currency": "USD",
	"Amount": 100.5,
	"PayDate": "2025-01-20T00:00:00Z"
}
```

Пример:

```bash
curl -i \
	-X POST http://localhost:8989/pay \
	-H 'Content-Type: application/json' \
	-d '{"ID":"payment-123","Currency":"USD","Amount":100.5,"PayDate":"2025-01-20T00:00:00Z"}'
```

Ответ: `202 Accepted`.

## Статусы платежей

Статус прилетает из RabbitMQ (routing key `status.key`) и прокидывается в workflow сигналом `StatusSignal`.
Модель статуса описана в `models.PaymentStatus`.
