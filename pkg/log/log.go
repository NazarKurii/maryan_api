package log

import (
	"bytes"
	"encoding/json"
	"errors"
	rfc7807 "maryan_api/pkg/problem"
	"text/template"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Logger interface {
	Do(db *gorm.DB) error
	SetProblem(problem rfc7807.Problem)
	SetError(err error, status int)
	GetID() string
	HTML() ([]byte, error)
}
type Log struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	Time        time.Time       `gorm:"autoCreateTime" json:"time"`
	IP          string          `gorm:"type:varchar(39);not null" json:"ip"`
	Route       string          `gorm:"type:varchar(255); not null" json:"route"`
	QueryParams json.RawMessage `gorm:"type:json; " json:"queryParams"`
	Headers     json.RawMessage `gorm:"type:json; not null" json:"headers"`
	Body        json.RawMessage `gorm:"type:json; " json:"body"`
	Method      string          `gorm:"type:varchar(7);not null" json:"method"`
	Failed      bool            `gorm:"not null" json:"failed"`
	Type        string          `gorm:"type:varchar(255);not null" json:"type"`
	Title       string          `gorm:"type:varchar(255);not null" json:"title"`
	Status      int             `gorm:"type:smallint;not null" json:"status"`
	Detail      string          `gorm:"type:varchar(255);not null" json:"detail"`
	Extentions  json.RawMessage `gorm:"type:json" json:"extensions"`
}

func (l *Log) Do(db *gorm.DB) error {
	if len(l.IP) < 2 {
		return errors.New("Could not define the IP of the request")
	}

	if l.Route == "" {
		return errors.New("Could not define the route of the request")
	}

	return db.Create(&l).Error
}

func New(ip string, route string, queryParams, headers, body json.RawMessage, method string) Log {
	return Log{
		ID:          uuid.New(),
		Time:        time.Now(),
		IP:          ip,
		Route:       route,
		QueryParams: queryParams,
		Headers:     headers,
		Body:        body,
		Method:      method,
	}
}

func (l *Log) SetProblem(problem rfc7807.Problem) {
	l.Failed = true
	l.Type = problem.Type
	l.Title = problem.Title
	l.Status = problem.Status
	l.Detail = problem.Detail

	if extsJSON := problem.MarshalExtensions(); extsJSON != nil {
		l.Extentions = append(l.Extentions, extsJSON...)
	}

}

func (l *Log) SetError(err error, status int) {
	l.Failed = true
	l.Type = "Unknown"
	l.Title = "Unknown"
	l.Status = status
	l.Detail = err.Error()
}

func (l *Log) GetID() string {
	return l.ID.String()
}

func (log Log) HTML() ([]byte, error) {
	const tpl = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Log Detail</title>
    <style>
        body {
            font-family: 'Segoe UI', sans-serif;
            background-color: #f9f9f9;
            color: #333;
            padding: 2rem;
            max-width: 900px;
            margin: auto;
        }
        h1 {
            text-align: center;
            color: #444;
        }
        .log-section {
            margin-bottom: 2rem;
        }
        .log-section h2 {
            color: #666;
            font-size: 1.2rem;
            border-bottom: 1px solid #ccc;
            padding-bottom: 0.2rem;
            margin-bottom: 0.5rem;
        }
        .log-section pre {
            background-color: #272822;
            color: #f8f8f2;
            padding: 1rem;
            border-radius: 5px;
            overflow-x: auto;
        }
        .log-entry {
            background: #fff;
            border-radius: 8px;
            padding: 1rem;
            box-shadow: 0 0 10px rgba(0,0,0,0.05);
        }
        .log-entry p {
            margin: 0.5rem 0;
        }
        .label {
            font-weight: bold;
        }
    </style>
</head>
<body>
    <h1>Log Entry</h1>
    <div class="log-entry">
        <p><span class="label">ID:</span> {{.ID}}</p>
        <p><span class="label">Time:</span> {{.Time}}</p>
        <p><span class="label">IP:</span> {{.IP}}</p>
        <p><span class="label">Route:</span> {{.Route}}</p>
        <p><span class="label">Method:</span> {{.Method}}</p>
        <p><span class="label">Failed:</span> {{.Failed}}</p>
        <p><span class="label">Status:</span> {{.Status}}</p>
        <p><span class="label">Type:</span> {{.Type}}</p>
        <p><span class="label">Title:</span> {{.Title}}</p>
        <p><span class="label">Detail:</span> {{.Detail}}</p>

        <div class="log-section">
            <h2>Headers</h2>
            <pre><code>{{.PrettyHeaders}}</code></pre>
        </div>

        <div class="log-section">
            <h2>Query Params</h2>
            <pre><code>{{.PrettyQueryParams}}</code></pre>
        </div>

        <div class="log-section">
            <h2>Body</h2>
            <pre><code>{{.PrettyBody}}</code></pre>
        </div>

        <div class="log-section">
            <h2>Extensions</h2>
            <pre><code>{{.PrettyExtensions}}</code></pre>
        </div>
    </div>
</body>
</html>
`

	type tmplLog struct {
		Log
		PrettyHeaders     string
		PrettyQueryParams string
		PrettyBody        string
		PrettyExtensions  string
	}

	// Helper to pretty-print JSON fields
	pp := func(raw json.RawMessage) string {
		var buf bytes.Buffer
		if len(raw) == 0 {
			return "{}"
		}
		err := json.Indent(&buf, raw, "", "  ")
		if err != nil {
			return string(raw)
		}
		return buf.String()
	}

	tmplData := tmplLog{
		Log:               log,
		PrettyHeaders:     pp(log.Headers),
		PrettyQueryParams: pp(log.QueryParams),
		PrettyBody:        pp(log.Body),
		PrettyExtensions:  pp(log.Extentions),
	}

	t, err := template.New("log").Parse(tpl)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	err = t.Execute(&out, tmplData)
	if err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}
