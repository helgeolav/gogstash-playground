package downloadfile

import (
	"context"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "downloadfile"

// ErrorTag tag added to event when process module failed
const ErrorTag = "gogstash_filter_downloadfile_error"

// Authenticator describes a way to authenticate to a web service
type Authenticator struct {
	Name    string            // name of this authenticator
	Headers map[string]string // extra headers to add on request
}

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig

	URL            string          `json:"URL" yaml:"URL"`                       // field name of file to download
	Headers        string          `json:"headers" yaml:"headers"`               // field name of extra headers to send in request, must be map[string]string
	Auth           string          `json:"auth" yaml:"auth"`                     // field name of specific authentication to use in client
	DownloadDir    string          `json:"download_dir" yaml:"download_dir"`     // folder to download file to
	Authenticators []Authenticator `json:"authenticators" yaml:"authenticators"` // a list of preconfigured authenticators
}

// DefaultFilterConfig returns an FilterConfig struct with default values
func DefaultFilterConfig() FilterConfig {
	return FilterConfig{
		FilterConfig: config.FilterConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		URL:         "url",
		Headers:     "extra_headers",
		DownloadDir: "/var/tmp",
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
	result := f.ValidateEvent(&event)
	return event, result == nil
}

// ValidateEvent is a pre-flight check that checks our input parameters to see if they are sane. An error is returned if they are not sane.
func (f *FilterConfig) ValidateEvent(event *logevent.LogEvent) error {

}
