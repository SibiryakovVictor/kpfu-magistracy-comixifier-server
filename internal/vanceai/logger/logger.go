package logger

type Logger interface {
	Info(msg string, context map[string]interface{})
	Debug(msg string, context map[string]interface{})
	Error(msg string, context map[string]interface{})
}
