package downloadfile

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/goglog"
	"github.com/tsaikd/gogstash/config/logevent"
	"io"
	"io/ioutil"
	"net/http"
	URL "net/url"
	"os"
)

// ModuleName is the name used in config file
const ModuleName = "downloadfile"

// ErrorTag tag added to event when process module failed
const ErrorTag = "gogstash_filter_downloadfile_error"

// Authenticator describes a way to authenticate to a web service by adding headers to the request
type Authenticator struct {
	Name       string            `json:"name" yaml:"name"`               // name of this authenticator
	Headers    map[string]string `json:"headers" yaml:"headers"`         // extra headers to add to request
	RestrictTo []string          `json:"restrict_to" yaml:"restrict_to"` // hostnames to limit this authenticator against
}

var (
	errEmptyUrl   = errors.New("empty url")
	errInvalidUrl = errors.New("invalid URL")
)

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig

	URL               string          `json:"URL" yaml:"URL"`                               // field name of file to download
	Headers           string          `json:"headers" yaml:"headers"`                       // field name of extra headers to send in request, must be map[string]string
	Auth              string          `json:"auth" yaml:"auth"`                             // field name of specific authentication to use in client
	Size              string          `json:"size" yaml:"size"`                             // field to store the file size in
	FileName          string          `json:"file_name" yaml:"file_name"`                   // field to store name of output file
	Response          string          `json:"response" yaml:"response"`                     // field to save response headers to
	DownloadDir       string          `json:"download_dir" yaml:"download_dir"`             // folder to download file to
	Authenticators    []Authenticator `json:"authenticators" yaml:"authenticators"`         // a list of preconfigured authenticators
	AuthenticatorFile string          `json:"authenticator_file" yaml:"authenticator_file"` // a file with authenticators to read
	SuccessCodes      []int           `json:"success_codes" yaml:"success_codes"`           // result codes that indicate success
	RetryCodes        []int           `json:"retry_codes" yaml:"retry_codes"`               // codes that indicate that this error can be retried
}

// DefaultFilterConfig returns an FilterConfig struct with default values
func DefaultFilterConfig() FilterConfig {
	return FilterConfig{
		FilterConfig: config.FilterConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
		URL:          "url",
		Headers:      "extra_headers",
		FileName:     "file_name",
		Size:         "file_size",
		SuccessCodes: []int{http.StatusOK},
		RetryCodes:   []int{http.StatusInternalServerError, http.StatusBadGateway, http.StatusGatewayTimeout, http.StatusLocked, http.StatusNotImplemented, http.StatusRequestTimeout, http.StatusServiceUnavailable, http.StatusTooManyRequests},
	}
}

// InitHandler initialize the filter plugin
func InitHandler(ctx context.Context, raw *config.ConfigRaw) (config.TypeFilterConfig, error) {
	conf := DefaultFilterConfig()
	err := config.ReflectConfig(raw, &conf)
	conf.Authenticators = append(conf.Authenticators, loadAuthenticators(conf.AuthenticatorFile)...)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}

// Event the main filter event
func (f *FilterConfig) Event(ctx context.Context, event logevent.LogEvent) (logevent.LogEvent, bool) {
	// validate
	err := f.ValidateEvent(&event)
	if err != nil {
		goglog.Logger.Errorf("%s: %s", ModuleName, err.Error())
		event.AddTag(ErrorTag)
		return event, false
	}
	err = f.DownloadFile(&event)
	if err != nil {
		goglog.Logger.Errorf("%s: %s", ModuleName, err.Error())
		event.AddTag(ErrorTag)
		return event, false
	}
	return event, true
}

// IsStringIn checks is a value is in the set
func IsStringIn(value string, set []string) bool {
	for _, v := range set {
		if value == v {
			return true
		}
	}
	return false
}

// IsIntIn checks if value is in set, returns true if true, false otherwise
func IsIntIn(value int, set []int) bool {
	for _, v := range set {
		if value == v {
			return true
		}
	}
	return false
}

// ValidateEvent is a pre-flight check that checks our input parameters to see if they are sane. An error is returned if they are not sane.
func (f *FilterConfig) ValidateEvent(event *logevent.LogEvent) error {
	url := event.GetString(f.URL)
	// check empty URL
	if len(url) == 0 {
		return errEmptyUrl
	}
	// check invalid url
	scheme, err := URL.ParseRequestURI(url)
	if err != nil {
		return err
	}
	if !(scheme.Scheme == "http" || scheme.Scheme == "https") {
		return errInvalidUrl
	}
	return nil
}

// DownloadFile downloads the file for the given event
func (f *FilterConfig) DownloadFile(event *logevent.LogEvent) error {
	url := event.GetString(f.URL)
	// prepare request
	client := &http.Client{}
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	rawHeaders := event.Get(f.Headers)
	if headers, ok := rawHeaders.(map[string]string); ok {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}
	rawAuthHeaders := f.GetAuthenticatorHeaders(event.GetString(f.Auth), req.URL.Hostname())
	for k, v := range rawAuthHeaders {
		req.Header.Set(k, v)
	}
	// now get file from URL
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	// check if ok
	if !IsIntIn(res.StatusCode, f.SuccessCodes) {
		// TODO: should add something if the error is retryable?
		return fmt.Errorf("%s downloaded %s, got HTTP status %v", ModuleName, url, res.StatusCode)
	}
	// and save it
	outputFile, err := ioutil.TempFile(f.DownloadDir, "gogstash-")
	if err != nil {
		return err
	}
	defer outputFile.Close()
	numBytes, err := io.Copy(outputFile, res.Body)
	savedFile := outputFile.Name()
	if err != nil {
		os.Remove(savedFile)
		return err
	}

	goglog.Logger.Debugf("%s downloaded %s to %s (size %v)", ModuleName, url, savedFile, numBytes)
	event.SetValue(f.FileName, savedFile)
	event.SetValue(f.Size, numBytes)
	if len(f.Response) > 0 {
		event.SetValue(f.Response, res.Header)
	}
	return nil
}

// GetAuthenticatorHeaders returns an empty map or the content of this authenticator.
// host is compared against allowed hostnames and both name and host must match if something is to be returned.
func (f *FilterConfig) GetAuthenticatorHeaders(name string, host string) map[string]string {
	for _, v := range f.Authenticators {
		if v.Name == name && IsStringIn(host, v.RestrictTo) {
			return v.Headers
		}
	}
	return make(map[string]string)
}

// loadAuthenticators attempts to get a set of authenticators from file, by first looking at the specified input, then the environment variable FILTERDOWNLOAD_AUTHENTICATOR
// and then trying to load the file authenticator.json from the current directory. All files are loaded in this order.
func loadAuthenticators(fn string) (result []Authenticator) {
	if len(fn) > 0 {
		result = loadAuthenticatorsFile(fn)
	}
	fn = os.Getenv("FILTERDOWNLOAD_AUTHENTICATOR")
	if len(fn) > 0 {
		result = append(result, loadAuthenticatorsFile(fn)...)
	}
	result = append(result, loadAuthenticatorsFile("authenticator.json")...)
	return
}

// loadAuthenticatorsFile internal - reads the file from disk, does not return any error (only logging them), only empty set
func loadAuthenticatorsFile(fn string) (result []Authenticator) {
	bytes, err := os.ReadFile(fn)
	if err != nil {
		//goglog.Logger.Info(err.Error())
		return
	}
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		goglog.Logger.Info(err.Error())
	}
	return
}
