package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	BaseURL            string
	Token              string
	ClientID           string
	UID                string
	Timeout            time.Duration
	Output             OutputFormat
	InsecureSkipVerify bool
	Quiet              bool
	ConfigFile         string
}

func bindConfig(cmd *cobra.Command, cfg *Config) error {
	v := viper.New()
	v.SetEnvPrefix("LIBRAREASE")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	if cfg.ConfigFile != "" {
		v.SetConfigFile(cfg.ConfigFile)
	} else {
		v.SetConfigName("librarease")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.config/librarease")
	}
	_ = v.ReadInConfig()

	if err := v.BindPFlags(cmd.PersistentFlags()); err != nil {
		return fmt.Errorf("bind flags: %w", err)
	}

	v.SetDefault("base-url", "https://localhost:8080")
	v.SetDefault("timeout", "30s")
	v.SetDefault("output", "json")

	cfg.BaseURL = strings.TrimRight(v.GetString("base-url"), "/")
	cfg.Token = v.GetString("token")
	cfg.ClientID = v.GetString("client-id")
	cfg.UID = v.GetString("uid")
	cfg.InsecureSkipVerify = v.GetBool("insecure-skip-verify")
	cfg.Quiet = v.GetBool("quiet")

	timeoutRaw := v.GetString("timeout")
	dur, err := time.ParseDuration(timeoutRaw)
	if err != nil {
		return fmt.Errorf("invalid timeout: %w", err)
	}
	cfg.Timeout = dur

	out, err := parseOutputFormat(v.GetString("output"))
	if err != nil {
		return err
	}
	cfg.Output = out
	return nil
}

