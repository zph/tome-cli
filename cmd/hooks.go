package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	shellescape "al.essio.dev/pkg/shellescape"
	"github.com/gobeam/stringy"
)

type Hook struct {
	Path    string
	Name    string
	Sourced bool // true if filename ends with .source
}

type HookRunner struct {
	rootDir string
	config  *Config
}

func NewHookRunner(config *Config) *HookRunner {
	return &HookRunner{
		rootDir: config.RootDir(),
		config:  config,
	}
}

// DiscoverHooks finds all hooks in .hooks.d/
func (hr *HookRunner) DiscoverHooks() ([]Hook, error) {
	hooksDir := filepath.Join(hr.rootDir, ".hooks.d")

	// Check if hooks directory exists
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		log.Debugw(".hooks.d directory not found", "path", hooksDir)
		return []Hook{}, nil
	}

	entries, err := os.ReadDir(hooksDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read .hooks.d: %w", err)
	}

	var hooks []Hook
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		fullPath := filepath.Join(hooksDir, name)

		// Check if this is a sourced hook by file extension
		sourced := strings.HasSuffix(name, ".source")

		// If not a sourced hook, verify it's executable
		if !sourced {
			fileInfo, err := entry.Info()
			if err != nil {
				log.Warnw("failed to stat hook", "path", fullPath, "error", err)
				continue
			}

			if !isExecutableByOwner(fileInfo.Mode()) {
				log.Warnw("skipping non-executable hook without .source suffix", "path", fullPath)
				continue
			}
		}

		hook := Hook{
			Path:    fullPath,
			Name:    name,
			Sourced: sourced,
		}

		hooks = append(hooks, hook)
		log.Debugw("discovered hook", "path", fullPath, "sourced", hook.Sourced)
	}

	// Sort by name lexicographically (00- comes before 10-, etc.)
	sort.Slice(hooks, func(i, j int) bool {
		return hooks[i].Name < hooks[j].Name
	})

	log.Debugw("discovered hooks", "count", len(hooks))
	return hooks, nil
}

const wrapperScriptTemplate = `set -e
{{range .Env -}}
export {{.}}
{{end}}
{{range .Hooks -}}
# Hook: {{.Name}}
{{if .Sourced -}}
if ! source "{{.Path}}"; then
  echo 'Error: pre-hook failed: {{.Name}} (sourcing failed)' >&2
  exit 1
fi
{{else -}}
if ! "{{.Path}}"; then
  echo 'Error: pre-hook failed: {{.Name}}' >&2
  exit 1
fi
{{end}}
{{end -}}
# Execute target script
{{if .ScriptArgs -}}
exec "{{.ScriptPath}}" {{.ScriptArgs}}
{{else -}}
exec "{{.ScriptPath}}"
{{end -}}
`

type wrapperScriptData struct {
	Env        []string
	Hooks      []Hook
	ScriptPath string
	ScriptArgs string // Will be properly quoted when built
}

// GenerateWrapperScriptContent creates shell script content that sources/executes hooks and execs the target
func (hr *HookRunner) GenerateWrapperScriptContent(hooks []Hook, scriptPath string, scriptArgs []string) (string, error) {
	if len(hooks) == 0 {
		// No hooks, no wrapper needed
		return "", nil
	}

	// Prepare template data with properly quoted args using shellescape library
	quotedArgs := ""
	if len(scriptArgs) > 0 {
		quoted := make([]string, len(scriptArgs))
		for i, arg := range scriptArgs {
			quoted[i] = shellescape.Quote(arg)
		}
		quotedArgs = strings.Join(quoted, " ")
	}

	data := wrapperScriptData{
		Env:        hr.buildHookEnv("", scriptPath, scriptArgs),
		Hooks:      hooks,
		ScriptPath: scriptPath,
		ScriptArgs: quotedArgs,
	}

	// Parse and execute template
	tmpl, err := template.New("wrapper").Parse(wrapperScriptTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse wrapper template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute wrapper template: %w", err)
	}

	log.Debugw("generated wrapper script content")
	return buf.String(), nil
}

func (hr *HookRunner) buildHookEnv(hookPath, scriptPath string, scriptArgs []string) []string {
	var env []string

	// Add tome-cli standard vars
	absRootDir, _ := filepath.Abs(hr.config.RootDir())
	env = append(env, fmt.Sprintf("TOME_ROOT=%s", absRootDir))
	env = append(env, fmt.Sprintf("TOME_EXECUTABLE=%s", hr.config.ExecutableName()))

	// Add uppercase executable-specific vars
	executableAsEnvPrefix := strings.ToUpper(stringy.New(hr.config.ExecutableName()).SnakeCase().Get())
	env = append(env, fmt.Sprintf("%s_ROOT=%s", executableAsEnvPrefix, absRootDir))
	env = append(env, fmt.Sprintf("%s_EXECUTABLE=%s", executableAsEnvPrefix, hr.config.ExecutableName()))

	// Add script-specific vars
	env = append(env, fmt.Sprintf("TOME_SCRIPT_PATH=%s", scriptPath))
	env = append(env, fmt.Sprintf("TOME_SCRIPT_NAME=%s", filepath.Base(scriptPath)))
	env = append(env, fmt.Sprintf(`TOME_SCRIPT_ARGS="%s"`, strings.Join(scriptArgs, " ")))

	return env
}
