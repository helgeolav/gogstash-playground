package deletefile

import (
	"context"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
	"os"
)

// ModuleName is the name used in config file
const ModuleName = "deletefile"

// ErrorTag tag added to event when process module failed
const ErrorTag = "gogstash_filter_deletefile_error"

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig

	Field string `json:"field" yaml:"field"` // field name of file to delete
}

// DefaultFilterConfig returns an FilterConfig struct with default values
func DefaultFilterConfig() FilterConfig {
	return FilterConfig{
		FilterConfig: config.FilterConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Field: "file_name",
	}
}

// InitHandler initialize the filter plugin
func InitHandler(ctx context.Context, raw *config.ConfigRaw) (config.TypeFilterConfig, error) {
	conf := DefaultFilterConfig()
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}

// Event the main filter event
func (f *FilterConfig) Event(ctx context.Context, event logevent.LogEvent) (logevent.LogEvent, bool) {
	fn := event.Get(f.Field)
	if filename, ok := fn.(string); ok {
		err := os.Remove(filename)
		if err != nil {
			goglog.Logger.Errorf("%s: %s", ModuleName, err.Error())
			return event, false
		}
		goglog.Logger.Debugf("%s: deleted %s", ModuleName, filename)
		return event, true
	}
	goglog.Logger.Debugf("%s: no file deleted (missing name)", ModuleName)
	return event, false
}
