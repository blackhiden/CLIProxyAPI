package config

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestConfigExampleDocumentsRequestTimeout(t *testing.T) {
	data, err := os.ReadFile("../../config.example.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var doc yaml.Node
	if err = yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("config.example.yaml is not valid YAML: %v", err)
	}
	root := doc.Content[0]
	for _, path := range []string{
		"requests.request-timeout",
		"requests.hard-request-deadline",
	} {
		node := yamlPath(root, path)
		if node == nil {
			t.Fatalf("config.example.yaml is missing %s", path)
		}
	}
}
