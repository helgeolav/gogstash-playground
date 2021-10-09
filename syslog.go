package main

import (
	"github.com/helgeolav/gogstash-playground/filter/syslog"
	"github.com/tsaikd/gogstash/config"
)

// init registers syslog filter
func init() {
	config.RegistFilterHandler(syslog.ModuleName, syslog.InitHandler)
}
