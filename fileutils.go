package main

import (
	"github.com/helgeolav/gogstash-playground/filter/deletefile"
	"github.com/helgeolav/gogstash-playground/filter/downloadfile"
	"github.com/helgeolav/gogstash-playground/filter/hashfile"
	"github.com/helgeolav/gogstash-playground/input/debugfile"
	"github.com/tsaikd/gogstash/config"
)

// init registers syslog filter
func init() {
	config.RegistFilterHandler(deletefile.ModuleName, deletefile.InitHandler)
	config.RegistFilterHandler(downloadfile.ModuleName, downloadfile.InitHandler)
	config.RegistFilterHandler(hashfile.ModuleName, hashfile.InitHandler)
	config.RegistInputHandler(debugfile.ModuleName, debugfile.InitHandler)
}
