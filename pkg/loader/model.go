package loader

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/thk-im/thk-im-base-server/conf"
	"github.com/thk-im/thk-im-base-server/snowflake"
	"gorm.io/gorm"
)

// LoadModels 与 sirius-server 一致：用 yaml Models 里 Name 与具体 Model 实现绑定
func LoadModels(modeConfigs []conf.Model, database *gorm.DB, logger *logrus.Entry, snowflakeNode *snowflake.Node) map[string]interface{} {
	modelMap := make(map[string]interface{})
	if database == nil {
		return modelMap
	}
	for _, ms := range modeConfigs {
		var m interface{}
		modelMap[ms.Name] = m
	}
	return modelMap
}

// LoadTables 与 sirius-server 一致：从 ./sql/{Name}.sql 建表，内容中表名需含 %s 作为分表后缀；Shards=1 时后缀为 ""。
func LoadTables(modeConfigs []conf.Model, database *gorm.DB) error {
	for _, ms := range modeConfigs {
		if ms.Name == "user_device" {
			continue
		}
		path := fmt.Sprintf("./sql/%s.sql", ms.Name)
		buffer, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if ms.Shards == 1 {
			sql := fmt.Sprintf(string(buffer), "")
			err = database.Exec(sql).Error
			if err != nil {
				return err
			}
		} else {
			for i := int64(0); i < ms.Shards; i++ {
				sql := fmt.Sprintf(string(buffer), fmt.Sprintf("%d", i))
				err = database.Exec(sql).Error
				if err != nil {
					return err
				}
			}
		}

	}
	return nil
}
