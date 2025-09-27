# Doc Service

Сервис для работы с иерархией документов (root → subdoc → sub-subdoc).


## Подготовка базы данных

Сервис использует [Reindexer](https://github.com/Restream/reindexer).

1. Скачайте образ:
   ```bash
   docker pull reindexer/reindexer


Запустите контейнер:

docker run -p 9088:9088 -p 6534:6534 \
   -e RX_USER=admin \
   -e RX_PASSWORD=admin \
   -it reindexer/reindexer


## Запуск сервиса
В VS Code

    - Используйте launch.json конфиг:
        Run Doc Service
    Запуск производится через встроенный дебаггер.

    - Через bash скрипт
        Из корня проекта:
        bash scripts/run.sh

