package main

import (
	"bitbucket.org/HelgeOlav/utils/version"
	gogversion "github.com/tsaikd/KDGoLib/version"
	"github.com/tsaikd/gogstash/cmd"
	"github.com/tsaikd/gogstash/config/goglog"
)

func main() {
	version.NAME = "gogstash-playground"
	version.VERSION = gogversion.VERSION // copy gogstash version
	ver := version.Get()
	goglog.Logger.Infof("starting %s", ver.String())
	cmd.Module.MustMainRun()
}
