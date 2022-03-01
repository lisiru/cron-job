package logger

import (
	"encoding/json"
	"go.uber.org/zap/zapcore"
)

const (
	consoleFormat ="console"
	jsonFormat = "json"
)

type Options struct {
	OutputPaths []string `json:"output-paths" mapstructure:"output-paths"`
	ErrorOutputPaths []string `json:"error-output-paths" mapstructure:"error-output-paths"`
	Level string `json:"level" mapstructure:"json"`
	Format string `json:"format" mapstructure:"format"`
	DisableCaller bool `json:"disable-caller" mapstructure:"disable-caller"`
	DisableStacktrace bool `json:"disable-stacktrace" mapstructure:"disable-stacktrace"`
	EnableColor bool `json:"enable-color" mapstructure:"enable-color"`
	Development bool `json:"development" mapstructure:"development"`
	Name string `json:"name" mapstructure:"name"`
}

func NewOptions() *Options  {
	return &Options{
		Level: zapcore.InfoLevel.String(),
		DisableCaller: false,
		DisableStacktrace: false,
		Format: consoleFormat,
		EnableColor: false,
		Development: false,
		OutputPaths: []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
}

func (o *Options)  String() string {
	data,_:=json.Marshal(o)
	return string(data)
}