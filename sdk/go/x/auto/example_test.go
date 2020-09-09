// Copyright 2016-2020, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//nolint
package auto

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v2/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v2/go/common/workspace"
	"github.com/pulumi/pulumi/sdk/v2/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v2/go/x/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v2/go/x/auto/optpreview"
	"github.com/pulumi/pulumi/sdk/v2/go/x/auto/optrefresh"
	"github.com/pulumi/pulumi/sdk/v2/go/x/auto/optup"
)

func Example() error {
	ctx := context.Background()

	// This stack creates an output
	fqsn := FullyQualifiedStackName("myOrg", "projA", "devStack")
	stackA, err := NewStackInlineSource(ctx, fqsn, func(pCtx *pulumi.Context) error {
		pCtx.Export("outputA", pulumi.String("valueA"))
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to create stackA")
	}
	// deploy the stack
	aRes, err := stackA.Up(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to update bucket stackA")
	}

	// this stack creates an uses stackA's output to create a new output
	fqsn = FullyQualifiedStackName("myOrg", "projB", "devStack")
	stackB, err := NewStackInlineSource(ctx, fqsn, func(pCtx *pulumi.Context) error {
		// output a new value "valueA/valueB"
		pCtx.Export(
			"outputB",
			pulumi.Sprintf(
				"%s/%s",
				pulumi.String(aRes.Outputs["outputA"].Value.(string)), pulumi.String("valueB"),
			),
		)
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to create object stackB")
	}
	// deploy the stack
	bRes, err := stackB.Up(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to update stackB")
	}

	// Success!
	fmt.Println(bRes.Summary.Result)
	return nil
}

func ExampleFullyQualifiedStackName() {
	fqsn := FullyQualifiedStackName("myOrgName", "myProjectName", "myStackName")
	// "myOrgName/myProjectName/myStackName"
	ctx := context.Background()
	_, _ = NewStackLocalSource(ctx, fqsn, filepath.Join(".", "project"))
}

func ExampleIsCompilationError() {
	ctx := context.Background()
	s, _ := NewStackLocalSource(ctx, "o/p/s", filepath.Join(".", "broken", "go", "program"))
	_, err := s.Up(ctx)
	if err != nil && IsCompilationError(err) {
		// time to fix up the program
	}
}

func ExampleIsConcurrentUpdateError() {
	ctx := context.Background()
	s, _ := NewStackLocalSource(ctx, "o/p/s", filepath.Join(".", "update", "inprogress", "program"))
	_, err := s.Up(ctx)
	if err != nil && IsConcurrentUpdateError(err) {
		// stack "o/p/s" already has an update in progress
		// implement some retry logic here...
	}
}

func ExampleIsRuntimeError() {
	ctx := context.Background()
	s, _ := NewStackLocalSource(ctx, "o/p/s", filepath.Join(".", "runtime", "error", "program"))
	_, err := s.Up(ctx)
	if err != nil && IsRuntimeError(err) {
		// oops, maybe there's an NPE in the program?
	}
}

func ExampleIsUnexpectedEngineError() {
	ctx := context.Background()
	s, _ := NewStackLocalSource(ctx, "o/p/s", filepath.Join(".", "engine", "error", "program"))
	_, err := s.Up(ctx)
	if err != nil && IsUnexpectedEngineError(err) {
		// oops, you've probably uncovered a bug in the pulumi engine.
		// please file a bug report :)
	}
}

func ExampleValidateFullyQualifiedStackName() {
	badFqsn := "project/stack"
	// false
	fmt.Println(ValidateFullyQualifiedStackName(badFqsn))
}

func ExampleConfigMap() {
	cfg := ConfigMap{
		"plaintext": {Value: "unencrypted"},
		"secret":    {Value: "encrypted", Secret: true},
	}
	ctx := context.Background()
	s, _ := NewStackLocalSource(ctx, "o/p/s", filepath.Join(".", "project"))
	s.SetAllConfig(ctx, cfg)
}

func ExampleConfigValue() {
	cfgVal := ConfigValue{
		Value:  "an secret that will be marked as and stored encrypted",
		Secret: true,
	}
	ctx := context.Background()
	s, _ := NewStackLocalSource(ctx, "o/p/s", filepath.Join(".", "project"))
	s.SetConfig(ctx, "cfgKey", cfgVal)
}

func ExampleDestroyResult() {
	ctx := context.Background()
	// select an existing stack
	s, _ := SelectStackLocalSource(ctx, "o/p/s", filepath.Join(".", "project"))
	destroyRes, _ := s.Destroy(ctx)
	// success!
	fmt.Println(destroyRes.Summary.Result)
}

func ExampleGitRepo() {
	ctx := context.Background()
	pName := "go_remote_proj"
	fqsn := FullyQualifiedStackName("myOrg", pName, "myStack")

	// we'll compile a the program into an executable with the name "examplesBinary"
	binName := "examplesBinary"
	repo := GitRepo{
		URL:         "https://github.com/pulumi/test-repo.git",
		ProjectPath: "goproj",
		// this call back will get executed post-clone to allow for additional program setup
		Setup: func(ctx context.Context, workspace Workspace) error {
			cmd := exec.Command("go", "build", "-o", binName, "main.go")
			cmd.Dir = workspace.WorkDir()
			return cmd.Run()
		},
	}
	// an override to the project file in the git repo, specifying our pre-built executable
	project := workspace.Project{
		Name: tokens.PackageName(pName),
		Runtime: workspace.NewProjectRuntimeInfo("go", map[string]interface{}{
			"binary": binName,
		}),
	}

	// initialize a stack from the git repo, specifying our project override
	NewStackRemoteSource(ctx, fqsn, repo, Project(project))
}

func ExampleLocalWorkspace() {
	ctx := context.Background()
	// create a workspace from a local project
	w, _ := NewLocalWorkspace(ctx, WorkDir(filepath.Join(".", "program")))
	// install the pulumi-aws plugin
	w.InstallPlugin(ctx, "aws", "v3.2.0")
	// fetch the config used with the last update on stack "org/proj/stack"
	w.RefreshConfig(ctx, "org/proj/stack")
}

func ExampleLocalWorkspace_CreateStack() {
	ctx := context.Background()
	// create a workspace from a local project
	w, _ := NewLocalWorkspace(ctx, WorkDir(filepath.Join(".", "program")))
	// initialize the stack
	err := w.CreateStack(ctx, "org/proj/stack")
	if err != nil {
		// failed to create the stack
	}
}

func ExampleLocalWorkspace_GetAllConfig() {
	ctx := context.Background()
	// create a workspace from a local project
	w, _ := NewLocalWorkspace(ctx, WorkDir(filepath.Join(".", "program")))
	// get all config from the workspace for the stack (read from Pulumi.stack.yaml)
	cfg, _ := w.GetAllConfig(ctx, "org/proj/stack")
	fmt.Println(cfg["config_key"].Value)
	fmt.Println(cfg["config_key"].Secret)
}

func ExampleLocalWorkspace_GetConfig() {
	ctx := context.Background()
	// create a workspace from a local project
	w, _ := NewLocalWorkspace(ctx, WorkDir(filepath.Join(".", "program")))
	// get all config from the workspace for the stack (read from Pulumi.stack.yaml)
	cfgVal, _ := w.GetConfig(ctx, "org/proj/stack", "config_key")
	fmt.Println(cfgVal.Value)
	fmt.Println(cfgVal.Secret)
}

func ExampleLocalWorkspace_GetEnvVars() {
	ctx := context.Background()
	// create a workspace from a local project
	w, _ := NewLocalWorkspace(ctx, WorkDir(filepath.Join(".", "program")))
	// get all environment values from the workspace for the stack
	e := w.GetEnvVars()
	for k, v := range e {
		fmt.Println(k)
		fmt.Println(v)
	}
}

func ExampleLocalWorkspace_SetEnvVars() {
	ctx := context.Background()
	// create a workspace from a local project
	w, _ := NewLocalWorkspace(ctx, WorkDir(filepath.Join(".", "program")))
	// set the environment values on the workspace for the stack
	e := map[string]string{
		"FOO": "bar",
	}
	w.SetEnvVars(e)
}

func ExampleLocalWorkspace_SetEnvVar() {
	ctx := context.Background()
	// create a workspace from a local project
	w, _ := NewLocalWorkspace(ctx, WorkDir(filepath.Join(".", "program")))
	// set the environment value on the workspace for the stack
	w.SetEnvVar("FOO", "bar")
}

func ExampleLocalWorkspace_UnsetEnvVar() {
	ctx := context.Background()
	// create a workspace from a local project
	w, _ := NewLocalWorkspace(ctx, WorkDir(filepath.Join(".", "program")))
	// set the environment value on the workspace for the stack
	w.SetEnvVar("FOO", "bar")
	// unset the environment value on the workspace for the stack
	w.UnsetEnvVar("FOO")
}

func ExampleLocalWorkspace_InstallPlugin() {
	ctx := context.Background()
	// create a workspace from a local project
	w, _ := NewLocalWorkspace(ctx, WorkDir(filepath.Join(".", "program")))
	w.InstallPlugin(ctx, "aws", "v3.2.0")
}

func ExampleNewLocalWorkspace() {
	ctx := context.Background()
	// WorkDir sets the working directory for the LocalWorkspace. The workspace will look for a default
	// project settings file (Pulumi.yaml) in this location for information about the Pulumi program.
	wd := WorkDir(filepath.Join("..", "path", "to", "pulumi", "project"))
	// PulumiHome customizes the location of $PULUMI_HOME where metadata is stored and plugins are installed.
	ph := PulumiHome(filepath.Join("~", ".pulumi"))
	// Project provides ProjectSettings to set once the workspace is created.
	proj := Project(workspace.Project{
		Name:    tokens.PackageName("myproject"),
		Runtime: workspace.NewProjectRuntimeInfo("go", nil),
		Backend: &workspace.ProjectBackend{
			URL: "https://url.to.custom.saas.backend.com",
		},
	})
	_, _ = NewLocalWorkspace(ctx, wd, ph, proj)
}

func ExampleLocalWorkspace_ListPlugins() {
	ctx := context.Background()
	// create a workspace from a local project
	w, _ := NewLocalWorkspace(ctx, WorkDir(filepath.Join(".", "program")))
	ps, _ := w.ListPlugins(ctx)
	fmt.Println(ps[0].Size)
}

func ExampleLocalWorkspace_ListStacks() {
	ctx := context.Background()
	// create a workspace from a local project
	w, _ := NewLocalWorkspace(ctx, WorkDir(filepath.Join(".", "program")))
	ss, _ := w.ListStacks(ctx)
	fmt.Println(ss[0].UpdateInProgress)
}

func ExampleLocalWorkspace_ProjectSettings() {
	ctx := context.Background()
	// create a workspace from a local project
	w, _ := NewLocalWorkspace(ctx, WorkDir(filepath.Join(".", "program")))
	// read existing project settings if any
	ps, err := w.ProjectSettings(ctx)
	if err != nil {
		// no Pulumi.yaml was found, so create a default
		ps = &workspace.Project{
			Name:    tokens.PackageName("myproject"),
			Runtime: workspace.NewProjectRuntimeInfo("go", nil),
		}
	}
	// make some changes
	author := "pulumipus"
	ps.Author = &author
	// save the settings back to the Workspace
	w.SaveProjectSettings(ctx, ps)
}

func ExampleLocalWorkspace_SaveProjectSettings() {
	ctx := context.Background()
	// create a workspace from a local project
	w, _ := NewLocalWorkspace(ctx, WorkDir(filepath.Join(".", "program")))
	// read existing project settings if any
	ps, err := w.ProjectSettings(ctx)
	if err != nil {
		// no Pulumi.yaml was found, so create a default
		ps = &workspace.Project{
			Name:    tokens.PackageName("myproject"),
			Runtime: workspace.NewProjectRuntimeInfo("go", nil),
		}
	}
	// make some changes
	author := "pulumipus"
	ps.Author = &author
	// save the settings back to the Workspace
	w.SaveProjectSettings(ctx, ps)
}

func ExampleLocalWorkspace_RefreshConfig() {
	ctx := context.Background()
	// create a workspace from a local project
	w, _ := NewLocalWorkspace(ctx, WorkDir(filepath.Join(".", "program")))
	fqsnA := FullyQualifiedStackName("org", "proj", "stackA")
	// get the last deployed config from stack A
	// this overwrites config in pulumi.stackA.yaml
	cfg, _ := w.RefreshConfig(ctx, fqsnA)
	// add a key to the ConfigMap
	cfg["addition_config_key"] = ConfigValue{Value: "additional_config_value"}
	// create a new stack
	fqsnB := FullyQualifiedStackName("org", "proj", "stackB")
	w.CreateStack(ctx, fqsnB)
	// save the modified config to stackB
	w.SetAllConfig(ctx, fqsnB, cfg)
}

func ExampleLocalWorkspace_RemoveAllConfig() {
	ctx := context.Background()
	// create a workspace from a local project
	w, _ := NewLocalWorkspace(ctx, WorkDir(filepath.Join(".", "program")))
	fqsn := FullyQualifiedStackName("org", "proj", "stackA")
	// get all config currently set in the workspace
	cfg, _ := w.GetAllConfig(ctx, fqsn)
	var keys []string
	for k := range cfg {
		keys = append(keys, k)
	}
	// remove those config values
	_ = w.RemoveAllConfig(ctx, fqsn, keys)
}

func ExampleLocalWorkspace_RemoveConfig() {
	ctx := context.Background()
	// create a workspace from a local project
	w, _ := NewLocalWorkspace(ctx, WorkDir(filepath.Join(".", "program")))
	fqsn := FullyQualifiedStackName("org", "proj", "stackA")
	_ = w.RemoveConfig(ctx, fqsn, "key_to_remove")
}

func ExampleLocalWorkspace_RemovePlugin() {
	ctx := context.Background()
	// create a workspace from a local project
	w, _ := NewLocalWorkspace(ctx, WorkDir(filepath.Join(".", "program")))
	_ = w.RemovePlugin(ctx, "aws", "v3.2.0")
}

func ExampleLocalWorkspace_RemoveStack() {
	ctx := context.Background()
	// create a workspace from a local project
	w, _ := NewLocalWorkspace(ctx, WorkDir(filepath.Join(".", "program")))
	fqsn := FullyQualifiedStackName("org", "proj", "stack")
	_ = w.RemoveStack(ctx, fqsn)
}

func ExampleLocalWorkspace_SelectStack() {
	ctx := context.Background()
	// create a workspace from a local project
	w, _ := NewLocalWorkspace(ctx, WorkDir(filepath.Join(".", "program")))
	fqsn := FullyQualifiedStackName("org", "proj", "existing_stack")
	_ = w.SelectStack(ctx, fqsn)
}

func ExampleLocalWorkspace_SetAllConfig() {
	ctx := context.Background()
	// create a workspace from a local project
	w, _ := NewLocalWorkspace(ctx, WorkDir(filepath.Join(".", "program")))
	fqsnA := FullyQualifiedStackName("org", "proj", "stackA")
	// get the last deployed config from stack A
	// this overwrites config in pulumi.stackA.yaml
	cfg, _ := w.RefreshConfig(ctx, fqsnA)
	// add a key to the ConfigMap
	cfg["addition_config_key"] = ConfigValue{Value: "additional_config_value"}
	// create a new stack
	fqsnB := FullyQualifiedStackName("org", "proj", "stackB")
	w.CreateStack(ctx, fqsnB)
	// save the modified config to stackB
	w.SetAllConfig(ctx, fqsnB, cfg)
}

func ExampleLocalWorkspace_SetProgram() {
	ctx := context.Background()
	// create a workspace from a local project
	w, _ := NewLocalWorkspace(ctx, WorkDir(filepath.Join(".", "program")))
	program := func(pCtx *pulumi.Context) error {
		pCtx.Export("an output", pulumi.String("an output value"))
		return nil
	}
	// this program will be used for all Stack.Update and Stack.Preview operations going forward
	w.SetProgram(program)
}

func ExampleLocalWorkspace_Stack() {
	ctx := context.Background()
	// create a workspace from a local project
	w, _ := NewLocalWorkspace(ctx, WorkDir(filepath.Join(".", "program")))
	s, _ := w.Stack(ctx)
	fmt.Println(s.UpdateInProgress)
}

func ExampleLocalWorkspace_StackSettings() {
	ctx := context.Background()
	// create a workspace from a local project
	w, _ := NewLocalWorkspace(ctx, WorkDir(filepath.Join(".", "program")))
	fqsnA := FullyQualifiedStackName("org", "proj", "stackA")
	// read existing stack settings for stackA if any
	ss, err := w.StackSettings(ctx, fqsnA)
	if err != nil {
		// no pulumi.stackA.yaml was found, so create a default
		ss = &workspace.ProjectStack{}
	}
	fqsnB := FullyQualifiedStackName("org", "proj", "stackB")
	// copy the settings to stackB
	w.SaveStackSettings(ctx, fqsnB, ss)
}

func ExampleLocalWorkspace_SaveStackSettings() {
	ctx := context.Background()
	// create a workspace from a local project
	w, _ := NewLocalWorkspace(ctx, WorkDir(filepath.Join(".", "program")))
	fqsnA := FullyQualifiedStackName("org", "proj", "stackA")
	// read existing stack settings for stackA if any
	ss, err := w.StackSettings(ctx, fqsnA)
	if err != nil {
		// no pulumi.stackA.yaml was found, so create a default
		ss = &workspace.ProjectStack{}
	}
	fqsnB := FullyQualifiedStackName("org", "proj", "stackB")
	// copy the settings to stackB
	w.SaveStackSettings(ctx, fqsnB, ss)
}

func ExampleLocalWorkspace_WhoAmI() {
	ctx := context.Background()
	// create a workspace from a local project
	w, _ := NewLocalWorkspace(ctx, WorkDir(filepath.Join(".", "program")))
	user, _ := w.WhoAmI(ctx)
	fmt.Println(user)
}

func ExampleSetupFn() {
	ctx := context.Background()
	fqsn := FullyQualifiedStackName("myOrg", "nodejs_project", "myStack")

	repo := GitRepo{
		URL:         "https://some.example.repo.git",
		ProjectPath: "goproj",
		// this call back will get executed post-clone to allow for additional program setup
		Setup: func(ctx context.Context, workspace Workspace) error {
			// restore node dependencies required to run the program
			cmd := exec.Command("npm", "install")
			cmd.Dir = workspace.WorkDir()
			return cmd.Run()
		},
	}

	// initialize a stack from the git repo
	NewStackRemoteSource(ctx, fqsn, repo)
}

func ExampleStack() error {
	ctx := context.Background()
	fqsn := FullyQualifiedStackName("org", "project", "dev_stack")
	cfg := ConfigMap{
		"bar": ConfigValue{
			Value: "abc",
		},
		"buzz": ConfigValue{
			Value:  "secret",
			Secret: true,
		},
	}

	// initialize
	pDir := filepath.Join(".", "project")
	s, err := NewStackLocalSource(ctx, fqsn, pDir)
	if err != nil {
		return err
	}

	defer func() {
		// Workspace operations can be accessed via Stack.Workspace()
		// -- pulumi stack rm --
		err = s.Workspace().RemoveStack(ctx, s.Name())
	}()

	err = s.SetAllConfig(ctx, cfg)
	if err != nil {
		return err
	}

	// -- pulumi up --
	res, err := s.Up(ctx)
	if err != nil {
		return err
	}
	fmt.Println(len(res.Outputs))

	// -- pulumi preview --

	prev, err := s.Preview(ctx)
	if err != nil {
		return err
	}
	// no changes after the update
	fmt.Println(prev.ChangeSummary["same"])

	// -- pulumi refresh --

	ref, err := s.Refresh(ctx)

	if err != nil {
		return err
	}
	// Success!
	fmt.Println(ref.Summary.Result)

	// -- pulumi destroy --

	_, err = s.Destroy(ctx)
	if err != nil {
		return err
	}
	return nil
}

func ExampleNewStack() error {
	ctx := context.Background()
	w, err := NewLocalWorkspace(ctx)
	if err != nil {
		return err
	}

	fqsn := FullyQualifiedStackName("org", "proj", "stack")
	stack, err := NewStack(ctx, fqsn, w)
	if err != nil {
		return err
	}
	_, err = stack.Up(ctx)
	return err
}

func ExampleNewStackInlineSource() {
	ctx := context.Background()
	fqsn := FullyQualifiedStackName("myOrg", "proj", "stack")
	// create a new Stack with a default ProjectSettings file, and a temporary WorkDir created in a new
	// LocalWorkspace on behalf of the user.
	stack, _ := NewStackInlineSource(ctx, fqsn, func(pCtx *pulumi.Context) error {
		pCtx.Export("outputA", pulumi.String("valueA"))
		return nil
	})
	// Stack.Up runs the inline pulumi.RunFunc specified above.
	stack.Up(ctx)
	// we can update the Workspace program for subsequent updates if desired
	stack.Workspace().SetProgram(func(pCtx *pulumi.Context) error {
		pCtx.Export("outputAA", pulumi.String("valueAA"))
		return nil
	})
	stack.Up(ctx)
}

func ExampleSelectStackInlineSource() {
	ctx := context.Background()
	fqsn := FullyQualifiedStackName("myOrg", "proj", "existing_stack")
	// selects an existing stack with a default ProjectSettings file, and a temporary WorkDir in a new LocalWorkspace
	// created on behalf of the user.
	stack, _ := SelectStackInlineSource(ctx, fqsn, func(pCtx *pulumi.Context) error {
		pCtx.Export("outputA", pulumi.String("valueA"))
		return nil
	})
	// Stack.Up runs the inline pulumi.RunFunc specified above.
	stack.Up(ctx)
	// we can update the Workspace program for subsequent updates if desired
	stack.Workspace().SetProgram(func(pCtx *pulumi.Context) error {
		pCtx.Export("outputAA", pulumi.String("valueAA"))
		return nil
	})
	stack.Up(ctx)
}

func ExampleNewStackLocalSource() {
	ctx := context.Background()
	fqsn := FullyQualifiedStackName("myOrg", "proj", "stack")
	workDir := filepath.Join(".", "program", "dir")
	// creates a new stack with a new LocalWorkspace created from provided WorkDir. This Workspace will pick up
	// any available Settings files (Pulumi.yaml, Pulumi.<stack>.yaml)
	stack, _ := NewStackLocalSource(ctx, fqsn, workDir)
	// Stack.Up runs the program in workDir
	stack.Up(ctx)
	// we can update the Workspace program for subsequent updates if desired
	stack.Workspace().SetProgram(func(pCtx *pulumi.Context) error {
		pCtx.Export("outputAA", pulumi.String("valueAA"))
		return nil
	})
	// Stack.Up now runs our inline program
	stack.Up(ctx)
}

func ExampleSelectStackLocalSource() {
	ctx := context.Background()
	fqsn := FullyQualifiedStackName("myOrg", "proj", "stack")
	workDir := filepath.Join(".", "program", "dir")
	// selects an existing stack with a new LocalWorkspace created from provided WorkDir. This Workspace will pick up
	// any available Settings files (Pulumi.yaml, Pulumi.<stack>.yaml)
	stack, _ := SelectStackLocalSource(ctx, fqsn, workDir)
	// Stack.Up runs the program in workDir
	stack.Up(ctx)
	// we can update the Workspace program for subsequent updates if desired
	stack.Workspace().SetProgram(func(pCtx *pulumi.Context) error {
		pCtx.Export("outputAA", pulumi.String("valueAA"))
		return nil
	})
	// Stack.Up now runs our inline program
	stack.Up(ctx)
}

func ExampleNewStackRemoteSource() {
	ctx := context.Background()
	pName := "go_remote_proj"
	fqsn := FullyQualifiedStackName("myOrg", pName, "myStack")

	// we'll compile a the program into an executable with the name "examplesBinary"
	binName := "examplesBinary"
	// a description of the git repo to clone
	repo := GitRepo{
		URL: "https://github.com/pulumi/test-repo.git",
		// the subdirectory relative to the root of the repo
		ProjectPath: "goproj",
		// this call back will get executed post-clone to allow for additional program setup
		Setup: func(ctx context.Context, workspace Workspace) error {
			cmd := exec.Command("go", "build", "-o", binName, "main.go")
			cmd.Dir = workspace.WorkDir()
			return cmd.Run()
		},
	}
	// an override to the project file in the git repo, specifying our pre-built executable
	project := workspace.Project{
		Name: tokens.PackageName(pName),
		Runtime: workspace.NewProjectRuntimeInfo("go", map[string]interface{}{
			"binary": binName,
		}),
	}

	// initialize a stack from the git repo, specifying our project override
	NewStackRemoteSource(ctx, fqsn, repo, Project(project))
}

func ExampleSelectStackRemoteSource() {
	ctx := context.Background()
	pName := "go_remote_proj"
	fqsn := FullyQualifiedStackName("myOrg", pName, "myStack")

	// we'll compile a the program into an executable with the name "examplesBinary"
	binName := "examplesBinary"
	// a description of the git repo to clone
	repo := GitRepo{
		URL: "https://github.com/pulumi/test-repo.git",
		// the subdirectory relative to the root of the repo
		ProjectPath: "goproj",
		// this call back will get executed post-clone to allow for additional program setup
		Setup: func(ctx context.Context, workspace Workspace) error {
			cmd := exec.Command("go", "build", "-o", binName, "main.go")
			cmd.Dir = workspace.WorkDir()
			return cmd.Run()
		},
	}
	// an override to the project file in the git repo, specifying our pre-built executable
	project := workspace.Project{
		Name: tokens.PackageName(pName),
		Runtime: workspace.NewProjectRuntimeInfo("go", map[string]interface{}{
			"binary": binName,
		}),
	}

	// select an existing stack using a LocalWorkspace created from the git repo, specifying our project override
	SelectStackRemoteSource(ctx, fqsn, repo, Project(project))
}

func ExampleStack_Destroy() {
	ctx := context.Background()
	fqsn := FullyQualifiedStackName("org", "project", "stack")
	// select an existing stack to destroy
	stack, _ := SelectStackLocalSource(ctx, fqsn, filepath.Join(".", "program"))
	stack.Destroy(ctx, optdestroy.Message("a message to save with the destroy operation"))
}

func ExampleStack_Up() {
	ctx := context.Background()
	fqsn := FullyQualifiedStackName("org", "project", "stack")
	// create a new stack to update
	stack, _ := NewStackLocalSource(ctx, fqsn, filepath.Join(".", "program"))
	stack.Up(ctx, optup.Message("a message to save with the up operation"), optup.Parallel(10000))
}

func ExampleStack_Preview() {
	ctx := context.Background()
	fqsn := FullyQualifiedStackName("org", "project", "stack")
	// create a new stack and preview changes
	stack, _ := NewStackLocalSource(ctx, fqsn, filepath.Join(".", "program"))
	stack.Preview(ctx, optpreview.Message("a message to save with the preive operation"))
}

func ExampleStack_Refresh() {
	ctx := context.Background()
	fqsn := FullyQualifiedStackName("org", "project", "stack")
	// select an existing stack and refresh the resources under management
	stack, _ := SelectStackLocalSource(ctx, fqsn, filepath.Join(".", "program"))
	stack.Refresh(ctx, optrefresh.Message("a message to save with the refresh operation"))
}

func ExampleStack_GetAllConfig() {
	ctx := context.Background()
	fqsn := FullyQualifiedStackName("org", "project", "stack")
	stack, _ := SelectStackLocalSource(ctx, fqsn, filepath.Join(".", "program"))
	cfg, _ := stack.GetAllConfig(ctx)
	fmt.Println(cfg["config_key"].Value)
	fmt.Println(cfg["config_key"].Secret)
}

func ExampleStack_GetConfig() {
	ctx := context.Background()
	fqsn := FullyQualifiedStackName("org", "project", "stack")
	stack, _ := SelectStackLocalSource(ctx, fqsn, filepath.Join(".", "program"))
	cfgVal, _ := stack.GetConfig(ctx, "config_key")
	fmt.Println(cfgVal.Value)
	fmt.Println(cfgVal.Secret)
}

func ExampleStack_History() {
	ctx := context.Background()
	fqsn := FullyQualifiedStackName("org", "project", "stack")
	stack, _ := SelectStackLocalSource(ctx, fqsn, filepath.Join(".", "program"))
	hist, _ := stack.History(ctx)
	// last operation start time
	fmt.Println(hist[0].StartTime)
}

func ExampleStack_Info() {
	ctx := context.Background()
	fqsn := FullyQualifiedStackName("org", "project", "stack")
	stack, _ := SelectStackLocalSource(ctx, fqsn, filepath.Join(".", "program"))
	info, _ := stack.Info(ctx)
	// url to view the Stack in the Pulumi SaaS or other backend
	fmt.Println(info.URL)
}

func ExampleStack_Outputs() {
	ctx := context.Background()
	fqsn := FullyQualifiedStackName("org", "project", "stack")
	stack, _ := SelectStackLocalSource(ctx, fqsn, filepath.Join(".", "program"))
	outs, _ := stack.Outputs(ctx)
	fmt.Println(outs["key"].Value.(string))
	fmt.Println(outs["key"].Secret)
}

func ExampleStack_RefreshConfig() {
	ctx := context.Background()
	fqsn := FullyQualifiedStackName("org", "project", "stack")
	stack, _ := SelectStackLocalSource(ctx, fqsn, filepath.Join(".", "program"))
	// get the last deployed config from stack and overwrite Pulumi.stack.yaml
	_, _ = stack.RefreshConfig(ctx)
}

func ExampleStack_RemoveConfig() {
	ctx := context.Background()
	fqsn := FullyQualifiedStackName("org", "project", "stack")
	stack, _ := SelectStackLocalSource(ctx, fqsn, filepath.Join(".", "program"))
	// remove config matching the specified key
	_ = stack.RemoveConfig(ctx, "key_to_remove")
}

func ExampleStack_RemoveAllConfig() {
	ctx := context.Background()
	fqsn := FullyQualifiedStackName("org", "project", "stack")
	stack, _ := SelectStackLocalSource(ctx, fqsn, filepath.Join(".", "program"))
	// remove config matching the specified keys
	_ = stack.RemoveAllConfig(ctx, []string{"key0", "key1", "...", "keyN"})
}

func ExampleStack_SetConfig() {
	ctx := context.Background()
	fqsn := FullyQualifiedStackName("org", "project", "stack")
	stack, _ := SelectStackLocalSource(ctx, fqsn, filepath.Join(".", "program"))
	// set config key-value pair
	_ = stack.SetConfig(ctx, "key_to_set", ConfigValue{Value: "abc", Secret: true})
}

func ExampleStack_SetAllConfig() {
	ctx := context.Background()
	fqsn := FullyQualifiedStackName("org", "project", "stack")
	stack, _ := SelectStackLocalSource(ctx, fqsn, filepath.Join(".", "program"))
	cfg := ConfigMap{
		"key0": ConfigValue{Value: "abc", Secret: true},
		"key1": ConfigValue{Value: "def"},
	}
	// set all config in map
	_ = stack.SetAllConfig(ctx, cfg)
}