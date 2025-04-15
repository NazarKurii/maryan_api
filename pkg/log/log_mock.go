package log

import (
	rfc7807 "maryan_api/pkg/problem"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LogMock struct {
}

func (l *LogMock) Do(db *gorm.DB) {

}

func (l *LogMock) SetProblem(problem rfc7807.Problem) {
	return
}

func (l *LogMock) SetError(err error, status int) {
	return
}

func (l LogMock) HTML() ([]byte, error) {
	return nil, nil
}

func (l LogMock) GetID() string {
	return uuid.NewString()
}

func Mock() Logger {
	return &LogMock{}
}
