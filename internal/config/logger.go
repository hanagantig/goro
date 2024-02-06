package config

var loggerPackges = map[Logger]string{
	"zap":  "\"go.uber.org/zap\"",
	"slog": "\"log/slog\"",
}

var serverHTTPStartFailMessage = map[Logger]string{
	"zap":  "a.logger.Fatal(\n\"Fail to start %s http server:\",\nzap.String(\"app\", a.cfg.App.Name),\nzap.Error(err),\n)",
	"slog": "a.logger.Error(\n\"Fail to start %s http server:\",\nslog.String(\"app\", a.cfg.App.Name),\nslog.String(\"err\", err.Error()),\n)",
}

var serverHTTPStopFailMessage = map[Logger]string{
	"zap":  "a.logger.Error(\"failed to gracefully shutdown server\", zap.Error(err))",
	"slog": "a.logger.Error(\"failed to gracefully shutdown server\", slog.String(\"err\", err.Error()))",
}

var examplePongHttpFailMessage = map[Logger]string{
	"zap":  "h.logger.Error(\"failed to encode json\", zap.Error(err))",
	"slog": "h.logger.Error(\"failed to encode json\", slog.String(\"err\", err.Error()))",
}

var supportedLoggers = map[Logger]struct{}{
	"zap":  {},
	"slog": {},
}

type Logger string

func (l Logger) String() string {
	return string(l)
}

func (l Logger) GetImportPackage() string {
	return loggerPackges[l]
}

func (l Logger) GetFailStartHTTPServerMessage() string {
	return serverHTTPStartFailMessage[l]
}

func (l Logger) GetFailStopHTTPServerMessage() string {
	return serverHTTPStopFailMessage[l]
}

func (l Logger) GetExamplePongHTTPFailMessage() string {
	return examplePongHttpFailMessage[l]
}
