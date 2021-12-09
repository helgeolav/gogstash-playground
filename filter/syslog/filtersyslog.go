package syslog

import (
	"context"
	"errors"
	syslog "github.com/influxdata/go-syslog/v3"
	"github.com/influxdata/go-syslog/v3/rfc3164"
	"github.com/influxdata/go-syslog/v3/rfc5424"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
	"strings"
)

// ModuleName is the name used in config file
const ModuleName = "syslog"

const MessageField = "syslog_message" // default field for syslog message
const HostnameField = "hostname"      // default hostname
const AppNameField = "appname"        // default appname
const SeverityField = "severity"      // default severity
const PriorityField = "priority"      // default priority
const MessageIdField = "message_id"   // default message id

// ErrorTag tag added to event when process module failed
const ErrorTag = "gogstash_filter_syslog_error"

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig

	Source         string `json:"source" yaml:"source"`                 // source message field name
	Format         string `json:"format" yaml:"format"`                 // input format, either RFC3164 or RFC5424
	SaveTime       bool   `json:"save_time" yaml:"save_time"`           // if true time from syslog is kept
	RemoveSource   bool   `json:"remove_source" yaml:"remove_source"`   // if true source message is removed (upon success)
	MessageField   string `json:"message_field" yaml:"message_field"`   // syslog message
	HostnameField  string `json:"hostname_field" yaml:"hostname_field"` // hostname
	AppNameField   string `yaml:"app_name_field" json:"app_name_field"` // appname
	SeverityField  string `json:"severity_field" yaml:"severity_field"`
	PriorityField  string `json:"priority_field" yaml:"priority_field"`
	MessageIdField string `json:"message_id_field" yaml:"message_id_field"`

	parser syslog.Machine // the parser that parses messages
}

// DefaultFilterConfig returns an FilterConfig struct with default values
func DefaultFilterConfig() FilterConfig {
	return FilterConfig{
		FilterConfig: config.FilterConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Source:         "message",
		Format:         "RFC5424",
		MessageField:   MessageField,
		HostnameField:  HostnameField,
		AppNameField:   AppNameField,
		SeverityField:  SeverityField,
		PriorityField:  PriorityField,
		MessageIdField: MessageIdField,
	}
}

// InitHandler initialize the filter plugin
func InitHandler(ctx context.Context, raw config.ConfigRaw, control config.Control) (config.TypeFilterConfig, error) {
	conf := DefaultFilterConfig()
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}

	conf.Format = strings.ToUpper(conf.Format)
	switch conf.Format {
	case "RFC5424":
		conf.parser = rfc5424.NewParser()
	case "RFC3164":
		conf.parser = rfc3164.NewParser()
	default:
		return nil, errors.New("Invalid format")
	}

	return &conf, nil
}

// Event the main filter event
func (f *FilterConfig) Event(ctx context.Context, event logevent.LogEvent) (logevent.LogEvent, bool) {
	if value, ok := event.Get(f.Source).(string); ok {
		msg, err := f.parser.Parse([]byte(value))
		if err != nil {
			goglog.Logger.Errorf("%s: %s", ModuleName, err.Error())
			return event, false
		}
		if !msg.Valid() {
			goglog.Logger.Errorf("%s: invalid message", ModuleName)
			return event, false
		}
		err = f.setsyslogfields(msg, &event)
		if err != nil {
			goglog.Logger.Errorf("%s: %s", ModuleName, err.Error())
			return event, false
		}
	} else {
		event.AddTag(ErrorTag)
		return event, false
	}
	if f.RemoveSource {
		event.SetValue(f.Source, "")
		event.Remove(f.Source)
	}
	return event, true
}

// setsyslogfields add fields for each part of the message
func (f *FilterConfig) setsyslogfields(message syslog.Message, event *logevent.LogEvent) (err error) {
	var msg syslog.Base
	switch t := message.(type) {
	case *syslog.Base:
		msg = *t
	case *rfc5424.SyslogMessage:
		msg = t.Base
	case *rfc3164.SyslogMessage:
		msg = t.Base
	default:
		return errors.New("incorrect syslog data format")
	}
	// save time
	if f.SaveTime {
		if msg.Timestamp != nil {
			event.Timestamp = *msg.Timestamp
		}
	}
	// message
	if msg.Message != nil {
		event.SetValue(f.MessageField, *msg.Message)
	}
	// hostname
	if msg.Hostname != nil {
		event.SetValue(f.HostnameField, *msg.Hostname)
	}
	// app name
	if msg.Appname != nil {
		event.SetValue(f.AppNameField, *msg.Appname)
	}
	// severity
	if msg.Severity != nil {
		event.SetValue(f.SeverityField, *msg.Severity)
	}
	// priority
	if msg.Priority != nil {
		event.SetValue(f.PriorityField, *msg.Priority)
	}
	// message id
	if msg.MsgID != nil {
		event.SetValue(f.MessageIdField, *msg.MsgID)
	}
	return
}
