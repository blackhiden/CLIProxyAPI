package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestParseConfigBytesSanitizesAPIKeyNames(t *testing.T) {
	cfg, errParse := ParseConfigBytes([]byte(`
api-keys:
  - client-one
  - client-two
api-key-names:
  client-one: "  Shell  "
  client-two: ""
  stale-key: Stale
`))
	if errParse != nil {
		t.Fatal(errParse)
	}
	if len(cfg.APIKeys) != 2 || cfg.APIKeys[0] != "client-one" || cfg.APIKeys[1] != "client-two" {
		t.Fatalf("APIKeys = %#v, want configured client keys", cfg.APIKeys)
	}
	if len(cfg.APIKeyNames) != 1 || cfg.APIKeyNames["client-one"] != "Shell" {
		t.Fatalf("APIKeyNames = %#v, want only trimmed configured names", cfg.APIKeyNames)
	}
}

func TestNormalizeV8ConfigPrunesStaleAPIKeyNames(t *testing.T) {
	data, _, errNormalize := NormalizeConfigLayout([]byte(`
access:
  api-keys: [client-one]
  api-key-names:
    client-one: "  Shell  "
    stale-key: Stale
    blank-name: " "
`), true)
	if errNormalize != nil {
		t.Fatal(errNormalize)
	}
	var doc yaml.Node
	if errUnmarshal := yaml.Unmarshal(data, &doc); errUnmarshal != nil {
		t.Fatal(errUnmarshal)
	}
	names := yamlPath(doc.Content[0], "access.api-key-names")
	if names == nil {
		t.Fatal("normalized config lost configured API-key name")
	}
	var got map[string]string
	if errDecode := names.Decode(&got); errDecode != nil {
		t.Fatal(errDecode)
	}
	if len(got) != 1 || got["client-one"] != "Shell" {
		t.Fatalf("normalized names = %#v, want only trimmed configured name", got)
	}
}

func TestAPIKeyNamesPersistInLegacyAndV8Layouts(t *testing.T) {
	for _, tc := range []struct {
		name, initial, namesPath string
		migrate                  bool
	}{
		{
			name: "legacy",
			initial: `api-keys:
  - client-one
  - client-two
api-key-names:
  stale-key: Stale
`,
			namesPath: "api-key-names",
		},
		{
			name: "v8",
			initial: `config-version: 8
access:
  api-keys:
    - client-one
    - client-two
  api-key-names:
    stale-key: Stale
`,
			namesPath: "access.api-key-names",
			migrate:   true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			if errWrite := os.WriteFile(path, []byte(tc.initial), 0600); errWrite != nil {
				t.Fatal(errWrite)
			}

			cfg, errLoad := LoadConfig(path)
			if errLoad != nil {
				t.Fatal(errLoad)
			}
			cfg.APIKeyNames = map[string]string{
				"client-one": "Shell",
				"client-two": "VSCode",
				"stale-key":  "Stale",
			}
			if errSave := SaveConfigPreserveComments(path, cfg, tc.migrate); errSave != nil {
				t.Fatal(errSave)
			}

			data, errRead := os.ReadFile(path)
			if errRead != nil {
				t.Fatal(errRead)
			}
			if tc.migrate {
				if errValidate := ValidateV8Config(data); errValidate != nil {
					t.Fatalf("ValidateV8Config() error = %v", errValidate)
				}
			}
			var doc yaml.Node
			if errUnmarshal := yaml.Unmarshal(data, &doc); errUnmarshal != nil {
				t.Fatal(errUnmarshal)
			}
			names := yamlPath(doc.Content[0], tc.namesPath)
			if names == nil {
				t.Fatalf("saved config missing %s", tc.namesPath)
			}
			var got map[string]string
			if errDecode := names.Decode(&got); errDecode != nil {
				t.Fatal(errDecode)
			}
			want := map[string]string{"client-one": "Shell", "client-two": "VSCode"}
			if len(got) != len(want) || got["client-one"] != want["client-one"] || got["client-two"] != want["client-two"] {
				t.Fatalf("saved names = %#v, want %#v", got, want)
			}

			reloaded, errReload := LoadConfig(path)
			if errReload != nil {
				t.Fatal(errReload)
			}
			if len(reloaded.APIKeyNames) != len(want) || reloaded.APIKeyNames["client-one"] != "Shell" || reloaded.APIKeyNames["client-two"] != "VSCode" {
				t.Fatalf("reloaded names = %#v, want %#v", reloaded.APIKeyNames, want)
			}
		})
	}
}
