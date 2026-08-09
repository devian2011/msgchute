package auth

type Config struct {
	Plugin string         `yaml:"plugin"`
	Params map[string]any `yaml:"params"`
}
