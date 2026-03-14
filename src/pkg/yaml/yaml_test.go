package yaml_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	pkgyaml "github.com/hoahm-ts/awesome-ai-skills/pkg/yaml"
)

type sampleConfig struct {
	Name  string `yaml:"name"`
	Value int    `yaml:"value"`
}

func TestParseFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		give      string // YAML file contents
		wantName  string
		wantValue int
		wantErr   bool
	}{
		{
			name:      "valid yaml file",
			give:      "name: hello\nvalue: 42\n",
			wantName:  "hello",
			wantValue: 42,
		},
		{
			name:    "unknown field rejected",
			give:    "name: hello\nunknown_key: oops\n",
			wantErr: true,
		},
		{
			name:    "invalid yaml syntax",
			give:    "name: :\n",
			wantErr: true,
		},
		{
			name:    "empty file returns error",
			give:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			path := filepath.Join(dir, "config.yml")
			require.NoError(t, os.WriteFile(path, []byte(tt.give), 0o600))

			got, err := pkgyaml.ParseFile[sampleConfig](path)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantName, got.Name)
			require.Equal(t, tt.wantValue, got.Value)
		})
	}
}

func TestParseFile_MissingFile(t *testing.T) {
	t.Parallel()

	_, err := pkgyaml.ParseFile[sampleConfig]("/nonexistent/path/config.yml")
	require.Error(t, err)
}
