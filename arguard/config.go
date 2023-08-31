package arguard

import "flag"

type Config struct {
	ReportErrors bool
}

func NewConfig() Config {
	return Config{
		ReportErrors: false,
	}
}

func (c *Config) flagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("contracts", flag.ExitOnError)
	fs.BoolVar(
		&c.ReportErrors, "report-errors", c.ReportErrors,
		"show errors occurring during contract execution",
	)
	return fs
}
