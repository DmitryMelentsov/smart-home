# Контроллер умного дома

Это учебный проект, который был разработан в рамках курса по Go от T-Банка.

Контроллер умного дома это сервис, который предоставляет интерфейс для мониторинга состояния систем дома. Он принимает информацию от датчиков, сохраняет ее, может отдать информацию по запросу.

## Технологии:

- Go
- PostgreSQL
- Docker

## Подготовка окружения

1. Установить docker ([windows](https://docs.docker.com/desktop/install/windows-install/), [Mac](https://docs.docker.com/desktop/install/mac-install/), [Linux](https://docs.docker.com/desktop/install/linux-install/))
    * Если установили не docker-desktop, а docker отдельно - необходимо установить [docker-compose](https://docs.docker.com/compose/install/)
2. Установить [migrate](https://github.com/golang-migrate/migrate/blob/master/cmd/migrate/README.md)
3. Базу данных можно развернуть с помощью docker-compose (файл в корне проекта). Для этого необходимо выполнить команду `docker-compose up -d`. После того, как она запустится, к ней можно подключаться - `postgres://postgres:postgres@127.0.0.1:5432/db`.
4. Для миграции нужно выполнить команду `migrate -path=./migrations -database postgres://postgres:postgres@127.0.0.1:5432/db?sslmode=disable up`. Также к проекту приложен Makefile, с помощью которого тоже можно выполнить миграцию - `make migrate-up`.

Если решили выполнить миграцию через Make (`make migrate-up`) на Windows - его нужно [установить](https://stackoverflow.com/questions/32127524/how-to-install-and-use-make-in-windows). В Mac и Linux установка не требуется.

## Запуск приложения

Для запуска приложения требуется [переменная окружения](https://gobyexample.com/environment-variables) `DATABASE_URL` - URL подключения к базе (`postgres://postgres:postgres@127.0.0.1:5432/db?sslmode=disable`).

## Запуск тестов

Тесты в процессе запуска используют docker. Убедитесь, что он у вас запущен.

1. зайти в терминале в каталог с домашним заданием
2. вызвать ```go test -v ./... -race```
