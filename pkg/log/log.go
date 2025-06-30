package log

import (
	"encoding/json"
	"fmt"
	rfc7807 "maryan_api/pkg/problem"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Logger interface {
	Do(db *gorm.DB)
	SetProblem(problem rfc7807.Problem)
	SetError(err error, status int)
	GetID() string
}
type Log struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	Time        time.Time       `gorm:"autoCreateTime" json:"time"`
	IP          string          `gorm:"type:varchar(39);not null" json:"ip"`
	Route       string          `gorm:"type:varchar(255); not null" json:"route"`
	QueryParams json.RawMessage `gorm:"type:json; " json:"queryParams"`
	Headers     json.RawMessage `gorm:"type:json; not null" json:"headers"`
	Body        json.RawMessage `gorm:"type:json" json:"body"`
	Method      string          `gorm:"type:varchar(7);not null" json:"method"`
	Failed      bool            `gorm:"not null" json:"failed"`
	Type        string          `gorm:"type:varchar(255);not null" json:"type"`
	Title       string          `gorm:"type:varchar(255);not null" json:"title"`
	Status      int             `gorm:"type:smallint;not null" json:"status"`
	Detail      string          `gorm:"type:varchar(255);not null" json:"detail"`
	Extentions  json.RawMessage `gorm:"type:json" json:"extensions"`
}

func Migrate(db *gorm.DB) {
	err := db.AutoMigrate(&Log{})
	if err != nil {
		panic(err)
	}
}

func (l *Log) Do(db *gorm.DB) {
	if err := db.Create(&l).Error; err != nil {
		fmt.Printf(
			`\n
						...............LOGGING ERROR...................
						%s
						...............................................
						\n
					`, err.Error())
	}
}

func New(ip string, route string, queryParams, headers, body json.RawMessage, method string) Logger {
	return &Log{
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

	if len(problem.InvalidParams) != 0 {
		extensions, _ := json.Marshal(problem.InvalidParams)
		l.Extentions = extensions
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
