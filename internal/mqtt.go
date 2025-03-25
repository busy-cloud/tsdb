package internal

import (
	"encoding/json"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"strings"
	"time"
)

func subscribe() {
	mqtt.Subscribe("device/+/+/property", func(topic string, payload []byte) {
		var values map[string]interface{}
		err := json.Unmarshal(payload, &values)
		if err != nil {
			log.Error(err)
			return
		}

		ss := strings.Split(topic, "/")
		pid := ss[1]
		id := ss[2]

		err = Write(pid, id, values, time.Now().UnixMilli())
		if err != nil {
			log.Error(err)
		}
	})
}
