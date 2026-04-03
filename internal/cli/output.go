package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"go.yaml.in/yaml/v3"
)

func printOutput(cfg *Config, data any) error {
	switch cfg.Output {
	case OutputYAML:
		b, err := yaml.Marshal(data)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(os.Stdout, string(b))
		return err
	case OutputTable:
		return printTable(data)
	case OutputRaw:
		_, err := fmt.Fprintln(os.Stdout, data)
		return err
	default:
		b, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(os.Stdout, string(b))
		return err
	}
}

func printMeta(meta *ResponseMeta) error {
	if meta == nil {
		return nil
	}
	return printOutput(&Config{Output: OutputJSON}, map[string]any{
		"meta": meta,
	})
}

func printTable(v any) error {
	switch x := v.(type) {
	case []any:
		for _, row := range x {
			b, err := json.Marshal(row)
			if err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, string(b))
		}
		return nil
	default:
		b, err := json.Marshal(x)
		if err != nil {
			return err
		}
		s := string(b)
		s = strings.TrimSpace(s)
		_, err = fmt.Fprintln(os.Stdout, s)
		return err
	}
}

