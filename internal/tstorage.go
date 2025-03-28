package internal

import (
	"errors"
	"github.com/busy-cloud/boat/config"
	"github.com/nakabonne/tstorage"
	"github.com/spf13/cast"
	"regexp"
	"strconv"
	"time"
)

type Point struct {
	Value any   `json:"value"`
	Time  int64 `json:"time"`
}

var storage tstorage.Storage

func Open() error {
	//tstorage.WithTimestampPrecision("ms")

	var options []tstorage.Option
	options = append(options,
		tstorage.WithTimestampPrecision(tstorage.Seconds),                                      //精确到秒
		tstorage.WithRetention(time.Hour*24*time.Duration(config.GetInt(MODULE, "retention"))), //失效期1年，后续改配置文件
		tstorage.WithPartitionDuration(time.Hour*time.Duration(config.GetInt(MODULE, "partition"))),
		tstorage.WithDataPath(config.GetString(MODULE, "path")),
	)

	var err error
	storage, err = tstorage.NewStorage(options...)
	return err
}

func Close() error {
	return storage.Close()
}

func Write(metric, id string, values map[string]interface{}, tm int64) error {
	rows := make([]tstorage.Row, 0)
	for k, v := range values {
		rows = append(rows, tstorage.Row{
			Metric:    metric,
			Labels:    []tstorage.Label{{"key", k}, {"id", id}},
			DataPoint: tstorage.DataPoint{Value: cast.ToFloat64(v), Timestamp: tm},
		})
	}
	return storage.InsertRows(rows)
}

func Query(metric, id string, field string, start, end, window, method string) ([]*Point, error) {
	//相对时间转化为时间戳
	s, err := parseTimeEx(start)
	if err != nil {
		return nil, err
	}

	e, err := parseTimeEx(end)
	if err != nil {
		return nil, err
	}

	w, err := parseTime(window)
	if err != nil {
		return nil, err
	}
	points, err := storage.Select(metric, []tstorage.Label{{"key", field}, {"id", id}}, s, e) //ms
	if err != nil {
		//无数据
		if err == tstorage.ErrNoDataPoints {
			return nil, nil
		}
		return nil, err
	}

	switch method {
	case "mean":
		return mean(points, s, w), nil
	case "last":
		return last(points, s, w), nil
	default:
		return nil, errors.New("invalid method")
	}
}

func mean(points []*tstorage.DataPoint, start, window int64) []*Point {
	results := make([]*Point, 0)
	var total float64 = 0
	var count float64 = 0
	var timestamp int64

	for _, p := range points {
		//按窗口划分
		for p.Timestamp > start+window {
			start += window
			if count > 0 {
				results = append(results, &Point{
					Value: total / count,
					Time:  timestamp,
				})
				total = 0
				count = 0
			}
		}

		total += p.Value
		count++
		timestamp = p.Timestamp
	}

	//最后一组
	if count > 0 {
		results = append(results, &Point{
			Value: total / count,
			Time:  timestamp,
		})
	}
	return results
}

func last(points []*tstorage.DataPoint, start, window int64) []*Point {
	results := make([]*Point, 0)
	var timestamp int64

	var value float64 = 0
	var has bool = false

	for _, p := range points {
		//按窗口划分
		for p.Timestamp > start+window {
			start += window
			if has {
				results = append(results, &Point{
					Value: value,
					Time:  timestamp,
				})
				has = false
			}
		}

		has = true
		value = p.Value
		timestamp = p.Timestamp
	}
	//最后一组
	if has {
		results = append(results, &Point{
			Value: value,
			Time:  timestamp,
		})
	}

	return results
}

var timeReg *regexp.Regexp

func init() {
	timeReg = regexp.MustCompile(`^(-?\d+)(d|h|m|s)$`)
}

func parseTimeEx(tm string) (int64, error) {
	//标准日期串
	t, err := time.Parse(time.DateTime, tm)
	if err == nil {
		return t.UnixMilli(), nil
	}
	//
	//t, err = time.Parse(time.ANSIC, tm)
	//if err == nil {
	//	return t.UnixMilli(), nil
	//}
	//
	//t, err = time.Parse(time.UnixDate, tm)
	//if err == nil {
	//	return t.UnixMilli(), nil
	//}
	//
	//t, err = time.Parse(time.RubyDate, tm)
	//if err == nil {
	//	return t.UnixMilli(), nil
	//}

	t, err = time.Parse(time.RFC3339, tm)
	if err == nil {
		return t.UnixMilli(), nil
	}

	tt, err := parseTime(tm)
	if err != nil {
		return 0, err
	}
	return tt + time.Now().UnixMilli(), nil
}

func parseTime(tm string) (int64, error) {
	//tsdb格式
	ss := timeReg.FindStringSubmatch(tm)
	if ss == nil || len(ss) != 3 {
		return 0, errors.New("错误时间")
	}
	val, _ := strconv.ParseInt(ss[1], 10, 64)
	switch ss[2] {
	case "d":
		val *= 24 * 60 * 60 * 1000
	case "h":
		val *= 60 * 60 * 1000
	case "m":
		val *= 60 * 1000
	case "s":
		val *= 1000
	}
	return val, nil
}
