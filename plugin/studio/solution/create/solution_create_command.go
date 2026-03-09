// Package create implements the command plugin for creating a new UiPath solution
// with an agent project scaffold.
package create

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/plugin"
)

const defaultModel = "anthropic.claude-haiku-4-5-20251001-v1:0"

// The SolutionCreateCommand creates a new solution with an agent project.
type SolutionCreateCommand struct{}

func (c SolutionCreateCommand) Command() plugin.Command {
	return *plugin.NewCommand("studio").
		WithCategory("solution", "UiPath Solution management", "Pack, unpack, push and pull UiPath Maestro solutions.").
		WithOperation("create", "Create Solution", "Creates a new UiPath solution with an agent project").
		WithParameter(plugin.NewParameter("name", plugin.ParameterTypeString, "Solution name").
			WithRequired(true)).
		WithParameter(plugin.NewParameter("project-name", plugin.ParameterTypeString, "Agent project name").
			WithDefaultValue("Agent")).
		WithParameter(plugin.NewParameter("model", plugin.ParameterTypeString, "LLM model identifier").
			WithDefaultValue(defaultModel)).
		WithParameter(plugin.NewParameter("system-prompt", plugin.ParameterTypeString, "System prompt for the agent").
			WithDefaultValue("You are a helpful assistant.")).
		WithParameter(plugin.NewParameter("destination", plugin.ParameterTypeString, "Parent directory for the solution").
			WithDefaultValue("."))
}

func (c SolutionCreateCommand) Execute(ctx plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	name := c.getStringParameter("name", "", ctx.Parameters)
	if name == "" {
		return errors.New("Solution name is required")
	}
	projectName := c.getStringParameter("project-name", "Agent", ctx.Parameters)
	model := c.getStringParameter("model", defaultModel, ctx.Parameters)
	systemPrompt := c.getStringParameter("system-prompt", "You are a helpful assistant.", ctx.Parameters)
	destination := c.getStringParameter("destination", ".", ctx.Parameters)
	destination, _ = filepath.Abs(destination)

	solutionDir := filepath.Join(destination, name)
	if _, err := os.Stat(solutionDir); err == nil {
		return fmt.Errorf("Directory already exists: %s", solutionDir)
	}

	ids := c.generateIds()
	err := c.createSolution(solutionDir, name, projectName, model, systemPrompt, ids)
	if err != nil {
		return err
	}

	result := struct {
		Status      string `json:"status"`
		Directory   string `json:"directory"`
		SolutionId  string `json:"solutionId"`
		ProjectId   string `json:"projectId"`
		Name        string `json:"name"`
		ProjectName string `json:"projectName"`
	}{"Succeeded", solutionDir, ids.solutionId, ids.projectId, name, projectName}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("Create command failed: %w", err)
	}
	return writer.WriteResponse(*output.NewResponseInfo(http.StatusOK, "200 OK", "HTTP/1.1", map[string][]string{}, bytes.NewReader(jsonData)))
}

type solutionIds struct {
	solutionId   string
	projectId    string
	projectKey   string
	packageKey   string
	processKey   string
	entryPointId string
	evalSetId    string
	evaluatorId  string
	trajectoryId string
}

func (c SolutionCreateCommand) generateIds() solutionIds {
	return solutionIds{
		solutionId:   c.generateUUID(),
		projectId:    c.generateUUID(),
		projectKey:   c.generateUUID(),
		packageKey:   c.generateUUID(),
		processKey:   c.generateUUID(),
		entryPointId: c.generateUUID(),
		evalSetId:    c.generateUUID(),
		evaluatorId:  c.generateUUID(),
		trajectoryId: c.generateUUID(),
	}
}

func (c SolutionCreateCommand) generateUUID() string {
	b := make([]byte, 16)
	_, _ = io.ReadFull(rand.Reader, b)
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func (c SolutionCreateCommand) createSolution(solutionDir string, name string, projectName string, model string, systemPrompt string, ids solutionIds) error {
	projectDir := filepath.Join(solutionDir, projectName)
	dirs := []string{
		projectDir,
		filepath.Join(projectDir, ".agent-builder"),
		filepath.Join(projectDir, ".project"),
		filepath.Join(projectDir, "evals", "eval-sets"),
		filepath.Join(projectDir, "evals", "evaluators"),
		filepath.Join(solutionDir, "resources", "solution_folder", "package"),
		filepath.Join(solutionDir, "resources", "solution_folder", "process", "agent"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0750); err != nil {
			return fmt.Errorf("Cannot create directory '%s': %w", dir, err)
		}
	}

	writers := []func() error{
		func() error { return c.writeSolutionStorage(solutionDir, ids, projectName) },
		func() error { return c.writeManifest(solutionDir, name, ids, projectName) },
		func() error { return c.writeProjectDescriptor(projectDir, projectName) },
		func() error { return c.writeAgentJson(projectDir, ids, model, systemPrompt) },
		func() error { return c.writeEntryPoints(projectDir, ids) },
		func() error { return c.writeEmptyJson(filepath.Join(projectDir, "flow-layout.json")) },
		func() error { return c.writeEmptyJson(filepath.Join(projectDir, ".project", "JitCustomTypes.json")) },
		func() error { return c.writeAgentBuilderJson(projectDir, projectName, ids, model, systemPrompt) },
		func() error { return c.writeBindings(projectDir) },
		func() error { return c.writeAgentBuilderEntryPoints(projectDir, ids) },
		func() error { return c.writeEvalSet(projectDir, ids) },
		func() error { return c.writeEvaluator(projectDir, ids) },
		func() error { return c.writeTrajectoryEvaluator(projectDir, ids) },
		func() error { return c.writePackageResource(solutionDir, projectName, ids) },
		func() error { return c.writeProcessResource(solutionDir, name, projectName, ids) },
	}
	for _, w := range writers {
		if err := w(); err != nil {
			return err
		}
	}
	return nil
}

func (c SolutionCreateCommand) writeJSONFile(path string, data interface{}) error {
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("Cannot create %s: %w", filepath.Base(path), err)
	}
	return os.WriteFile(path, append(content, '\n'), 0600)
}

func (c SolutionCreateCommand) writeEmptyJson(path string) error {
	return os.WriteFile(path, []byte("{}\n"), 0600)
}

// --- Solution-level files ---

func (c SolutionCreateCommand) writeSolutionStorage(dir string, ids solutionIds, projectName string) error {
	data := struct {
		SolutionId string `json:"SolutionId"`
		Projects   []struct {
			ProjectId           string `json:"ProjectId"`
			ProjectRelativePath string `json:"ProjectRelativePath"`
		} `json:"Projects"`
	}{
		SolutionId: ids.solutionId,
		Projects: []struct {
			ProjectId           string `json:"ProjectId"`
			ProjectRelativePath string `json:"ProjectRelativePath"`
		}{
			{ids.projectId, projectName + "/project.uiproj"},
		},
	}
	return c.writeJSONFile(filepath.Join(dir, "SolutionStorage.json"), data)
}

func (c SolutionCreateCommand) writeManifest(dir string, name string, ids solutionIds, projectName string) error {
	data := struct {
		DocVersion       string `json:"DocVersion"`
		StudioMinVersion string `json:"StudioMinVersion"`
		SolutionId       string `json:"SolutionId"`
		Projects         []struct {
			Type                string `json:"Type"`
			ProjectRelativePath string `json:"ProjectRelativePath"`
			Id                  string `json:"Id"`
		} `json:"Projects"`
	}{
		DocVersion:       "1.0.0",
		StudioMinVersion: "2025.04.0",
		SolutionId:       ids.solutionId,
		Projects: []struct {
			Type                string `json:"Type"`
			ProjectRelativePath string `json:"ProjectRelativePath"`
			Id                  string `json:"Id"`
		}{
			{"Agent", projectName + "/project.uiproj", ids.projectKey},
		},
	}
	return c.writeJSONFile(filepath.Join(dir, name+".uipx"), data)
}

// --- Project files ---

func (c SolutionCreateCommand) writeProjectDescriptor(dir string, projectName string) error {
	data := struct {
		ProjectType string  `json:"ProjectType"`
		Name        string  `json:"Name"`
		Description *string `json:"Description"`
		MainFile    *string `json:"MainFile"`
	}{
		ProjectType: "Agent",
		Name:        projectName,
	}
	return c.writeJSONFile(filepath.Join(dir, "project.uiproj"), data)
}

type agentSettings struct {
	Model         string `json:"model"`
	MaxTokens     int    `json:"maxTokens"`
	Temperature   int    `json:"temperature"`
	Engine        string `json:"engine"`
	MaxIterations int    `json:"maxIterations"`
}

type agentMessage struct {
	Role          string `json:"role"`
	Content       string `json:"content"`
	ContentTokens []struct {
		Type      string `json:"type"`
		RawString string `json:"rawString"`
	} `json:"contentTokens"`
}

type jsonSchemaProperty struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

func (c SolutionCreateCommand) newDefaultSettings(model string) agentSettings {
	return agentSettings{
		Model:         model,
		MaxTokens:     16384,
		Temperature:   0,
		Engine:        "basic-v2",
		MaxIterations: 25,
	}
}

func (c SolutionCreateCommand) newMessages(systemPrompt string) []agentMessage {
	return []agentMessage{
		{
			Role:    "system",
			Content: systemPrompt,
			ContentTokens: []struct {
				Type      string `json:"type"`
				RawString string `json:"rawString"`
			}{
				{"simpleText", systemPrompt},
			},
		},
		{
			Role:    "user",
			Content: "query: {{query}}",
			ContentTokens: []struct {
				Type      string `json:"type"`
				RawString string `json:"rawString"`
			}{
				{"simpleText", "query: "},
				{"variable", "input.query"},
				{"simpleText", ""},
			},
		},
	}
}

func (c SolutionCreateCommand) newInputSchema() interface{} {
	return struct {
		Type       string                        `json:"type"`
		Properties map[string]jsonSchemaProperty  `json:"properties"`
		Required   []string                       `json:"required"`
	}{
		Type: "object",
		Properties: map[string]jsonSchemaProperty{
			"query": {Type: "string"},
		},
		Required: []string{"query"},
	}
}

func (c SolutionCreateCommand) newOutputSchema() interface{} {
	return struct {
		Type       string                        `json:"type"`
		Properties map[string]jsonSchemaProperty  `json:"properties"`
	}{
		Type: "object",
		Properties: map[string]jsonSchemaProperty{
			"content": {Type: "string", Description: "Output content"},
		},
	}
}

func (c SolutionCreateCommand) writeAgentJson(dir string, ids solutionIds, model string, systemPrompt string) error {
	data := struct {
		Version      string        `json:"version"`
		Settings     agentSettings `json:"settings"`
		InputSchema  interface{}   `json:"inputSchema"`
		OutputSchema interface{}   `json:"outputSchema"`
		Metadata     struct {
			StorageVersion                string `json:"storageVersion"`
			IsConversational              bool   `json:"isConversational"`
			ShowProjectCreationExperience bool   `json:"showProjectCreationExperience"`
			TargetRuntime                 string `json:"targetRuntime"`
		} `json:"metadata"`
		Type      string         `json:"type"`
		ProjectId string         `json:"projectId"`
		Messages  []agentMessage `json:"messages"`
	}{
		Version:      "1.1.0",
		Settings:     c.newDefaultSettings(model),
		InputSchema:  c.newInputSchema(),
		OutputSchema: c.newOutputSchema(),
		Metadata: struct {
			StorageVersion                string `json:"storageVersion"`
			IsConversational              bool   `json:"isConversational"`
			ShowProjectCreationExperience bool   `json:"showProjectCreationExperience"`
			TargetRuntime                 string `json:"targetRuntime"`
		}{
			StorageVersion:                "44.0.0",
			IsConversational:              false,
			ShowProjectCreationExperience: true,
			TargetRuntime:                 "pythonAgent",
		},
		Type:      "lowCode",
		ProjectId: ids.projectId,
		Messages:  c.newMessages(systemPrompt),
	}
	return c.writeJSONFile(filepath.Join(dir, "agent.json"), data)
}

func (c SolutionCreateCommand) writeEntryPoints(dir string, ids solutionIds) error {
	data := struct {
		Schema      string `json:"$schema"`
		Id          string `json:"$id"`
		EntryPoints []struct {
			UniqueId string      `json:"uniqueId"`
			Type     string      `json:"type"`
			Input    interface{} `json:"input"`
			Output   interface{} `json:"output"`
		} `json:"entryPoints"`
	}{
		Schema: "https://cloud.uipath.com/draft/2024-12/entry-point",
		Id:     "entry-points.json",
		EntryPoints: []struct {
			UniqueId string      `json:"uniqueId"`
			Type     string      `json:"type"`
			Input    interface{} `json:"input"`
			Output   interface{} `json:"output"`
		}{
			{
				UniqueId: ids.entryPointId,
				Type:     "agent",
				Input:    c.newInputSchema(),
				Output:   c.newOutputSchema(),
			},
		},
	}
	return c.writeJSONFile(filepath.Join(dir, "entry-points.json"), data)
}

// --- Agent builder files ---

func (c SolutionCreateCommand) writeAgentBuilderJson(dir string, projectName string, ids solutionIds, model string, systemPrompt string) error {
	data := struct {
		Id           string        `json:"id"`
		Version      string        `json:"version"`
		Name         string        `json:"name"`
		Metadata     struct {
			StorageVersion                string `json:"storageVersion"`
			IsConversational              bool   `json:"isConversational"`
			ShowProjectCreationExperience bool   `json:"showProjectCreationExperience"`
		} `json:"metadata"`
		Messages     []agentMessage `json:"messages"`
		InputSchema  interface{}    `json:"inputSchema"`
		OutputSchema interface{}    `json:"outputSchema"`
		Settings     agentSettings  `json:"settings"`
		Resources    []interface{}  `json:"resources"`
		Features     []interface{}  `json:"features"`
	}{
		Id:      ids.projectId,
		Version: "1.1.0",
		Name:    projectName,
		Metadata: struct {
			StorageVersion                string `json:"storageVersion"`
			IsConversational              bool   `json:"isConversational"`
			ShowProjectCreationExperience bool   `json:"showProjectCreationExperience"`
		}{
			StorageVersion:                "44.0.0",
			IsConversational:              false,
			ShowProjectCreationExperience: true,
		},
		Messages:     c.newMessages(systemPrompt),
		InputSchema:  c.newInputSchema(),
		OutputSchema: c.newOutputSchema(),
		Settings:     c.newDefaultSettings(model),
		Resources:    []interface{}{},
		Features:     []interface{}{},
	}
	return c.writeJSONFile(filepath.Join(dir, ".agent-builder", "agent.json"), data)
}

func (c SolutionCreateCommand) writeBindings(dir string) error {
	data := struct {
		Version   string        `json:"version"`
		Resources []interface{} `json:"resources"`
	}{"2.0", []interface{}{}}
	return c.writeJSONFile(filepath.Join(dir, ".agent-builder", "bindings.json"), data)
}

func (c SolutionCreateCommand) writeAgentBuilderEntryPoints(dir string, ids solutionIds) error {
	return c.writeEntryPoints(filepath.Join(dir, ".agent-builder"), ids)
}

// --- Evaluation files ---

func (c SolutionCreateCommand) writeEvalSet(dir string, ids solutionIds) error {
	data := struct {
		FileName      string        `json:"fileName"`
		Id            string        `json:"id"`
		Name          string        `json:"name"`
		BatchSize     int           `json:"batchSize"`
		EvaluatorRefs []string      `json:"evaluatorRefs"`
		Evaluations   []interface{} `json:"evaluations"`
	}{
		FileName:      "evaluation-set-default.json",
		Id:            ids.evalSetId,
		Name:          "Default Evaluation Set",
		BatchSize:     10,
		EvaluatorRefs: []string{ids.evaluatorId},
		Evaluations:   []interface{}{},
	}
	return c.writeJSONFile(filepath.Join(dir, "evals", "eval-sets", "evaluation-set-default.json"), data)
}

func (c SolutionCreateCommand) writeEvaluator(dir string, ids solutionIds) error {
	data := struct {
		Version         string `json:"version"`
		Id              string `json:"id"`
		Description     string `json:"description"`
		EvaluatorTypeId string `json:"evaluatorTypeId"`
		EvaluatorConfig struct {
			Name                      string      `json:"name"`
			TargetOutputKey           string      `json:"targetOutputKey"`
			Model                     string      `json:"model"`
			Prompt                    string      `json:"prompt"`
			Temperature               float64     `json:"temperature"`
			DefaultEvaluationCriteria interface{} `json:"defaultEvaluationCriteria"`
		} `json:"evaluatorConfig"`
	}{
		Version:         "1.0",
		Id:              ids.evaluatorId,
		Description:     "Uses an LLM to judge semantic similarity.",
		EvaluatorTypeId: "uipath-llm-judge-output-semantic-similarity",
		EvaluatorConfig: struct {
			Name                      string      `json:"name"`
			TargetOutputKey           string      `json:"targetOutputKey"`
			Model                     string      `json:"model"`
			Prompt                    string      `json:"prompt"`
			Temperature               float64     `json:"temperature"`
			DefaultEvaluationCriteria interface{} `json:"defaultEvaluationCriteria"`
		}{
			Name:            "SemanticSimilarityEvaluator",
			TargetOutputKey: "*",
			Model:           "gpt-4.1-2025-04-14",
			Prompt:          "Compare the outputs and evaluate semantic similarity.\n\nActual: {{ActualOutput}}\nExpected: {{ExpectedOutput}}\n\nScore 0-100.",
			Temperature:     0.0,
			DefaultEvaluationCriteria: struct {
				ExpectedOutput struct {
					Content string `json:"content"`
				} `json:"expectedOutput"`
			}{},
		},
	}
	return c.writeJSONFile(filepath.Join(dir, "evals", "evaluators", "evaluator-default.json"), data)
}

func (c SolutionCreateCommand) writeTrajectoryEvaluator(dir string, ids solutionIds) error {
	data := struct {
		Version         string `json:"version"`
		Id              string `json:"id"`
		Description     string `json:"description"`
		EvaluatorTypeId string `json:"evaluatorTypeId"`
		EvaluatorConfig struct {
			Name                      string      `json:"name"`
			Model                     string      `json:"model"`
			Prompt                    string      `json:"prompt"`
			Temperature               float64     `json:"temperature"`
			DefaultEvaluationCriteria interface{} `json:"defaultEvaluationCriteria"`
		} `json:"evaluatorConfig"`
	}{
		Version:         "1.0",
		Id:              ids.trajectoryId,
		Description:     "Evaluates agent execution trajectory.",
		EvaluatorTypeId: "uipath-llm-judge-trajectory-similarity",
		EvaluatorConfig: struct {
			Name                      string      `json:"name"`
			Model                     string      `json:"model"`
			Prompt                    string      `json:"prompt"`
			Temperature               float64     `json:"temperature"`
			DefaultEvaluationCriteria interface{} `json:"defaultEvaluationCriteria"`
		}{
			Name:        "TrajectoryEvaluator",
			Model:       "gpt-4.1-2025-04-14",
			Prompt:      "Evaluate trajectory.\n\nExpected: {{ExpectedAgentBehavior}}\nHistory: {{AgentRunHistory}}\n\nScore 0-100.",
			Temperature: 0.0,
			DefaultEvaluationCriteria: struct {
				ExpectedAgentBehavior string `json:"expectedAgentBehavior"`
			}{"The agent should correctly perform the task."},
		},
	}
	return c.writeJSONFile(filepath.Join(dir, "evals", "evaluators", "evaluator-default-trajectory.json"), data)
}

// --- Resource files ---

func (c SolutionCreateCommand) writePackageResource(solutionDir string, projectName string, ids solutionIds) error {
	data := struct {
		DocVersion string `json:"docVersion"`
		Resource   struct {
			Name                string        `json:"name"`
			Kind                string        `json:"kind"`
			ApiVersion          string        `json:"apiVersion"`
			ProjectKey          string        `json:"projectKey"`
			Dependencies        []interface{} `json:"dependencies"`
			RuntimeDependencies []interface{} `json:"runtimeDependencies"`
			Files               []interface{} `json:"files"`
			Folders             []struct {
				FullyQualifiedName string `json:"fullyQualifiedName"`
			} `json:"folders"`
			Spec struct {
				FileName      *string `json:"fileName"`
				FileReference *string `json:"fileReference"`
				Name          string  `json:"name"`
				Description   *string `json:"description"`
			} `json:"spec"`
			Locks []interface{} `json:"locks"`
			Key   string        `json:"key"`
		} `json:"resource"`
	}{
		DocVersion: "1.0.0",
	}
	data.Resource.Name = projectName
	data.Resource.Kind = "package"
	data.Resource.ApiVersion = "orchestrator.uipath.com/v1"
	data.Resource.ProjectKey = ids.projectKey
	data.Resource.Dependencies = []interface{}{}
	data.Resource.RuntimeDependencies = []interface{}{}
	data.Resource.Files = []interface{}{}
	data.Resource.Folders = []struct {
		FullyQualifiedName string `json:"fullyQualifiedName"`
	}{{"solution_folder"}}
	data.Resource.Spec.Name = projectName
	data.Resource.Locks = []interface{}{}
	data.Resource.Key = ids.packageKey
	return c.writeJSONFile(filepath.Join(solutionDir, "resources", "solution_folder", "package", projectName+".json"), data)
}

func (c SolutionCreateCommand) writeProcessResource(solutionDir string, solutionName string, projectName string, ids solutionIds) error {
	data := struct {
		DocVersion string `json:"docVersion"`
		Resource   struct {
			Name                string        `json:"name"`
			Kind                string        `json:"kind"`
			Type                string        `json:"type"`
			ApiVersion          string        `json:"apiVersion"`
			ProjectKey          string        `json:"projectKey"`
			Dependencies        []struct {
				Name string `json:"name"`
				Kind string `json:"kind"`
			} `json:"dependencies"`
			RuntimeDependencies []interface{} `json:"runtimeDependencies"`
			Files               []interface{} `json:"files"`
			Folders             []struct {
				FullyQualifiedName string `json:"fullyQualifiedName"`
			} `json:"folders"`
			Spec struct {
				EntryPointUniqueId   *string       `json:"entryPointUniqueId"`
				Type                 string        `json:"type"`
				Name                 string        `json:"name"`
				Description          *string       `json:"description"`
				Package              struct {
					Key string `json:"key"`
				} `json:"package"`
				PackageName          string        `json:"packageName"`
				InputArguments       string        `json:"inputArguments"`
				HiddenForAttendedUser bool         `json:"hiddenForAttendedUser"`
				AlwaysRunning        bool          `json:"alwaysRunning"`
				AutoStartProcess     bool          `json:"autoStartProcess"`
				TargetFrameworkValue string        `json:"targetFrameworkValue"`
				AgentMemory          bool          `json:"agentMemory"`
				RetentionAction      string        `json:"retentionAction"`
				RetentionPeriod      int           `json:"retentionPeriod"`
				StaleRetentionAction string        `json:"staleRetentionAction"`
				StaleRetentionPeriod int           `json:"staleRetentionPeriod"`
				Tags                 []interface{} `json:"tags"`
			} `json:"spec"`
			Locks []interface{} `json:"locks"`
			Key   string        `json:"key"`
		} `json:"resource"`
	}{
		DocVersion: "1.0.0",
	}
	data.Resource.Name = projectName
	data.Resource.Kind = "process"
	data.Resource.Type = "agent"
	data.Resource.ApiVersion = "orchestrator.uipath.com/v1"
	data.Resource.ProjectKey = ids.projectKey
	data.Resource.Dependencies = []struct {
		Name string `json:"name"`
		Kind string `json:"kind"`
	}{{projectName, "package"}}
	data.Resource.RuntimeDependencies = []interface{}{}
	data.Resource.Files = []interface{}{}
	data.Resource.Folders = []struct {
		FullyQualifiedName string `json:"fullyQualifiedName"`
	}{{"solution_folder"}}
	data.Resource.Spec.Type = "Agent"
	data.Resource.Spec.Name = projectName
	data.Resource.Spec.Package.Key = ids.packageKey
	data.Resource.Spec.PackageName = solutionName + ".agent." + projectName
	data.Resource.Spec.InputArguments = "{}"
	data.Resource.Spec.TargetFrameworkValue = "Portable"
	data.Resource.Spec.RetentionAction = "Delete"
	data.Resource.Spec.RetentionPeriod = 30
	data.Resource.Spec.StaleRetentionAction = "Delete"
	data.Resource.Spec.StaleRetentionPeriod = 180
	data.Resource.Spec.Tags = []interface{}{}
	data.Resource.Locks = []interface{}{}
	data.Resource.Key = ids.processKey
	return c.writeJSONFile(filepath.Join(solutionDir, "resources", "solution_folder", "process", "agent", projectName+".json"), data)
}

func (c SolutionCreateCommand) getStringParameter(name string, defaultValue string, parameters []plugin.ExecutionParameter) string {
	result := defaultValue
	for _, p := range parameters {
		if p.Name == name {
			if data, ok := p.Value.(string); ok {
				result = data
				break
			}
		}
	}
	return result
}

func NewSolutionCreateCommand() *SolutionCreateCommand {
	return &SolutionCreateCommand{}
}
