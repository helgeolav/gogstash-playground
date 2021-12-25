package main

import (
	"github.com/helgeolav/gogstash-playground/filter/iplookup"
	"github.com/tsaikd/gogstash/config"
)

func init() {
	config.RegistFilterHandler(iplookup.ModuleName, iplookup.InitHandler)
}
