package loader

import (
	"context"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/sirupsen/logrus"
	"github.com/thk-im/thk-im-livecall-server/pkg/conf"
	"github.com/thk-im/thk-im-livecall-server/pkg/service/stat"
	"os"
	"strings"
	"time"
)

func LoadStatService(source *conf.Stat, logger *logrus.Entry) stat.Service {
	if source == nil || source.ClickHouse == nil {
		return nil
	}
	ck := loadClickhouse(source.ClickHouse)
	return stat.NewStatService(ck, logger)
}

func loadClickhouse(source *conf.ClickHouse) driver.Conn {
	addresses := strings.Split(source.Url, ",")
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: addresses,
		Auth: clickhouse.Auth{
			Database: source.Db,
			Username: source.UserName,
			Password: source.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: time.Second * 30,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
	})
	if err != nil {
		panic(err)
	}
	var (
		e      error
		buffer []byte
	)
	buffer, e = os.ReadFile("deploy/table.sql")
	if e == nil {
		e = conn.Exec(context.Background(), string(buffer))
		if e != nil {
			return nil
		}
	}
	return conn
}
