package main

import (
	"github.com/helgeolav/gogstash-playground/codec/xml"
	"github.com/tsaikd/gogstash/config"
)

func init() {
	config.RegistCodecHandler(xml.ModuleName, xml.InitHandler)
}
