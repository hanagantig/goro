func NewLogger() *slog.Logger {
  handler := slog.NewJSONHandler(os.Stdout, nil)

  return slog.New(handler) 
}
