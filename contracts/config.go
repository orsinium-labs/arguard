package contracts

import "flag"

type Config struct {
	FollowImports   bool
	ReportContracts bool
}

func NewConfig() Config {
	return Config{
		FollowImports:   true,
		ReportContracts: false,
	}
}

func (c *Config) flagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("contracts", flag.ExitOnError)
	fs.BoolVar(
		&c.FollowImports, "follow-imports", c.FollowImports,
		"extract contracts defined in the imported packages",
	)
	fs.BoolVar(
		&c.ReportContracts, "report-contracts", c.ReportContracts,
		"report all detected contracts, useful for debugging and testing",
	)
	return fs
}
