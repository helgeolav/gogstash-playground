package debugfile

import (
	"context"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
	"io/ioutil"
	"time"
)

// ModuleName is the name used in config file
const ModuleName = "debugfile"

// InputConfig holds the configuration json fields and internal objects
type InputConfig struct {
	config.InputConfig

	Input    string `json:"input" yaml:"input"`       // input file to read from
	First    int    `json:"first" yaml:"first"`       // delay for first run
	Interval int    `json:"interval" yaml:"interval"` // interval between each run
}

// DefaultInputConfig returns an InputConfig struct with default values
func DefaultInputConfig() InputConfig {
	return InputConfig{
		InputConfig: config.InputConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Interval: 30,
	}
}

// InitHandler initialize the input plugin
func InitHandler(ctx context.Context, raw config.ConfigRaw, control config.Control) (config.TypeInputConfig, error) {
	conf := DefaultInputConfig()
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}

	conf.Codec, err = config.GetCodecOrDefault(ctx, raw)
	return &conf, err
}

// Start wraps the actual function starting the plugin
func (t *InputConfig) Start(ctx context.Context, msgChan chan<- logevent.LogEvent) (err error) {
	outputmsg, err := ioutil.ReadFile(t.Input)
	if err != nil {
		return err
	}
	if t.First < 1 {
		t.First = 1
	}
	ticker := time.NewTicker(time.Second * time.Duration(t.First))
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			_, err = t.Codec.Decode(ctx, outputmsg, nil, []string{ModuleName}, msgChan)
			if err != nil {
				goglog.Logger.Errorf("%s: %s", ModuleName, err.Error())
			}
			ticker.Reset(time.Second * time.Duration(t.Interval))
		}
	}
	return nil
}
