docker run -p 9088:9088 -p 6534:6534 -e RX_USER=admin -e RX_PASSWORD=admin -it reindexer/reindexer - использовал для БД 
Предварительно: docker pull reindexer/reindexer
Конфиги уже содержат все необходимое для подключения к БД