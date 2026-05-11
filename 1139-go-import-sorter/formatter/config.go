package formatter

type Config struct {
	LineWidth      int
	RemoveEmptyLines bool
	ModuleName     string
}

func DefaultConfig() Config {
	return Config{
		LineWidth:      120,
		RemoveEmptyLines: true,
		ModuleName:     "",
	}
}

type Stats struct {
	LinesModified   int
	ImportsMoved    int
	LinesSplit      int
}
