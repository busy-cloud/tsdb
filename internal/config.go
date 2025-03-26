package internal

import (
	"github.com/busy-cloud/boat/config"
	"github.com/busy-cloud/boat/smart"
)

const MODULE = "influxdb"

func init() {
	config.SetDefault(MODULE, "enable", false)
	config.SetDefault(MODULE, "path", "tsdb")
	config.SetDefault(MODULE, "retention", 365)
	config.SetDefault(MODULE, "partition", 60*60)

	config.Register(MODULE, &config.Form{
		Title:  "TSDB时序数据库配置",
		Module: MODULE,
		Form: smart.Form{
			Fields: []smart.Field{
				{Key: "enable", Label: "启用", Type: "switch"},
				{Key: "path", Label: "保存路径", Type: "text"},
				{Key: "retention", Label: "失效期（天）", Type: "number"},
				{Key: "partition", Label: "数据分区（小时）", Type: "number"},
			},
		},
	})
}
