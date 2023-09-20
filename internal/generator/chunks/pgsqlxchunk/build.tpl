func (a *App) newPgSqlxConnect(cfg config.SQLConfig) (*sqlx.DB, error) {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("host=%s port=%s ", cfg.Host, cfg.Port))
	builder.WriteString(fmt.Sprintf("user=%s password=%s ", cfg.User, cfg.Password))
	builder.WriteString(fmt.Sprintf("dbname=%s ", cfg.DBName))
	builder.WriteString(fmt.Sprintf("timezone=%s ", cfg.Timezone))
	builder.WriteString("sslmode=disable ")

	params := builder.String()

	db, err := sqlx.Open("postgres", params)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)

	return db, err
}