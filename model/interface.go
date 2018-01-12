package model

import (
	log "github.com/kdar/factorlog"
)

var (
	logger *log.FactorLog
)

// Set to use a specific logger
func SetLogger(alogger *log.FactorLog) {
	logger = alogger
}

type CscErr struct {
	errMsg string
}

func (e CscErr) Error() string {
	return e.errMsg
}

func NewCscErr(err string) *CscErr {
	return &CscErr{
		errMsg: err,
	}
}
