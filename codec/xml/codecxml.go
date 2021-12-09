package xml

import (
	"context"
	"errors"
	mjx "github.com/clbanning/mxj/v2"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
	"time"
)

// ModuleName is the name used in config file
const ModuleName = "xml"

// ErrorTag tag added to event when process module failed
const ErrorTag = "gogstash_codec_xml_error"

var NotImplementedError = errors.New("not implemented")

// Codec default struct for XML
type Codec struct {
	config.CodecConfig
	TryCast bool   `json:"try_cast" yaml:"try_cast"` // attempt to cast types from string
	RootTag string `json:"root_tag" yaml:"root_tag"` // root tag on encoder
}

// InitHandler initialize the codec plugin
func InitHandler(_ context.Context, raw config.ConfigRaw) (config.TypeCodecConfig, error) {
	c := &Codec{
		CodecConfig: config.CodecConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
	}
	if res, ok := (raw)["try_cast"].(bool); ok {
		c.TryCast = res
	}
	if res, ok := (raw)["root_tag"].(string); ok {
		c.RootTag = res
	}
	return c, nil
}

// Decode returns an event from 'data' as XML format
func (c *Codec) Decode(_ context.Context, data interface{}, eventExtra map[string]interface{}, tags []string, msgChan chan<- logevent.LogEvent) (ok bool, err error) {
	event := logevent.LogEvent{
		Timestamp: time.Now(),
		Extra:     eventExtra,
	}
	event.AddTag(tags...)
	// identify incoming message
	var m mjx.Map
	switch data.(type) {
	case string:
		m, err = mjx.NewMapXml([]byte(data.(string)), c.TryCast)
	case []byte:
		m, err = mjx.NewMapXml(data.([]byte), c.TryCast)
	default:
		err = errors.New("unsupported input format")
	}
	if err != nil {
		return false, err
	}
	event.Extra = m
	msgChan <- event
	return true, nil
}

// DecodeEvent decodes 'data' as XML format to event
func (c *Codec) DecodeEvent(data []byte, event *logevent.LogEvent) (err error) {
	return NotImplementedError
}

// Encode encodes the event to a XML encoded message
func (c *Codec) Encode(_ context.Context, event logevent.LogEvent, dataChan chan<- []byte) (ok bool, err error) {
	json, err := event.MarshalJSON()
	if err != nil {
		return false, err
	}
	m, err := mjx.NewMapJson(json)
	if err != nil {
		return false, err
	}
	output, err := m.Xml(c.RootTag)
	if err != nil {
		return false, err
	}
	dataChan <- output
	return true, nil
}
