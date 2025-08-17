# RandomCoffeeBotTelegram

## Как запустить
Создайте .env файл
TELEGRAM_BOT_TOKEN=<ваш токен>
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=
POSTGRES_DB=random_coffee

Запустите через Docker
```shell
docker-compose up --build
```

## Проверьте команды в Telegram:
/start — регистрация.
/match — распределение пар.

## Как остановить
```shell
docker compose down
```