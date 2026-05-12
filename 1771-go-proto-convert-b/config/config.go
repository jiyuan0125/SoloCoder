package config

type Rule struct {
	Name          string            `json:"name" yaml:"name"`
	SourceSystem  string            `json:"source_system" yaml:"source_system"`
	TargetSystem  string            `json:"target_system" yaml:"target_system"`
	FieldMappings []FieldMapping    `json:"field_mappings" yaml:"field_mappings"`
	ErrorStrategy string            `json:"error_strategy" yaml:"error_strategy"`
	Enabled       bool              `json:"enabled" yaml:"enabled"`
}

type FieldMapping struct {
	SourcePath      string      `json:"source_path" yaml:"source_path"`
	TargetPath      string      `json:"target_path" yaml:"target_path"`
	SourceType      string      `json:"source_type" yaml:"source_type"`
	TargetType      string      `json:"target_type" yaml:"target_type"`
	DefaultValue    interface{} `json:"default_value" yaml:"default_value"`
	Required        bool        `json:"required" yaml:"required"`
	Transform       *Transform  `json:"transform" yaml:"transform"`
	SkipIfMissing   bool        `json:"skip_if_missing" yaml:"skip_if_missing"`
}

type Transform struct {
	Type      string            `json:"type" yaml:"type"`
	Params    map[string]string `json:"params" yaml:"params"`
}

type Config struct {
	Port    string            `json:"port" yaml:"port"`
	Rules   []Rule            `json:"rules" yaml:"rules"`
}
