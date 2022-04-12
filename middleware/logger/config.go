package logger

import (
	"io"
	"os"
	"time"
)

type Config struct {
	Format string
	//Default: 2006-01-02 15:04:05
	TimeFormat string
	//Default: "Local"
	TimeZone string
	// Default: os.Stdout
	Output io.Writer

	timeZoneLocation *time.Location
}

var DefaultConfig = Config{
	Format: `{"time":"${time_custom}","id":"${request_id}","remote_ip":"${remote_ip}",` +
		`"host":"${host}","method":"${method}","uri":"${uri}",` +
		`"status":${status},"error":"${error}","latency":${latency},"latency_human":"${latency_human}"` +
		`,"bytes_in":${bytes_in},"bytes_out":${bytes_out}}` + "\n",
	TimeFormat: "2006-01-02 15:04:05",
	TimeZone:   "Local",
	Output:     os.Stdout,
}

func configDefault(config ...Config) Config {
	if len(config) == 0 {
		return DefaultConfig
	}
	cfg := config[0]
	if cfg.Format == "" {
		cfg.Format = DefaultConfig.Format
	}
	if cfg.TimeZone == "" {
		cfg.TimeZone = DefaultConfig.TimeZone
	}
	if cfg.TimeFormat == "" {
		cfg.TimeFormat = DefaultConfig.TimeFormat
	}
	if cfg.Output == nil {
		cfg.Output = DefaultConfig.Output
	}
	return cfg
}
