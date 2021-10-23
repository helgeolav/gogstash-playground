package main

import (
	"github.com/helgeolav/gogstash-playground/filter/deletefile"
	"github.com/helgeolav/gogstash-playground/filter/downloadfile"
	"github.com/tsaikd/gogstash/config"
)

// init registers syslog filter
func init() {
	config.RegistFilterHandler(deletefile.ModuleName, deletefile.InitHandler)
	config.RegistFilterHandler(downloadfile.ModuleName, downloadfile.InitHandler)
}
