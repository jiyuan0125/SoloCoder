package crawler

type Config struct {
	StartURL       string
	MaxDepth       int
	Concurrency    int
	RateLimit      int
	Timeout        int
	OutputDir      string
	UserAgent      string
}

func (c *Config) Validate() {
	if c.Concurrency > 100 {
		c.Concurrency = 100
	}
	if c.RateLimit <= 0 {
		c.RateLimit = 2
	}
	if c.OutputDir == "" {
		c.OutputDir = "./crawled"
	}
	if c.UserAgent == "" {
		c.UserAgent = "Mozilla/5.0 (compatible; GoCrawler/1.0)"
	}
}
