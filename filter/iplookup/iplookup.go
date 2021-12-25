package iplookup

import (
	"bitbucket.org/HelgeOlav/geoiplookup"
	"context"
	"errors"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
	"github.com/tsaikd/gogstash/filter/ip2location"
	"google.golang.org/grpc"
	"net"
)

// ModuleName is the name used in config file
const ModuleName = "iplookup"

// defaultTimeoutMS is the default timeout in milliseconds
const defaultTimeoutMS = 2000

// ErrorTag tag added to event when process module failed
const ErrorTag = "iplookup_error"

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig
	Server      string   `json:"server" yaml:"server"`             // server to connect to
	Timeout     uint     `json:"timeout" yaml:"timeout"`           // timeout in milliseconds
	Key         string   `json:"key" yaml:"key"`                   // field to save the result to
	IPField     string   `json:"ip_field" yaml:"ip_field"`         // IP field to get ip address from
	SkipPrivate bool     `json:"skip_private" yaml:"skip_private"` // skip private IP addresses
	PrivateNet  []string `json:"private_net" yaml:"private_net"`   // list of own defined private IP addresses

	privateCIDRs []*net.IPNet     // our parsed private ranges
	conn         *grpc.ClientConn // our connection object
}

// DefaultFilterConfig returns an FilterConfig struct with default values
func DefaultFilterConfig() FilterConfig {
	return FilterConfig{
		FilterConfig: config.FilterConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Key:         "iplookup",
		Timeout:     defaultTimeoutMS,
		SkipPrivate: true,
		PrivateNet:  filterip2location.DefaultCIDR,
	}
}

// InitHandler initialize the filter plugin
func InitHandler(ctx context.Context, raw config.ConfigRaw, control config.Control) (config.TypeFilterConfig, error) {
	conf := DefaultFilterConfig()
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}

	if conf.Timeout == 0 {
		conf.Timeout = defaultTimeoutMS
	}
	if len(conf.IPField) == 0 {
		return nil, errors.New("missing ip_field")
	}
	if len(conf.Server) == 0 {
		return nil, errors.New("missing server")
	}
	conf.conn, err = grpc.Dial(conf.Server, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	for _, cidr := range conf.PrivateNet {
		_, privateCIDR, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, err
		}
		conf.privateCIDRs = append(conf.privateCIDRs, privateCIDR)
	}

	return &conf, nil
}

// privateIP checks if IP in list of configured networks
func (f *FilterConfig) privateIP(ip net.IP) bool {
	for idx := range f.privateCIDRs {
		if f.privateCIDRs[idx].Contains(ip) {
			return true
		}
	}
	return false
}

// Event the main filter event
func (f *FilterConfig) Event(ctx context.Context, event logevent.LogEvent) (logevent.LogEvent, bool) {
	ipstr := event.GetString(f.IPField)
	if ipstr == "" {
		// Passthru if empty
		return event, false
	}
	ip := net.ParseIP(ipstr)
	if ip == nil || (f.SkipPrivate && f.privateIP(ip)) {
		// Passthru
		return event, false
	}
	c := geoiplookup.NewGeoIpLookupClient(f.conn)
	req := geoiplookup.GeoIpRequest{Ip: ipstr}
	result, err := c.Lookup(ctx, &req)
	if err == nil {
		event.SetValue(f.Key, *result)
	} else {
		event.AddTag(ErrorTag)
		goglog.Logger.Error(err.Error())
		return event, false
	}
	return event, true
}
