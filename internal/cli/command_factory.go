package cli

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type commandBuilder struct {
	spec CommandSpec
	cfg  *Config
	api  *APIClient
}

func newCommandFromSpec(spec CommandSpec, cfg *Config, api *APIClient) *cobra.Command {
	b := &commandBuilder{spec: spec, cfg: cfg, api: api}
	cmd := &cobra.Command{
		Use:     spec.Use,
		Short:   spec.Short,
		Long:    spec.Long,
		Example: spec.Example,
		Args:    cobra.ExactArgs(len(spec.PathParamNames)),
		RunE:    b.run,
	}
	b.bindFlags(cmd)
	return cmd
}

type flagValue struct {
	stringVal string
	intVal    int
	boolVal   bool
	sliceVal  []string
	changed   bool
}

func (b *commandBuilder) bindFlags(cmd *cobra.Command) {
	for _, p := range b.spec.QueryParams {
		b.bindParamFlag(cmd, p)
	}
	for _, p := range b.spec.BodyParams {
		b.bindParamFlag(cmd, p)
	}
}

func (b *commandBuilder) bindParamFlag(cmd *cobra.Command, p ParamSpec) {
	switch p.Kind {
	case ParamInt:
		cmd.Flags().Int(p.Flag, 0, p.Help)
	case ParamBool:
		cmd.Flags().Bool(p.Flag, false, p.Help)
	case ParamSlice:
		cmd.Flags().StringSlice(p.Flag, nil, p.Help)
	default:
		cmd.Flags().String(p.Flag, "", p.Help)
	}
	if p.Required {
		_ = cmd.MarkFlagRequired(p.Flag)
	}
}

func (b *commandBuilder) run(cmd *cobra.Command, args []string) error {
	path := b.spec.Path
	for i, n := range b.spec.PathParamNames {
		path = strings.ReplaceAll(path, "{"+n+"}", args[i])
	}

	query := map[string][]string{}
	body := map[string]any{}

	for _, p := range b.spec.QueryParams {
		val, err := b.getFlagValue(cmd, p)
		if err != nil {
			return err
		}
		if !val.changed {
			continue
		}
		switch p.Kind {
		case ParamInt:
			query[p.Name] = []string{fmt.Sprintf("%d", val.intVal)}
		case ParamBool:
			query[p.Name] = []string{fmt.Sprintf("%t", val.boolVal)}
		case ParamSlice:
			query[p.Name] = val.sliceVal
		default:
			query[p.Name] = []string{val.stringVal}
		}
	}

	for _, p := range b.spec.BodyParams {
		val, err := b.getFlagValue(cmd, p)
		if err != nil {
			return err
		}
		if !val.changed {
			continue
		}
		switch p.Kind {
		case ParamInt:
			body[p.Name] = val.intVal
		case ParamBool:
			body[p.Name] = val.boolVal
		case ParamSlice:
			body[p.Name] = val.sliceVal
		default:
			body[p.Name] = val.stringVal
		}
	}

	var bodyObj any
	if len(body) > 0 {
		bodyObj = body
	}

	resp, data, err := b.api.Do(b.spec.Method, path, query, bodyObj)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return httpError(resp, data)
	}

	if resp.StatusCode == 204 {
		if !b.cfg.Quiet {
			msg := b.spec.NoContentSuccessText
			if msg == "" {
				msg = "success"
			}
			fmt.Fprintln(cmd.OutOrStdout(), msg)
		}
		return nil
	}

	if b.spec.ExpectHTML {
		if b.cfg.Output == OutputRaw {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return err
		}
		if b.cfg.Output == OutputTable {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return err
		}
		return printOutput(b.cfg, map[string]any{"html": string(data)})
	}

	if b.spec.ExpectEnvelope {
		env, err := decodeEnvelope(data)
		if err != nil {
			return err
		}
		if env.Error != "" {
			return fmt.Errorf("%s", env.Error)
		}
		if err := printOutput(b.cfg, env.Data); err != nil {
			return err
		}
		if env.Meta != nil {
			return printMeta(env.Meta)
		}
		if env.Message != "" && !b.cfg.Quiet {
			_, err = fmt.Fprintln(cmd.OutOrStdout(), env.Message)
			return err
		}
		return nil
	}

	var anyJSON any
	if err := json.Unmarshal(data, &anyJSON); err == nil {
		return printOutput(b.cfg, anyJSON)
	}
	return printOutput(b.cfg, string(data))
}

func (b *commandBuilder) getFlagValue(cmd *cobra.Command, p ParamSpec) (flagValue, error) {
	f := cmd.Flags().Lookup(p.Flag)
	if f == nil {
		return flagValue{}, fmt.Errorf("flag not found: %s", p.Flag)
	}
	v := flagValue{changed: f.Changed}
	switch p.Kind {
	case ParamInt:
		n, err := cmd.Flags().GetInt(p.Flag)
		if err != nil {
			return v, err
		}
		v.intVal = n
	case ParamBool:
		bv, err := cmd.Flags().GetBool(p.Flag)
		if err != nil {
			return v, err
		}
		v.boolVal = bv
	case ParamSlice:
		sv, err := cmd.Flags().GetStringSlice(p.Flag)
		if err != nil {
			return v, err
		}
		v.sliceVal = sv
	default:
		sv, err := cmd.Flags().GetString(p.Flag)
		if err != nil {
			return v, err
		}
		v.stringVal = sv
		if p.Name == "output_file" && sv != "" {
			v.stringVal, _ = filepath.Abs(sv)
		}
	}
	return v, nil
}
