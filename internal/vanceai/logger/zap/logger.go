package zap

import "go.uber.org/zap"

type Logger struct {
	sugar *zap.SugaredLogger
}

func NewLogger(sugar *zap.SugaredLogger) *Logger {
	return &Logger{sugar: sugar}
}

func (l *Logger) Info(msg string, context map[string]interface{}) {
	l.sugar.Infow(msg, l.composeContext(context)...)
}

func (l *Logger) Debug(msg string, context map[string]interface{}) {
	l.sugar.Debugw(msg, l.composeContext(context)...)
}

func (l *Logger) Error(msg string, context map[string]interface{}) {
	l.sugar.Errorw(msg, l.composeContext(context)...)
}

func (l *Logger) composeContext(c map[string]interface{}) []interface{} {
	i := 0
	cList := make([]interface{}, len(c)*2, len(c)*2)
	for k, v := range c {
		cList[i] = k
		i++
		cList[i] = v
		i++
	}

	return cList
}
