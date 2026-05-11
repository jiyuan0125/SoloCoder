package codegen

type YAMLDefinition struct {
	Errors []ErrorSpec `yaml:"errors"`
}

type ErrorSpec struct {
	Name       string `yaml:"name"`
	Code       int    `yaml:"code"`
	Message    string `yaml:"message"`
	HTTPStatus int    `yaml:"http_status"`
	GRPCCode   string `yaml:"grpc_code"`
	Category   string `yaml:"category"`
	Deprecated bool   `yaml:"deprecated"`
}

type ValidatedError struct {
	Name       string
	Code       int
	Message    string
	HTTPStatus int
	GRPCCode   string
	Category   string
	Deprecated bool
}

type GenerateResult struct {
	Code       string
	Warnings   []string
}
