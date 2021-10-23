package downloadfile

import (
	"errors"
	"github.com/tsaikd/gogstash/config/logevent"
	"net/http"
	"os"
	"testing"
	"time"
)

func getTestEvent() logevent.LogEvent {
	event := logevent.LogEvent{
		Timestamp: time.Now(),
		Message:   "This is a message that should be ignored",
		Tags:      nil,
		Extra:     nil,
	}
	event.SetValue("url", "https://www.helge.net/images/IMG_2987.jpg")
	headers := make(map[string]string)
	headers["X-Request"] = "Testevent/go-test"
	event.SetValue("headers", headers)
	return event
}

const restrictAuthSite = "www.github.com" // used for testing below

func getAuthenticators() (result []Authenticator) {
	a1 := Authenticator{
		Name:       "test1",
		RestrictTo: []string{restrictAuthSite},
		Headers: map[string]string{
			"X-Auth-Name": "test1",
			"X-Else":      "else",
		},
	}
	a2 := Authenticator{
		Name:       "test2",
		RestrictTo: []string{restrictAuthSite},
		Headers: map[string]string{
			"X-Auth-Name": "test2",
			"X-What":      "what",
			"X-Sum":       "number-three",
		},
	}
	return append(result, a1, a2)
}

func TestFilterConfig_GetAuthenticatorHeaders(t *testing.T) {
	f := FilterConfig{
		Authenticators: getAuthenticators(),
	}
	type check struct {
		name    string // input name
		headers int    // number of headers expected
	}
	checks := []check{
		{
			name:    "test-of-a-header",
			headers: 0,
		}, {
			name:    "test1",
			headers: 2,
		}, {
			name:    "test2",
			headers: 3,
		},
	}
	for _, v := range checks {
		result := len(f.GetAuthenticatorHeaders(v.name, restrictAuthSite))
		if result != v.headers {
			t.Errorf("%s failed, expected %v but got %v", v.name, v.headers, result)
		}
	}
}

func TestFilterConfig_Download(t *testing.T) {
	event := getTestEvent()
	const fileSizeField = "filesize"
	f := FilterConfig{
		URL:            "url",
		Headers:        "headers",
		FileName:       "filename",
		Size:           fileSizeField,
		SuccessCodes:   []int{http.StatusOK},
		Authenticators: nil,
	}
	err := f.DownloadFile(&event)
	if err != nil {
		t.Error(err)
	}
	// get filename
	fn := event.GetString(f.FileName)
	if len(fn) == 0 {
		t.Error("No filename of saved file in event")
	} else {
		os.Remove(fn)
	}
	// validate that we have a size
	iSize, _ := event.GetValue(fileSizeField)
	if size, ok := iSize.(int64); ok {
		if size <= 0 {
			t.Error("Invalid size of file")
		}
	} else {
		t.Error("Could not determine size of file")
	}
}

func TestFilterConfig_ValidateEvent(t *testing.T) {
	event := getTestEvent()
	var err error
	f := FilterConfig{
		URL:            "url",
		Headers:        "headers",
		FileName:       "filename",
		Authenticators: nil,
	}
	// make sure first run is as expected
	err = f.ValidateEvent(&event)
	if err != nil {
		t.Errorf("First run: %s", err)
	}
	// empty URL
	event.SetValue("url", "")
	err = f.ValidateEvent(&event)
	if !errors.Is(err, errEmptyUrl) {
		t.Errorf("empty URL did not return error")
	}
	// invalid URL
	event.SetValue("url", "gogstash")
	err = f.ValidateEvent(&event)
	if err == nil {
		t.Errorf("invalid URL did not return error")
	}
}

func TestIsIn(t *testing.T) {
	type args struct {
		value int
		set   []int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"Test 1", args{value: 100, set: []int{99, 101}}, false},
		{"Test 2", args{value: 100, set: []int{}}, false},
		{"Test 3", args{value: 100, set: []int{99, 101, 100}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsIntIn(tt.args.value, tt.args.set); got != tt.want {
				t.Errorf("IsIntIn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterConfig_DownloadWithPwd(t *testing.T) {
	event := getTestEvent()
	event.SetValue("url", "https://www.helge.net/golang-test-protected/private.txt")
	const fileSizeField = "filesize"
	f := FilterConfig{
		URL:            "url",
		Headers:        "headers",
		FileName:       "filename",
		Size:           fileSizeField,
		SuccessCodes:   []int{http.StatusOK},
		Authenticators: nil,
	}
	err := f.DownloadFile(&event)
	if err == nil {
		t.Error("File was downloaded but should fail (server error)")
	}
	// add authentication and download
	myAuth := Authenticator{
		Name:       "admin-basic-auth",
		RestrictTo: []string{"www.helge.net"},
		Headers: map[string]string{
			"Authorization": "Basic YWRtaW46YWRtaW4=",
		},
	}
	f.Authenticators = append(f.Authenticators, myAuth)
	f.Auth = "auth-key"
	event.SetValue("auth-key", myAuth.Name)
	err = f.DownloadFile(&event)
	if err != nil {
		t.Error(err)
		return
	}
	// get filename
	fn := event.GetString(f.FileName)
	if len(fn) == 0 {
		t.Error("No filename of saved file in event")
	} else {
		os.Remove(fn)
	}
	// validate that we have a size
	iSize, _ := event.GetValue(fileSizeField)
	if size, ok := iSize.(int64); ok {
		if size <= 0 {
			t.Error("Invalid size of file")
		}
	} else {
		t.Error("Could not determine size of file")
	}
}
