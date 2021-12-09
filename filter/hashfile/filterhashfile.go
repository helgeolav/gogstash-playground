package hashfile

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
	"io"
	"os"
)

// ModuleName is the name used in config file
const ModuleName = "hashfile"

const defaultBufSize = 20000 // default buffer size

// ErrorTag tag added to event when process module failed
const ErrorTag = "gogstash_filter_hashfile_error"

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig

	Field   string   `json:"field" yaml:"field"`       // field name of file to hash
	Output  string   `json:"output" yaml:"output"`     // field name to store output to
	Algos   []string `json:"algos" yaml:"algos"`       // hash algos to hash on
	BufSize int      `json:"buf_size" yaml:"buf_size"` // buffer size for copy
}

// DefaultFilterConfig returns an FilterConfig struct with default values
func DefaultFilterConfig() FilterConfig {
	var allHash []string
	for k, _ := range SupportedHashes {
		allHash = append(allHash, k)
	}
	return FilterConfig{
		FilterConfig: config.FilterConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		Field:   "file_name",
		Algos:   allHash,
		BufSize: defaultBufSize,
		Output:  "hash",
	}
}

// Hash is our interface to a hash. Each hash that implements this interface will work with the module.
type Hash interface {
	io.Writer
	Sum() []byte
}

// SupportedHashes the list of supported hashes and their init functions
var SupportedHashes map[string]func(interface{}) Hash = make(map[string]func(interface{}) Hash)

// InitHandler initialize the filter plugin
func InitHandler(ctx context.Context, raw config.ConfigRaw, control config.Control) (config.TypeFilterConfig, error) {
	conf := DefaultFilterConfig()
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}

	// check that all hashes are supported
	for _, v := range conf.Algos {
		if _, ok := SupportedHashes[v]; !ok {
			return &conf, fmt.Errorf("%s not supported", v)
		}
	}

	if conf.BufSize < 1 {
		conf.BufSize = defaultBufSize
	}
	return &conf, nil
}

// Event the main filter event
func (f *FilterConfig) Event(ctx context.Context, event logevent.LogEvent) (logevent.LogEvent, bool) {
	// init hashers
	hashers := []Hash{}
	for _, v := range f.Algos {
		hashers = append(hashers, SupportedHashes[v](nil))
	}
	// read file and hash
	fn := event.GetString(f.Field)
	fi, err := os.Open(fn)
	if err != nil {
		goglog.Logger.Errorf("%s: %s", ModuleName, err.Error())
		event.AddTag(ErrorTag)
		return event, false
	}
	defer fi.Close()
	input := bufio.NewReader(fi)
	buf := make([]byte, f.BufSize)
	for {
		n, err := input.Read(buf)
		if n > 0 {
			for x := 0; x < len(hashers); x++ {
				hashers[x].Write(buf[:n])
			}
			continue
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			goglog.Logger.Errorf("%s: %s", ModuleName, err.Error())
			return event, false
		}
	}
	// add result to event
	result := map[string][]byte{}
	for k, v := range f.Algos {
		result[v] = hashers[k].Sum()
	}
	event.SetValue(f.Output, result)
	return event, true
}
