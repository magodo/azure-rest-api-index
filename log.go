package main

import (
	"github.com/hashicorp/go-hclog"
)

var logger Logger

func SetLogger(l Logger) {
	logger = l
}

type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

type DefaultLogger hclog.Logger
