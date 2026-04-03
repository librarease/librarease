package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestBindConfigFlags(t *testing.T) {
	cfg := &Config{}
	cmd := &cobra.Command{Use: "t"}
	cmd.PersistentFlags().String("config", "", "")
	cmd.PersistentFlags().String("base-url", "https://localhost:8080", "")
	cmd.PersistentFlags().String("token", "", "")
	cmd.PersistentFlags().String("client-id", "", "")
	cmd.PersistentFlags().String("uid", "", "")
	cmd.PersistentFlags().String("timeout", "30s", "")
	cmd.PersistentFlags().String("output", "json", "")
	cmd.PersistentFlags().Bool("insecure-skip-verify", false, "")
	cmd.PersistentFlags().Bool("quiet", false, "")

	if err := cmd.PersistentFlags().Set("base-url", "https://example.com"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.PersistentFlags().Set("output", "yaml"); err != nil {
		t.Fatal(err)
	}
	if err := bindConfig(cmd, cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.BaseURL != "https://example.com" {
		t.Fatalf("unexpected base-url: %s", cfg.BaseURL)
	}
	if cfg.Output != OutputYAML {
		t.Fatalf("unexpected output: %s", cfg.Output)
	}
}

