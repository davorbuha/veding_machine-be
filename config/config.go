package config

type Config struct {
	DSN string `env:"VEDING_MACHINE_PSQL_DSN"`
	Port string `env:"VEDING_MACHINE_PORT"`
}