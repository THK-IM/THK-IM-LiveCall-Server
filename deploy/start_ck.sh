docker run -d --rm --name clickhouse-server --ulimit nofile=262144:262144 \
-p 18123:8123 -p 19000:9000 -p 19009:9009 \
-v /home/data/clickhouse/data:/var/lib/clickhouse:rw \
-v /home/data/clickhouse/log:/var/log/clickhouse-server:rw \
-v /home/data/clickhouse/etc:/etc/clickhouse-server:rw \
clickhouse/clickhouse-server:23.4.2.11-alpine