package cli

import "fmt"

type OutputFormat string

const (
	OutputJSON  OutputFormat = "json"
	OutputYAML  OutputFormat = "yaml"
	OutputTable OutputFormat = "table"
	OutputRaw   OutputFormat = "raw"
)

func parseOutputFormat(v string) (OutputFormat, error) {
	switch OutputFormat(v) {
	case OutputJSON, OutputYAML, OutputTable, OutputRaw:
		return OutputFormat(v), nil
	default:
		return "", fmt.Errorf("invalid output format %q", v)
	}
}

type ResponseMeta struct {
	Total  int  `json:"total" yaml:"total"`
	Skip   int  `json:"skip" yaml:"skip"`
	Limit  int  `json:"limit" yaml:"limit"`
	Unread *int `json:"unread,omitempty" yaml:"unread,omitempty"`
}

type EnvelopeResponse struct {
	Data    any           `json:"data" yaml:"data"`
	Error   string        `json:"error,omitempty" yaml:"error,omitempty"`
	Message string        `json:"message,omitempty" yaml:"message,omitempty"`
	Meta    *ResponseMeta `json:"meta,omitempty" yaml:"meta,omitempty"`
}

type CommandSpec struct {
	Use                  string
	Short                string
	Long                 string
	Example              string
	Method               string
	Path                 string
	PathParamNames       []string
	QueryParams          []ParamSpec
	BodyParams           []ParamSpec
	ExpectEnvelope       bool
	ExpectHTML           bool
	NoContentSuccessText string
}

type ParamKind string

const (
	ParamString ParamKind = "string"
	ParamInt    ParamKind = "int"
	ParamBool   ParamKind = "bool"
	ParamSlice  ParamKind = "slice"
)

type ParamSpec struct {
	Name       string
	Kind       ParamKind
	Required   bool
	Positional bool
	Help       string
}

