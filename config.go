package gqfsl

import "time"

type Config struct {
	Email EmailConf
	Cron  CronConf
	Sql   SqlConf
}

// EmailConf ...
type EmailConf struct {
	// Host represents the host of the SMTP server.
	Host string
	// Port represents the port of the SMTP server.
	Port int
	// Username is the username to use to authenticate to the SMTP server.
	Username string
	// Password is the password to use to authenticate to the SMTP server.
	Password string
}

// CronConf ...
type CronConf struct {
	// DurationRetry - duration retry send
	DurationRetry time.Duration

	// SaveSend - save sent, default false
	SaveSend bool
	// DurationSaveSuccess - duration save success, default 0
	DurationSaveSuccess time.Duration
	// DurationSaveFailMessage - duration save fail message, default 24h
	DurationSaveFailMessage time.Duration
	// CountTry - count try send message, default 100
	CountTry int
	// KeepingTheOrder - keeping the order, default false
	//
	// if true - until the current one is sent, the following ones are blocked
	KeepingTheOrder bool
}

// SqlConf ...
type SqlConf struct {
	// BaseName, name base, default "email.sqlite"
	BaseName string
	Path     string
	//
	Timeout time.Duration
}

const (
	defaultCronCountTry                = 100
	defaultCronDurationSaveFailMessage = 24 * time.Hour
	defaultCronDurationRetry           = 5 * time.Minute
	defaultSqlBaseName                 = "email.sqlite"
	defaultSqlTimeout                  = 5 * time.Second
	defaultSqlPath                     = "."
)

func checkDefaultConfig(cgf Config) Config {

	// Cron
	if cgf.Cron.DurationRetry == 0 {
		cgf.Cron.DurationRetry = defaultCronDurationRetry
	}
	if cgf.Cron.DurationSaveFailMessage == 0 {
		cgf.Cron.DurationSaveFailMessage = defaultCronDurationSaveFailMessage
	}
	if cgf.Cron.CountTry == 0 {
		cgf.Cron.CountTry = defaultCronCountTry
	}

	// Sql
	if cgf.Sql.BaseName == "" {
		cgf.Sql.BaseName = defaultSqlBaseName
	}
	if cgf.Sql.Timeout == 0 {
		cgf.Sql.Timeout = defaultSqlTimeout
	}
	if cgf.Sql.Path == "" {
		cgf.Sql.Path = defaultSqlPath
	}

	return cgf
}
