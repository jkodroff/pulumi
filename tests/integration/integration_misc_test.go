// Copyright 2016-2018, Pulumi Corporation.  All rights reserved.
//go:build misc || all
// +build misc all

package ints

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"

	"github.com/stretchr/testify/assert"

	"github.com/pulumi/pulumi/pkg/v3/resource/deploy/providers"
	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/config"
	ptesting "github.com/pulumi/pulumi/sdk/v3/go/common/testing"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
)

// TestStackTagValidation verifies various error scenarios related to stack names and tags.
func TestStackTagValidation(t *testing.T) {
	t.Run("Error_StackName", func(t *testing.T) {
		e := ptesting.NewEnvironment(t)
		defer func() {
			if !t.Failed() {
				e.DeleteEnvironment()
			}
		}()
		e.RunCommand("git", "init")

		e.ImportDirectory("stack_project_name")
		e.RunCommand("pulumi", "login", "--cloud-url", e.LocalURL())

		stdout, stderr := e.RunCommandExpectError("pulumi", "stack", "init", "invalid name (spaces, parens, etc.)")
		assert.Equal(t, "", stdout)
		assert.Contains(t, stderr, "stack names may only contain alphanumeric, hyphens, underscores, or periods")
	})

	t.Run("Error_DescriptionLength", func(t *testing.T) {
		e := ptesting.NewEnvironment(t)
		defer func() {
			if !t.Failed() {
				e.DeleteEnvironment()
			}
		}()
		e.RunCommand("git", "init")

		e.ImportDirectory("stack_project_name")
		e.RunCommand("pulumi", "login", "--cloud-url", e.LocalURL())

		prefix := "lorem ipsum dolor sit amet"     // 26
		prefix = prefix + prefix + prefix + prefix // 104
		prefix = prefix + prefix + prefix + prefix // 416 + the current Pulumi.yaml's description

		// Change the contents of the Description property of Pulumi.yaml.
		yamlPath := filepath.Join(e.CWD, "Pulumi.yaml")
		err := integration.ReplaceInFile("description: ", "description: "+prefix, yamlPath)
		assert.NoError(t, err)

		stdout, stderr := e.RunCommandExpectError("pulumi", "stack", "init", "valid-name")
		assert.Equal(t, "", stdout)
		assert.Contains(t, stderr, "error: could not create stack:")
		assert.Contains(t, stderr, "validating stack properties:")
		assert.Contains(t, stderr, "stack tag \"pulumi:description\" value is too long (max length 256 characters)")
	})
}

// TestStackInitValidation verifies various error scenarios related to init'ing a stack.
func TestStackInitValidation(t *testing.T) {
	t.Run("Error_InvalidStackYaml", func(t *testing.T) {
		e := ptesting.NewEnvironment(t)
		defer func() {
			if !t.Failed() {
				e.DeleteEnvironment()
			}
		}()
		e.RunCommand("git", "init")

		e.ImportDirectory("stack_project_name")
		e.RunCommand("pulumi", "login", "--cloud-url", e.LocalURL())

		// Starting a yaml value with a quote string and then more data is invalid
		invalidYaml := "\"this is invalid\" yaml because of trailing data after quote string"

		// Change the contents of the Description property of Pulumi.yaml.
		yamlPath := filepath.Join(e.CWD, "Pulumi.yaml")
		err := integration.ReplaceInFile("description: ", "description: "+invalidYaml, yamlPath)
		assert.NoError(t, err)

		stdout, stderr := e.RunCommandExpectError("pulumi", "stack", "init", "valid-name")
		assert.Equal(t, "", stdout)
		assert.Contains(t, stderr,
			"error: could not get cloud url: could not load current project: "+
				"invalid YAML file: yaml: line 1: did not find expected key")
	})
}

// TestConfigSave ensures that config commands in the Pulumi CLI work as expected.
func TestConfigSave(t *testing.T) {
	e := ptesting.NewEnvironment(t)
	defer func() {
		if !t.Failed() {
			e.DeleteEnvironment()
		}
	}()

	// Initialize an empty stack.
	path := filepath.Join(e.RootPath, "Pulumi.yaml")
	err := (&workspace.Project{
		Name:    "testing-config",
		Runtime: workspace.NewProjectRuntimeInfo("nodejs", nil),
	}).Save(path)
	assert.NoError(t, err)
	e.RunCommand("pulumi", "login", "--cloud-url", e.LocalURL())
	e.RunCommand("pulumi", "stack", "init", "testing-2")
	e.RunCommand("pulumi", "stack", "init", "testing-1")

	// Now configure and save a few different things:
	e.RunCommand("pulumi", "config", "set", "configA", "value1")
	e.RunCommand("pulumi", "config", "set", "configB", "value2", "--stack", "testing-2")

	e.RunCommand("pulumi", "stack", "select", "testing-2")

	e.RunCommand("pulumi", "config", "set", "configD", "value4")
	e.RunCommand("pulumi", "config", "set", "configC", "value3", "--stack", "testing-1")

	// Now read back the config using the CLI:
	{
		stdout, _ := e.RunCommand("pulumi", "config", "get", "configB")
		assert.Equal(t, "value2\n", stdout)
	}
	{
		// the config in a different stack, so this should error.
		stdout, stderr := e.RunCommandExpectError("pulumi", "config", "get", "configA")
		assert.Equal(t, "", stdout)
		assert.NotEqual(t, "", stderr)
	}
	{
		// but selecting the stack should let you see it
		stdout, _ := e.RunCommand("pulumi", "config", "get", "configA", "--stack", "testing-1")
		assert.Equal(t, "value1\n", stdout)
	}

	// Finally, check that the stack file contains what we expected.
	validate := func(k string, v string, cfg config.Map) {
		key, err := config.ParseKey("testing-config:config:" + k)
		assert.NoError(t, err)
		d, ok := cfg[key]
		assert.True(t, ok, "config key %v should be set", k)
		dv, err := d.Value(nil)
		assert.NoError(t, err)
		assert.Equal(t, v, dv)
	}

	testStack1, err := workspace.LoadProjectStack(filepath.Join(e.CWD, "Pulumi.testing-1.yaml"))
	assert.NoError(t, err)
	testStack2, err := workspace.LoadProjectStack(filepath.Join(e.CWD, "Pulumi.testing-2.yaml"))
	assert.NoError(t, err)

	assert.Equal(t, 2, len(testStack1.Config))
	assert.Equal(t, 2, len(testStack2.Config))

	validate("configA", "value1", testStack1.Config)
	validate("configC", "value3", testStack1.Config)

	validate("configB", "value2", testStack2.Config)
	validate("configD", "value4", testStack2.Config)

	e.RunCommand("pulumi", "stack", "rm", "--yes")
}

// TestConfigPaths ensures that config commands with paths work as expected.
func TestConfigPaths(t *testing.T) {
	e := ptesting.NewEnvironment(t)
	defer func() {
		if !t.Failed() {
			e.DeleteEnvironment()
		}
	}()

	// Initialize an empty stack.
	path := filepath.Join(e.RootPath, "Pulumi.yaml")
	err := (&workspace.Project{
		Name:    "testing-config",
		Runtime: workspace.NewProjectRuntimeInfo("nodejs", nil),
	}).Save(path)
	assert.NoError(t, err)
	e.RunCommand("pulumi", "login", "--cloud-url", e.LocalURL())
	e.RunCommand("pulumi", "stack", "init", "testing")

	namespaces := []string{"", "my:"}

	tests := []struct {
		Key                   string
		Value                 string
		Secret                bool
		Path                  bool
		TopLevelKey           string
		TopLevelExpectedValue string
	}{
		{
			Key:                   "aConfigValue",
			Value:                 "this value is a value",
			TopLevelKey:           "aConfigValue",
			TopLevelExpectedValue: "this value is a value",
		},
		{
			Key:                   "anotherConfigValue",
			Value:                 "this value is another value",
			TopLevelKey:           "anotherConfigValue",
			TopLevelExpectedValue: "this value is another value",
		},
		{
			Key:                   "bEncryptedSecret",
			Value:                 "this super secret is encrypted",
			Secret:                true,
			TopLevelKey:           "bEncryptedSecret",
			TopLevelExpectedValue: "this super secret is encrypted",
		},
		{
			Key:                   "anotherEncryptedSecret",
			Value:                 "another encrypted secret",
			Secret:                true,
			TopLevelKey:           "anotherEncryptedSecret",
			TopLevelExpectedValue: "another encrypted secret",
		},
		{
			Key:                   "[]",
			Value:                 "square brackets value",
			TopLevelKey:           "[]",
			TopLevelExpectedValue: "square brackets value",
		},
		{
			Key:                   "x.y",
			Value:                 "x.y value",
			TopLevelKey:           "x.y",
			TopLevelExpectedValue: "x.y value",
		},
		{
			Key:                   "0",
			Value:                 "0 value",
			Path:                  true,
			TopLevelKey:           "0",
			TopLevelExpectedValue: "0 value",
		},
		{
			Key:                   "true",
			Value:                 "value",
			Path:                  true,
			TopLevelKey:           "true",
			TopLevelExpectedValue: "value",
		},
		{
			Key:                   `["test.Key"]`,
			Value:                 "test key value",
			Path:                  true,
			TopLevelKey:           "test.Key",
			TopLevelExpectedValue: "test key value",
		},
		{
			Key:                   `nested["test.Key"]`,
			Value:                 "nested test key value",
			Path:                  true,
			TopLevelKey:           "nested",
			TopLevelExpectedValue: `{"test.Key":"nested test key value"}`,
		},
		{
			Key:                   "outer.inner",
			Value:                 "value",
			Path:                  true,
			TopLevelKey:           "outer",
			TopLevelExpectedValue: `{"inner":"value"}`,
		},
		{
			Key:                   "names[0]",
			Value:                 "a",
			Path:                  true,
			TopLevelKey:           "names",
			TopLevelExpectedValue: `["a"]`,
		},
		{
			Key:                   "names[1]",
			Value:                 "b",
			Path:                  true,
			TopLevelKey:           "names",
			TopLevelExpectedValue: `["a","b"]`,
		},
		{
			Key:                   "names[2]",
			Value:                 "c",
			Path:                  true,
			TopLevelKey:           "names",
			TopLevelExpectedValue: `["a","b","c"]`,
		},
		{
			Key:                   "names[3]",
			Value:                 "super secret name",
			Path:                  true,
			Secret:                true,
			TopLevelKey:           "names",
			TopLevelExpectedValue: `["a","b","c","super secret name"]`,
		},
		{
			Key:                   "servers[0].port",
			Value:                 "80",
			Path:                  true,
			TopLevelKey:           "servers",
			TopLevelExpectedValue: `[{"port":80}]`,
		},
		{
			Key:                   "servers[0].host",
			Value:                 "example",
			Path:                  true,
			TopLevelKey:           "servers",
			TopLevelExpectedValue: `[{"host":"example","port":80}]`,
		},
		{
			Key:                   "a.b[0].c",
			Value:                 "true",
			Path:                  true,
			TopLevelKey:           "a",
			TopLevelExpectedValue: `{"b":[{"c":true}]}`,
		},
		{
			Key:                   "a.b[1].c",
			Value:                 "false",
			Path:                  true,
			TopLevelKey:           "a",
			TopLevelExpectedValue: `{"b":[{"c":true},{"c":false}]}`,
		},
		{
			Key:                   "tokens[0]",
			Value:                 "shh",
			Path:                  true,
			Secret:                true,
			TopLevelKey:           "tokens",
			TopLevelExpectedValue: `["shh"]`,
		},
		{
			Key:                   "foo.bar",
			Value:                 "don't tell",
			Path:                  true,
			Secret:                true,
			TopLevelKey:           "foo",
			TopLevelExpectedValue: `{"bar":"don't tell"}`,
		},
		{
			Key:                   "semiInner.a.b.c.d",
			Value:                 "1",
			Path:                  true,
			TopLevelKey:           "semiInner",
			TopLevelExpectedValue: `{"a":{"b":{"c":{"d":1}}}}`,
		},
		{
			Key:                   "wayInner.a.b.c.d.e.f.g.h.i.j.k",
			Value:                 "false",
			Path:                  true,
			TopLevelKey:           "wayInner",
			TopLevelExpectedValue: `{"a":{"b":{"c":{"d":{"e":{"f":{"g":{"h":{"i":{"j":{"k":false}}}}}}}}}}}`,
		},
		{
			Key:                   "foo1[0]",
			Value:                 "false",
			Path:                  true,
			TopLevelKey:           "foo1",
			TopLevelExpectedValue: `[false]`,
		},
		{
			Key:                   "foo2[0]",
			Value:                 "true",
			Path:                  true,
			TopLevelKey:           "foo2",
			TopLevelExpectedValue: `[true]`,
		},
		{
			Key:                   "foo3[0]",
			Value:                 "10",
			Path:                  true,
			TopLevelKey:           "foo3",
			TopLevelExpectedValue: `[10]`,
		},
		{
			Key:                   "foo4[0]",
			Value:                 "0",
			Path:                  true,
			TopLevelKey:           "foo4",
			TopLevelExpectedValue: `[0]`,
		},
		{
			Key:                   "foo5[0]",
			Value:                 "00",
			Path:                  true,
			TopLevelKey:           "foo5",
			TopLevelExpectedValue: `["00"]`,
		},
		{
			Key:                   "foo6[0]",
			Value:                 "01",
			Path:                  true,
			TopLevelKey:           "foo6",
			TopLevelExpectedValue: `["01"]`,
		},
		{
			Key:                   "foo7[0]",
			Value:                 "0123456",
			Path:                  true,
			TopLevelKey:           "foo7",
			TopLevelExpectedValue: `["0123456"]`,
		},
		{
			Key:                   "bar1.inner",
			Value:                 "false",
			Path:                  true,
			TopLevelKey:           "bar1",
			TopLevelExpectedValue: `{"inner":false}`,
		},
		{
			Key:                   "bar2.inner",
			Value:                 "true",
			Path:                  true,
			TopLevelKey:           "bar2",
			TopLevelExpectedValue: `{"inner":true}`,
		},
		{
			Key:                   "bar3.inner",
			Value:                 "10",
			Path:                  true,
			TopLevelKey:           "bar3",
			TopLevelExpectedValue: `{"inner":10}`,
		},
		{
			Key:                   "bar4.inner",
			Value:                 "0",
			Path:                  true,
			TopLevelKey:           "bar4",
			TopLevelExpectedValue: `{"inner":0}`,
		},
		{
			Key:                   "bar5.inner",
			Value:                 "00",
			Path:                  true,
			TopLevelKey:           "bar5",
			TopLevelExpectedValue: `{"inner":"00"}`,
		},
		{
			Key:                   "bar6.inner",
			Value:                 "01",
			Path:                  true,
			TopLevelKey:           "bar6",
			TopLevelExpectedValue: `{"inner":"01"}`,
		},
		{
			Key:                   "bar7.inner",
			Value:                 "0123456",
			Path:                  true,
			TopLevelKey:           "bar7",
			TopLevelExpectedValue: `{"inner":"0123456"}`,
		},

		// Overwriting a top-level string value is allowed.
		{
			Key:                   "aConfigValue.inner",
			Value:                 "new value",
			Path:                  true,
			TopLevelKey:           "aConfigValue",
			TopLevelExpectedValue: `{"inner":"new value"}`,
		},
		{
			Key:                   "anotherConfigValue[0]",
			Value:                 "new value",
			Path:                  true,
			TopLevelKey:           "anotherConfigValue",
			TopLevelExpectedValue: `["new value"]`,
		},
		{
			Key:                   "bEncryptedSecret.inner",
			Value:                 "new value",
			Path:                  true,
			TopLevelKey:           "bEncryptedSecret",
			TopLevelExpectedValue: `{"inner":"new value"}`,
		},
		{
			Key:                   "anotherEncryptedSecret[0]",
			Value:                 "new value",
			Path:                  true,
			TopLevelKey:           "anotherEncryptedSecret",
			TopLevelExpectedValue: `["new value"]`,
		},
	}

	validateConfigGet := func(key string, value string, path bool) {
		args := []string{"config", "get", key}
		if path {
			args = append(args, "--path")
		}
		stdout, stderr := e.RunCommand("pulumi", args...)
		assert.Equal(t, fmt.Sprintf("%s\n", value), stdout)
		assert.Equal(t, "", stderr)
	}

	for _, ns := range namespaces {
		for _, test := range tests {
			key := fmt.Sprintf("%s%s", ns, test.Key)
			topLevelKey := fmt.Sprintf("%s%s", ns, test.TopLevelKey)

			// Set the value.
			args := []string{"config", "set"}
			if test.Secret {
				args = append(args, "--secret")
			}
			if test.Path {
				args = append(args, "--path")
			}
			args = append(args, key, test.Value)
			stdout, stderr := e.RunCommand("pulumi", args...)
			assert.Equal(t, "", stdout)
			assert.Equal(t, "", stderr)

			// Get the value and validate it.
			validateConfigGet(key, test.Value, test.Path)

			// Get the top-level value and validate it.
			validateConfigGet(topLevelKey, test.TopLevelExpectedValue, false /*path*/)
		}
	}

	badKeys := []string{
		// Syntax errors.
		"root[",
		`root["nested]`,
		"root.array[abc]",
		"root.[1]",

		// First path segment must be a non-empty string.
		`[""]`,
		"[0]",

		// Index out of range.
		"names[-1]",
		"names[5]",

		// A "secure" key that is a map with a single string value is reserved by the system.
		"key.secure",
		"super.nested.map.secure",

		// Type mismatch.
		"outer[0]",
		"names.nested",
		"outer.inner.nested",
		"outer.inner[0]",
	}

	for _, ns := range namespaces {
		for _, badKey := range badKeys {
			key := fmt.Sprintf("%s%s", ns, badKey)
			stdout, stderr := e.RunCommandExpectError("pulumi", "config", "set", "--path", key, "value")
			assert.Equal(t, "", stdout)
			assert.NotEqual(t, "", stderr)
		}
	}

	e.RunCommand("pulumi", "stack", "rm", "--yes")
}

func TestRotatePassphrase(t *testing.T) {
	e := ptesting.NewEnvironment(t)
	defer func() {
		if !t.Failed() {
			e.DeleteEnvironment()
		}
	}()

	e.ImportDirectory("rotate_passphrase")
	e.RunCommand("pulumi", "login", "--cloud-url", e.LocalURL())

	e.RunCommand("pulumi", "stack", "init", "dev")
	e.RunCommand("pulumi", "up", "--skip-preview", "--yes")

	e.RunCommand("pulumi", "config", "set", "--secret", "foo", "bar")

	e.SetEnvVars([]string{"PULUMI_TEST_PASSPHRASE=true"})
	e.Stdin = strings.NewReader("qwerty\nqwerty\n")
	e.RunCommand("pulumi", "stack", "change-secrets-provider", "passphrase")

	e.Stdin, e.Passphrase = nil, "qwerty"
	e.RunCommand("pulumi", "config", "get", "foo")
}

func TestJSONOutput(t *testing.T) {
	stdout := &bytes.Buffer{}

	// Test without env var for streaming preview (should print previewSummary).
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir:          filepath.Join("stack_outputs", "nodejs"),
		Dependencies: []string{"@pulumi/pulumi"},
		Stdout:       stdout,
		Verbose:      true,
		JSONOutput:   true,
		ExtraRuntimeValidation: func(t *testing.T, stack integration.RuntimeValidationStackInfo) {
			output := stdout.String()

			// Check that the previewSummary is present.
			assert.Regexp(t, previewSummaryRegex, output)

			// Check that each event present in the event stream is also in stdout.
			for _, evt := range stack.Events {
				assertOutputContainsEvent(t, evt, output)
			}
		},
	})
}

func TestJSONOutputWithStreamingPreview(t *testing.T) {
	stdout := &bytes.Buffer{}

	// Test with env var for streaming preview (should *not* print previewSummary).
	integration.ProgramTest(t, &integration.ProgramTestOptions{
		Dir:          filepath.Join("stack_outputs", "nodejs"),
		Dependencies: []string{"@pulumi/pulumi"},
		Stdout:       stdout,
		Verbose:      true,
		JSONOutput:   true,
		Env:          []string{"PULUMI_ENABLE_STREAMING_JSON_PREVIEW=1"},
		ExtraRuntimeValidation: func(t *testing.T, stack integration.RuntimeValidationStackInfo) {
			output := stdout.String()

			// Check that the previewSummary is *not* present.
			assert.NotRegexp(t, previewSummaryRegex, output)

			// Check that each event present in the event stream is also in stdout.
			for _, evt := range stack.Events {
				assertOutputContainsEvent(t, evt, output)
			}
		},
	})
}

func TestExcludeProtected(t *testing.T) {
	e := ptesting.NewEnvironment(t)
	defer func() {
		if !t.Failed() {
			e.DeleteEnvironment()
		}
	}()

	e.ImportDirectory("exclude_protected")

	e.RunCommand("pulumi", "login", "--cloud-url", e.LocalURL())

	e.RunCommand("pulumi", "stack", "init", "dev")

	e.RunCommand("yarn", "link", "@pulumi/pulumi")
	e.RunCommand("yarn", "install")

	e.RunCommand("pulumi", "up", "--skip-preview", "--yes")

	stdout, _ := e.RunCommand("pulumi", "destroy", "--skip-preview", "--yes", "--exclude-protected")
	assert.Contains(t, stdout, "All unprotected resources were destroyed. There are still 7 protected resources")
	// We run the command again, but this time there are not unprotected resources to destroy.
	stdout, _ = e.RunCommand("pulumi", "destroy", "--skip-preview", "--yes", "--exclude-protected")
	assert.Contains(t, stdout, "There were no unprotected resources to destroy. There are still 7")
}

func TestProviderDownloadURL(t *testing.T) {
	validate := func(t *testing.T, stdout []byte) {
		deployment := &apitype.UntypedDeployment{}
		err := json.Unmarshal(stdout, deployment)
		assert.NoError(t, err)
		data := &apitype.DeploymentV3{}
		err = json.Unmarshal(deployment.Deployment, data)
		assert.NoError(t, err)
		urlKey := "pluginDownloadURL"
		for _, resource := range data.Resources {
			switch {
			case providers.IsDefaultProvider(resource.URN):
				assert.Equalf(t, "get.com", resource.Inputs[urlKey], "Inputs")
				assert.Equalf(t, "get.com", resource.Outputs[urlKey], "Outputs")
			case providers.IsProviderType(resource.Type):
				assert.Equalf(t, "get.pulumi/test/providers", resource.Inputs[urlKey], "Inputs")
				assert.Equal(t, "get.pulumi/test/providers", resource.Outputs[urlKey], "Outputs")
			default:
				_, hasURL := resource.Inputs[urlKey]
				assert.False(t, hasURL)
				_, hasURL = resource.Outputs[urlKey]
				assert.False(t, hasURL)
			}
		}
		assert.Greater(t, len(data.Resources), 1, "We should construct more then just the stack")
	}

	languages := []struct {
		name       string
		dependency string
	}{

		{"python", filepath.Join("..", "..", "sdk", "python", "env", "src")},
		{"nodejs", "@pulumi/pulumi"},
		{"dotnet", "Pulumi"},
		{"go", "github.com/pulumi/pulumi/sdk/v3"},
	}

	for _, lang := range languages {
		t.Run(lang.name, func(t *testing.T) {
			env := pathEnv(t, filepath.Join("..", "testprovider"))
			dir := filepath.Join("gather_plugin", lang.name)
			integration.ProgramTest(t, &integration.ProgramTestOptions{
				Dir:                    dir,
				Env:                    []string{env},
				ExportStateValidator:   validate,
				SkipPreview:            true,
				SkipEmptyPreviewUpdate: true,
				Dependencies:           []string{lang.dependency},
			})
		})
	}
}