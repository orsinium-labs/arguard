package contracts

import "flag"

type Config struct {
	FollowImports bool
}

func NewConfig() Config {
	return Config{
		FollowImports: true,
	}
}

func (c *Config) flagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("contracts", flag.ExitOnError)
	fs.BoolVar(
		&c.FollowImports, "follow-imports", c.FollowImports,
		"extract contracts defined in the imported packages",
	)
	return fs
}
