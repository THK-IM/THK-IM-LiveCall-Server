CREATE TABLE IF NOT EXISTS `stat` (
    `room_id`         String,
    `uid`             String,
    `stream_id`       String,
    `stream_key`      String,
    `stream_type`     Int64,
    `sfu_stream_key`  String,
    `c_time`          Int64,
    `p_size`          Int64,
    `h_size`          Int64,
    `p_count`         Int64,
    `p_lost_count`    Int64,
    `jitter`          Float64
)
Engine = MergeTree()
PRIMARY KEY (`room_id`, `uid`)
ORDER BY (`room_id`, `uid`, `c_time`)
SETTINGS index_granularity = 8192, index_granularity_bytes = 0;