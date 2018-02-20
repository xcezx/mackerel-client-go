package mackerel

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func pfloat64(x float64) *float64 {
	return &x
}

func puint64(x uint64) *uint64 {
	return &x
}

func TestFindMonitors(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/api/v0/monitors" {
			t.Error("request URL should be /api/v0/monitors but :", req.URL.Path)
		}

		respJSON, _ := json.Marshal(map[string][]map[string]interface{}{
			"monitors": {
				{
					"id":            "2cSZzK3XfmG",
					"type":          "connectivity",
					"memo":          "connectivity monitor",
					"scopes":        []string{},
					"excludeScopes": []string{},
				},
				{
					"id":                              "2c5bLca8d",
					"type":                            "external",
					"name":                            "testMonitorExternal",
					"memo":                            "this monitor checks example.com.",
					"method":                          "GET",
					"url":                             "https://www.example.com/",
					"maxCheckAttempts":                3,
					"service":                         "someService",
					"notificationInterval":            60,
					"responseTimeCritical":            5000,
					"responseTimeWarning":             10000,
					"responseTimeDuration":            5,
					"certificationExpirationCritical": 15,
					"certificationExpirationWarning":  30,
					"containsString":                  "Foo Bar Baz",
					"skipCertificateVerification":     true,
					"headers": []map[string]interface{}{
						{"name": "Cache-Control", "value": "no-cache"},
					},
				},
				{
					"id":         "2DujfcR2kA9",
					"name":       "expression test",
					"memo":       "a monitor for expression",
					"type":       "expression",
					"expression": "avg(roleSlots('service:role','loadavg5'))",
					"operator":   ">",
					"warning":    20,
					"critical":   30,
				},
			},
		})

		res.Header()["Content-Type"] = []string{"application/json"}
		fmt.Fprint(res, string(respJSON))
	}))
	defer ts.Close()

	client, _ := NewClientWithOptions("dummy-key", ts.URL, false)
	monitors, err := client.FindMonitors()

	if err != nil {
		t.Error("err shoud be nil but: ", err)
	}

	{
		m, ok := monitors[0].(*MonitorConnectivity)
		if !ok || m.Type != "connectivity" {
			t.Error("request sends json including type but: ", m)
		}
		if m.Memo != "connectivity monitor" {
			t.Error("request sends json including memo but: ", m)
		}
	}

	{
		m, ok := monitors[1].(*MonitorExternalHTTP)
		if !ok || m.Type != "external" {
			t.Error("request sends json including type but: ", m)
		}
		if m.Memo != "this monitor checks example.com." {
			t.Error("request sends json including memo but: ", m)
		}
		if m.Service != "someService" {
			t.Error("request sends json including service but: ", m)
		}
		if m.NotificationInterval != 60 {
			t.Error("request sends json including notificationInterval but: ", m)
		}

		if m.URL != "https://www.example.com/" {
			t.Error("request sends json including url but: ", m)
		}
		if m.MaxCheckAttempts != 3 {
			t.Error("request sends json including maxCheckAttempts but: ", m)
		}
		if *m.ResponseTimeCritical != 5000 {
			t.Error("request sends json including responseTimeCritical but: ", m)
		}

		if *m.ResponseTimeWarning != 10000 {
			t.Error("request sends json including responseTimeWarning but: ", m)
		}

		if *m.ResponseTimeDuration != 5 {
			t.Error("request sends json including responseTimeDuration but: ", m)
		}

		if *m.CertificationExpirationCritical != 15 {
			t.Error("request sends json including certificationExpirationCritical but: ", m)
		}

		if *m.CertificationExpirationWarning != 30 {
			t.Error("request sends json including certificationExpirationWarning but: ", m)
		}

		if m.ContainsString != "Foo Bar Baz" {
			t.Error("request sends json including containsString but: ", m)
		}

		if m.SkipCertificateVerification != true {
			t.Error("request sends json including skipCertificateVerification but: ", m)
		}

		if !reflect.DeepEqual(m.Headers, []HeaderField{{Name: "Cache-Control", Value: "no-cache"}}) {
			t.Error("request sends json including headers but: ", m)
		}
	}

	{
		m, ok := monitors[2].(*MonitorExpression)
		if !ok || m.Type != "expression" {
			t.Error("request sends json including expression but: ", monitors[2])
		}
		if m.Memo != "a monitor for expression" {
			t.Error("request sends json including memo but: ", m)
		}
	}
}

// ensure that it supports `"headers":[]` and headers must be nil by default.
func TestMonitorExternalHTTP_headers(t *testing.T) {
	tests := []struct {
		name string
		in   *MonitorExternalHTTP
		want string
	}{
		{
			name: "default",
			in:   &MonitorExternalHTTP{},
			want: `{"headers":null}`,
		},
		{
			name: "empty list",
			in:   &MonitorExternalHTTP{Headers: []HeaderField{}},
			want: `{"headers":[]}`,
		},
	}

	for _, tt := range tests {
		b, err := json.Marshal(tt.in)
		if err != nil {
			t.Error(err)
			continue
		}
		if got := string(b); got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}

const monitorsjson = `
{
  "monitors": [
    {
      "id": "2cSZzK3XfmA",
      "type": "connectivity",
      "scopes": [],
      "excludeScopes": []
    },
    {
      "id"  : "2cSZzK3XfmB",
      "type": "host",
      "name": "disk.aa-00.writes.delta",
      "duration": 3,
      "metric": "disk.aa-00.writes.delta",
      "operator": ">",
      "warning": 20000.0,
      "critical": 400000.0,
      "maxCheckAttempts": 3,
      "scopes": [
        "Hatena-Blog"
      ],
      "excludeScopes": [
        "Hatena-Bookmark: db-master"
      ]
    },
    {
      "id"  : "2cSZzK3XfmF",
      "type": "host",
      "name": "Foo Bar",
      "duration": 3,
      "metric": "custom.foo.bar",
      "operator": ">",
      "warning": 200.0,
      "maxCheckAttempts": 5
    },
    {
      "id"  : "2cSZzK3XfmC",
      "type": "service",
      "name": "Hatena-Blog - access_num.4xx_count",
      "service": "Hatena-Blog",
      "duration": 1,
      "metric": "access_num.4xx_count",
      "operator": ">",
      "warning": 50.0,
      "critical": 100.0,
      "maxCheckAttempts": 5,
      "notificationInterval": 60
    },
    {
      "id"  : "2cSZzK3XfmG",
      "type": "service",
      "name": "Hatena-Blog - access_num.5xx_count",
      "service": "Hatena-Blog",
      "duration": 3,
      "metric": "access_num.5xx_count",
      "operator": ">",
      "critical": 100.0,
      "maxCheckAttempts": 3,
      "notificationInterval": 60
    },
    {
      "id"  : "2cSZzK3XfmD",
      "type": "external",
      "name": "example.com",
      "method": "POST",
      "url": "https://example.com",
      "service": "Hatena-Blog",
      "headers": [{"name":"Cache-Control", "value":"no-cache"}],
      "requestBody": "Request Body",
      "maxCheckAttempts": 7,
      "responseTimeCritical": 3000,
      "responseTimeWarning": 2000,
      "responseTimeDuration": 7,
      "certificationExpirationCritical": 60,
      "certificationExpirationWarning": 90
    },
    {
      "id"  : "2cSZzK3XfmH",
      "type": "external",
      "name": "example.com",
      "method": "GET",
      "url": "https://example.com",
      "service": "Hatena-Blog",
      "headers": [{"name":"Cache-Control", "value":"no-cache"}],
      "requestBody": "Request Body",
      "maxCheckAttempts": 5,
      "responseTimeWarning": 3000,
      "responseTimeDuration": 7,
      "certificationExpirationCritical": 30
    },
    {
      "id"  : "2cSZzK3XfmE",
      "type": "expression",
      "name": "role average",
      "expression": "avg(roleSlots(\"server:role\",\"loadavg5\"))",
      "operator": ">",
      "warning": 5.0,
      "critical": 10.0,
      "notificationInterval": 60
    }
  ]
}
`

var wantMonitors = []Monitor{
	&MonitorConnectivity{
		ID:                   "2cSZzK3XfmA",
		Name:                 "",
		Type:                 "connectivity",
		IsMute:               false,
		NotificationInterval: 0,
		Scopes:               []string{},
		ExcludeScopes:        []string{},
	},
	&MonitorHostMetric{
		ID:                   "2cSZzK3XfmB",
		Name:                 "disk.aa-00.writes.delta",
		Type:                 "host",
		IsMute:               false,
		NotificationInterval: 0,
		Metric:               "disk.aa-00.writes.delta",
		Operator:             ">",
		Warning:              pfloat64(20000.000000),
		Critical:             pfloat64(400000.000000),
		Duration:             3,
		MaxCheckAttempts:     3,
		Scopes: []string{
			"Hatena-Blog",
		},
		ExcludeScopes: []string{
			"Hatena-Bookmark: db-master",
		},
	},
	&MonitorHostMetric{
		ID:                   "2cSZzK3XfmF",
		Name:                 "Foo Bar",
		Type:                 "host",
		IsMute:               false,
		NotificationInterval: 0,
		Metric:               "custom.foo.bar",
		Operator:             ">",
		Warning:              pfloat64(200.0),
		Critical:             nil,
		Duration:             3,
		MaxCheckAttempts:     5,
		Scopes:               nil,
		ExcludeScopes:        nil,
	},
	&MonitorServiceMetric{
		ID:                   "2cSZzK3XfmC",
		Name:                 "Hatena-Blog - access_num.4xx_count",
		Type:                 "service",
		IsMute:               false,
		NotificationInterval: 60,
		Service:              "Hatena-Blog",
		Metric:               "access_num.4xx_count",
		Operator:             ">",
		Warning:              pfloat64(50.000000),
		Critical:             pfloat64(100.000000),
		Duration:             1,
		MaxCheckAttempts:     5,
	},
	&MonitorServiceMetric{
		ID:                   "2cSZzK3XfmG",
		Name:                 "Hatena-Blog - access_num.5xx_count",
		Type:                 "service",
		IsMute:               false,
		NotificationInterval: 60,
		Service:              "Hatena-Blog",
		Metric:               "access_num.5xx_count",
		Operator:             ">",
		Warning:              nil,
		Critical:             pfloat64(100.000000),
		Duration:             3,
		MaxCheckAttempts:     3,
	},
	&MonitorExternalHTTP{
		ID:                              "2cSZzK3XfmD",
		Name:                            "example.com",
		Type:                            "external",
		IsMute:                          false,
		NotificationInterval:            0,
		Method:                          "POST",
		URL:                             "https://example.com",
		MaxCheckAttempts:                7,
		Service:                         "Hatena-Blog",
		ResponseTimeCritical:            pfloat64(3000.0),
		ResponseTimeWarning:             pfloat64(2000.0),
		ResponseTimeDuration:            puint64(7),
		RequestBody:                     "Request Body",
		ContainsString:                  "",
		CertificationExpirationCritical: puint64(60),
		CertificationExpirationWarning:  puint64(90),
		SkipCertificateVerification:     false,
		Headers: []HeaderField{
			{
				Name:  "Cache-Control",
				Value: "no-cache",
			},
		},
	},
	&MonitorExternalHTTP{
		ID:                              "2cSZzK3XfmH",
		Name:                            "example.com",
		Type:                            "external",
		IsMute:                          false,
		NotificationInterval:            0,
		Method:                          "GET",
		URL:                             "https://example.com",
		MaxCheckAttempts:                5,
		Service:                         "Hatena-Blog",
		ResponseTimeCritical:            nil,
		ResponseTimeWarning:             pfloat64(3000.0),
		ResponseTimeDuration:            puint64(7),
		RequestBody:                     "Request Body",
		ContainsString:                  "",
		CertificationExpirationCritical: puint64(30),
		CertificationExpirationWarning:  nil,
		SkipCertificateVerification:     false,
		Headers: []HeaderField{
			{
				Name:  "Cache-Control",
				Value: "no-cache",
			},
		},
	},
	&MonitorExpression{
		ID:                   "2cSZzK3XfmE",
		Name:                 "role average",
		Type:                 "expression",
		IsMute:               false,
		NotificationInterval: 60,
		Expression:           "avg(roleSlots(\"server:role\",\"loadavg5\"))",
		Operator:             ">",
		Warning:              pfloat64(5.000000),
		Critical:             pfloat64(10.000000),
	},
}

func TestDecodeMonitor(t *testing.T) {
	if got := decodeMonitorsJSON(t); !reflect.DeepEqual(got, wantMonitors) {
		t.Errorf("fail to get correct data: diff: (-got +want)\n%v", pretty.Compare(got, wantMonitors))
	}
}

func BenchmarkDecodeMonitor(b *testing.B) {
	for i := 0; i < b.N; i++ {
		decodeMonitorsJSON(b)
	}
}

func decodeMonitorsJSON(t testing.TB) []Monitor {
	var data struct {
		Monitors []json.RawMessage `json:"monitors"`
	}
	if err := json.NewDecoder(strings.NewReader(monitorsjson)).Decode(&data); err != nil {
		t.Error(err)
	}
	ms := make([]Monitor, 0, len(data.Monitors))
	for _, rawmes := range data.Monitors {
		m, err := decodeMonitor(rawmes)
		if err != nil {
			t.Error(err)
		}
		ms = append(ms, m)
	}
	return ms
}

var monitorsToBeEncoded = []Monitor{
	&MonitorHostMetric{
		ID:       "2cSZzK3XfmA",
		Warning:  pfloat64(0.000000),
		Critical: pfloat64(400000.000000),
	},
	&MonitorHostMetric{
		ID:      "2cSZzK3XfmB",
		Warning: pfloat64(600000.000000),
	},
	&MonitorHostMetric{
		ID:       "2cSZzK3XfmB",
		Critical: pfloat64(500000.000000),
	},
	&MonitorServiceMetric{
		ID:       "2cSZzK3XfmC",
		Warning:  pfloat64(50.000000),
		Critical: pfloat64(0.000000),
	},
	&MonitorServiceMetric{
		ID:      "2cSZzK3XfmC",
		Warning: pfloat64(50.000000),
	},
	&MonitorServiceMetric{
		ID:       "2cSZzK3XfmC",
		Critical: pfloat64(0.000000),
	},
	&MonitorExpression{
		ID:       "2cSZzK3XfmE",
		Warning:  pfloat64(0.000000),
		Critical: pfloat64(0.000000),
	},
	&MonitorExpression{
		ID:      "2cSZzK3XfmE",
		Warning: pfloat64(0.000000),
	},
	&MonitorExpression{
		ID:       "2cSZzK3XfmE",
		Critical: pfloat64(0.000000),
	},
}

func TestEncodeMonitor(t *testing.T) {
	b, err := json.MarshalIndent(monitorsToBeEncoded, "", "    ")
	if err != nil {
		t.Error("err shoud be nil but: ", err)
	}

	want := `[
    {
        "id": "2cSZzK3XfmA",
        "warning": 0,
        "critical": 400000
    },
    {
        "id": "2cSZzK3XfmB",
        "warning": 600000,
        "critical": null
    },
    {
        "id": "2cSZzK3XfmB",
        "warning": null,
        "critical": 500000
    },
    {
        "id": "2cSZzK3XfmC",
        "warning": 50,
        "critical": 0
    },
    {
        "id": "2cSZzK3XfmC",
        "warning": 50,
        "critical": null
    },
    {
        "id": "2cSZzK3XfmC",
        "warning": null,
        "critical": 0
    },
    {
        "id": "2cSZzK3XfmE",
        "warning": 0,
        "critical": 0
    },
    {
        "id": "2cSZzK3XfmE",
        "warning": 0,
        "critical": null
    },
    {
        "id": "2cSZzK3XfmE",
        "warning": null,
        "critical": 0
    }
]`
	if got := string(b); got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
