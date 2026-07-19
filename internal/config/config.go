package config

// Config holds runtime settings for the forum server.
type Config struct {
	Addr      string
	DBPath    string
	StaticDir string
}

// Default returns sensible local development defaults.
func Default() Config {
	return Config{
		Addr:      ":8080",
		DBPath:    "./forum.db",
		StaticDir: "./frontend/static",
	}
}
