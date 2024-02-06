func NewLogger() (*zap.Logger, error) {
	return zap.NewProduction()
}
