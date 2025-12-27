package app

import "go.uber.org/zap"

func NewZapLogger(env string) *zap.Logger {
	if env == "prod" {
		l, _ := zap.NewProduction()
		return l
	}
	l, _ := zap.NewDevelopment()
	return l
}

func ZapErr(err error) zap.Field       { return zap.Error(err) }
func ZapString(k, v string) zap.Field  { return zap.String(k, v) }
func ZapInt(k string, v int) zap.Field { return zap.Int(k, v) }
