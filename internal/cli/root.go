package cli

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func NewRootCmd() *cobra.Command {
	cfg := &Config{}
	root := &cobra.Command{
		Use:   "librarease",
		Short: "LibrarEase CLI over HTTPS",
		Long: `LibrarEase CLI

Auth & configuration:
  Use --token for Bearer auth, or --client-id and --uid for internal auth mode.
  Config precedence: flags > env vars > config file.
  Env keys: LIBRAREASE_BASE_URL, LIBRAREASE_TOKEN, LIBRAREASE_CLIENT_ID,
  LIBRAREASE_UID, LIBRAREASE_TIMEOUT, LIBRAREASE_OUTPUT.
`,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return bindConfig(cmd.Root(), cfg)
		},
	}

	root.PersistentFlags().String("config", "", "Path to config file")
	root.PersistentFlags().String("base-url", "https://localhost:8080", "Base API URL")
	root.PersistentFlags().String("token", "", "Bearer token")
	root.PersistentFlags().String("client-id", "", "Internal client ID (X-Client-Id)")
	root.PersistentFlags().String("uid", "", "Internal UID (X-Uid)")
	root.PersistentFlags().String("timeout", "30s", "HTTP timeout")
	root.PersistentFlags().String("output", "json", "Output: json|yaml|table|raw")
	root.PersistentFlags().Bool("insecure-skip-verify", false, "Skip TLS certificate verify")
	root.PersistentFlags().Bool("quiet", false, "Suppress non-data success messages")

	_ = root.PersistentFlags().Lookup("config")
	root.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		cf, _ := cmd.Flags().GetString("config")
		cfg.ConfigFile = cf
		return bindConfig(cmd.Root(), cfg)
	}

	api := NewAPIClient(cfg)
	specs := endpointSpecs()

	groupNames := make([]string, 0, len(specs))
	for n := range specs {
		if n == "users_push_token" || n == "users_watchlist" {
			continue
		}
		groupNames = append(groupNames, n)
	}
	sort.Strings(groupNames)

	groupCmds := map[string]*cobra.Command{}
	for _, gn := range groupNames {
		gc := &cobra.Command{Use: gn, Short: fmt.Sprintf("%s commands", gn)}
		for _, sp := range specs[gn] {
			gc.AddCommand(newCommandFromSpec(sp, cfg, api))
		}
		groupCmds[gn] = gc
		root.AddCommand(gc)
	}

	users := groupCmds["users"]
	if users != nil {
		pushToken := &cobra.Command{Use: "push-token", Short: "Push token commands"}
		for _, sp := range specs["users_push_token"] {
			pushToken.AddCommand(newCommandFromSpec(sp, cfg, api))
		}
		users.AddCommand(pushToken)

		watchlist := &cobra.Command{Use: "watchlist", Short: "Watchlist commands"}
		for _, sp := range specs["users_watchlist"] {
			watchlist.AddCommand(newCommandFromSpec(sp, cfg, api))
		}
		users.AddCommand(watchlist)
	}

	notifications := groupCmds["notifications"]
	if notifications != nil {
		notifications.AddCommand(newNotificationStreamCmd(cfg))
	}

	root.AddCommand(newDocsCmd(root))
	return root
}

func newNotificationStreamCmd(cfg *Config) *cobra.Command {
	var userID string
	cmd := &cobra.Command{
		Use:   "stream",
		Short: "Stream notifications (SSE)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if userID == "" {
				return fmt.Errorf("--user-id is required")
			}

			u := strings.TrimRight(cfg.BaseURL, "/") + "/api/v1/notifications/stream?user_id=" + userID
			req, err := http.NewRequest(http.MethodGet, u, nil)
			if err != nil {
				return err
			}
			if cfg.Token != "" {
				req.Header.Set("Authorization", "Bearer "+cfg.Token)
			}
			if cfg.ClientID != "" {
				req.Header.Set("X-Client-Id", cfg.ClientID)
			}
			if cfg.UID != "" {
				req.Header.Set("X-Uid", cfg.UID)
			}
			client := NewAPIClient(cfg).http
			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			if resp.StatusCode >= 400 {
				b, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("http %d: %s", resp.StatusCode, string(b))
			}
			sc := bufio.NewScanner(resp.Body)
			for sc.Scan() {
				line := strings.TrimSpace(sc.Text())
				if strings.HasPrefix(line, "data:") {
					fmt.Fprintln(cmd.OutOrStdout(), strings.TrimSpace(strings.TrimPrefix(line, "data:")))
				}
			}
			return sc.Err()
		},
	}
	cmd.Flags().StringVar(&userID, "user-id", "", "User ID")
	_ = cmd.MarkFlagRequired("user-id")
	return cmd
}

func newDocsCmd(root *cobra.Command) *cobra.Command {
	var outDir string
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Generate markdown command docs",
		RunE: func(_ *cobra.Command, _ []string) error {
			if outDir == "" {
				outDir = "./docs/cli"
			}
			if err := os.MkdirAll(outDir, 0o755); err != nil {
				return err
			}
			prepender := func(_ string) string { return "" }
			linkHandler := func(s string) string { return s }
			return doc.GenMarkdownTreeCustom(root, outDir, prepender, linkHandler)
		},
	}
	cmd.Flags().StringVar(&outDir, "out", "./docs/cli", "Output directory")
	return cmd
}
