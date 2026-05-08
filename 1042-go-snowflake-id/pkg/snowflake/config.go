package snowflake

type Config struct {
	Epoch           int64
	MachineID       int64
	AutoMachineID   bool
	MaxWaitMillis   int64
}

func DefaultConfig() *Config {
	return &Config{
		Epoch:         DefaultEpoch,
		AutoMachineID: true,
		MaxWaitMillis: DefaultMaxWait,
	}
}

func (c *Config) Validate() error {
	if c.Epoch < 0 {
		return ErrInvalidEpoch
	}
	if c.MachineID < 0 || c.MachineID > MaxMachineID {
		if !c.AutoMachineID {
			return ErrInvalidMachineID
		}
	}
	if c.MaxWaitMillis < 0 {
		return ErrInvalidMaxWait
	}
	return nil
}
