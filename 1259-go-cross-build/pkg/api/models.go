package api

type Platform struct {
	GOOS   string `json:"goos"`
	GOARCH string `json:"goarch"`
}

type BuildTarget struct {
	Platform   Platform            `json:"platform"`
	Ldflags    string              `json:"ldflags,omitempty"`
	Env        map[string]string   `json:"env,omitempty"`
	OutputName string              `json:"output_name,omitempty"`
}

type BuildConfig struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Targets     []BuildTarget          `json:"targets"`
	ReleaseMode bool                   `json:"release_mode"`
	OutputDir   string                 `json:"output_dir,omitempty"`
	MainPackage string                 `json:"main_package"`
}

type BuildRequest struct {
	Config      BuildConfig    `json:"config"`
	BuildTarget *BuildTarget   `json:"build_target,omitempty"`
}

type BuildResult struct {
	Target         BuildTarget `json:"target"`
	Success        bool        `json:"success"`
	OutputPath     string      `json:"output_path,omitempty"`
	FileSize       int64       `json:"file_size,omitempty"`
	Checksum       string      `json:"checksum,omitempty"`
	DebugStripped  bool        `json:"debug_stripped"`
	HasCGODependency bool      `json:"has_cgo_dependency"`
	ErrorMessage   string      `json:"error_message,omitempty"`
	Warnings       []string    `json:"warnings,omitempty"`
}

type BuildResponse struct {
	Results       []BuildResult `json:"results"`
	TotalTargets  int           `json:"total_targets"`
	SuccessCount  int           `json:"success_count"`
	FailureCount  int           `json:"failure_count"`
}

type CommandResult struct {
	Target     BuildTarget `json:"target"`
	Command    string      `json:"command"`
	EnvVars    []string    `json:"env_vars"`
	OutputFile string      `json:"output_file"`
}

type ScriptRequest struct {
	Config     BuildConfig `json:"config"`
	Format     string      `json:"format"`
}

type ScriptResponse struct {
	Content string `json:"content"`
	Format  string `json:"format"`
}

type ValidatePlatformRequest struct {
	Platforms []Platform `json:"platforms"`
}

type ValidatePlatformResponse struct {
	Results map[string]bool `json:"results"`
	Errors  []string        `json:"errors,omitempty"`
}
