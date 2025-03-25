package internal

import (
	"github.com/busy-cloud/boat/boot"
)

func init() {
	boot.Register("tsdb", &boot.Task{
		Startup:  Startup,
		Shutdown: Close,
		Depends:  []string{"log", "mqtt", "database"},
	})
}

func Startup() error {
	subscribe()
	return Open()
}
