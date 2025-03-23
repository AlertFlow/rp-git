package main

import (
	"errors"
	"net/rpc"
	"os"
	"time"

	"github.com/v1Flows/runner/pkg/executions"
	"github.com/v1Flows/runner/pkg/plugins"
	"github.com/v1Flows/shared-library/pkg/models"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/go-plugin"
)

// Plugin is an implementation of the Plugin interface
type Plugin struct{}

func (p *Plugin) ExecuteTask(request plugins.ExecuteTaskRequest) (plugins.Response, error) {
	url := ""
	directory := ""
	username := ""
	password := ""
	token := ""

	// access action params
	for _, param := range request.Step.Action.Params {
		if param.Key == "url" {
			url = param.Value
		}
		if param.Key == "directory" {
			directory = request.Workspace + "/" + param.Value
		}
		if param.Key == "username" {
			username = param.Value
		}
		if param.Key == "password" {
			password = param.Value
		}
		if param.Key == "token" {
			token = param.Value
		}
	}

	// update the step with the messages
	err := executions.UpdateStep(request.Config, request.Execution.ID.String(), models.ExecutionSteps{
		ID: request.Step.ID,
		Messages: []models.Message{
			{
				Title: "Git",
				Lines: []models.Line{
					{
						Content: "Cloning repository " + url + " to " + directory,
					},
				},
			},
		},
		Status:    "running",
		StartedAt: time.Now(),
	}, request.Platform)
	if err != nil {
		return plugins.Response{
			Success: false,
		}, err
	}

	if token != "" {
		_, err = git.PlainClone(directory, false, &git.CloneOptions{
			Auth: &http.BasicAuth{
				Username: "abc123",
				Password: token,
			},
			URL:      url,
			Progress: os.Stdout,
		})
		if err != nil {
			err := executions.UpdateStep(request.Config, request.Execution.ID.String(), models.ExecutionSteps{
				ID: request.Step.ID,
				Messages: []models.Message{
					{
						Title: "Git",
						Lines: []models.Line{
							{
								Content: "Error cloning repository",
								Color:   "danger",
							},
							{
								Content: err.Error(),
								Color:   "danger",
							},
						},
					},
				},
				Status:     "error",
				FinishedAt: time.Now(),
			}, request.Platform)
			if err != nil {
				return plugins.Response{
					Success: false,
				}, err
			}
			return plugins.Response{
				Success: false,
			}, err
		}
	} else {
		_, err = git.PlainClone(directory, false, &git.CloneOptions{
			Auth: &http.BasicAuth{
				Username: username,
				Password: password,
			},
			URL:      url,
			Progress: os.Stdout,
		})
		if err != nil {
			err := executions.UpdateStep(request.Config, request.Execution.ID.String(), models.ExecutionSteps{
				ID: request.Step.ID,
				Messages: []models.Message{
					{
						Title: "Git",
						Lines: []models.Line{
							{
								Content: "Error cloning repository",
								Color:   "danger",
							},
							{
								Content: err.Error(),
								Color:   "danger",
							},
						},
					},
				},
				Status:     "error",
				FinishedAt: time.Now(),
			}, request.Platform)
			if err != nil {
				return plugins.Response{
					Success: false,
				}, err
			}
			return plugins.Response{
				Success: false,
			}, err
		}
	}

	err = executions.UpdateStep(request.Config, request.Execution.ID.String(), models.ExecutionSteps{
		ID: request.Step.ID,
		Messages: []models.Message{
			{
				Title: "Git",
				Lines: []models.Line{
					{
						Content: "Repository cloned successfully",
						Color:   "success",
					},
				},
			},
		},
		Status:     "success",
		FinishedAt: time.Now(),
	}, request.Platform)
	if err != nil {
		return plugins.Response{
			Success: false,
		}, err
	}

	return plugins.Response{
		Success: true,
	}, nil
}

func (p *Plugin) EndpointRequest(request plugins.EndpointRequest) (plugins.Response, error) {
	return plugins.Response{
		Success: false,
	}, errors.New("not implemented")
}

func (p *Plugin) Info(request plugins.InfoRequest) (models.Plugin, error) {
	var plugin = models.Plugin{
		Name:    "Git",
		Type:    "action",
		Version: "1.0.0",
		Author:  "JustNZ",
		Action: models.Action{
			Name:        "Git",
			Description: "Clone a repository",
			Plugin:      "git",
			Icon:        "mdi:git",
			Category:    "Utility",
			Params: []models.Params{
				{
					Key:         "url",
					Title:       "URL",
					Type:        "text",
					Default:     "",
					Required:    true,
					Description: "URL of the repository to clone",
					Category:    "Repository",
				},
				{
					Key:         "directory",
					Title:       "Directory",
					Type:        "text",
					Default:     "",
					Required:    true,
					Description: "Path to clone the repository to. The path prefix is the workspace directory: " + request.Workspace,
					Category:    "Repository",
				},
				{
					Key:         "username",
					Title:       "Username",
					Type:        "text",
					Default:     "",
					Required:    false,
					Description: "Username for authentication",
					Category:    "Authentication",
				},
				{
					Key:         "password",
					Title:       "Password",
					Type:        "password",
					Default:     "",
					Required:    false,
					Description: "Password for authentication",
					Category:    "Authentication",
				},
				{
					Key:         "token",
					Title:       "Token",
					Type:        "password",
					Default:     "",
					Required:    false,
					Description: "Token for authentication. If provided, username and password will be ignored",
					Category:    "Authentication",
				},
			},
		},
		Endpoint: models.Endpoint{},
	}

	return plugin, nil
}

// PluginRPCServer is the RPC server for Plugin
type PluginRPCServer struct {
	Impl plugins.Plugin
}

func (s *PluginRPCServer) ExecuteTask(request plugins.ExecuteTaskRequest, resp *plugins.Response) error {
	result, err := s.Impl.ExecuteTask(request)
	*resp = result
	return err
}

func (s *PluginRPCServer) EndpointRequest(request plugins.EndpointRequest, resp *plugins.Response) error {
	result, err := s.Impl.EndpointRequest(request)
	*resp = result
	return err
}

func (s *PluginRPCServer) Info(request plugins.InfoRequest, resp *models.Plugin) error {
	result, err := s.Impl.Info(request)
	*resp = result
	return err
}

// PluginServer is the implementation of plugin.Plugin interface
type PluginServer struct {
	Impl plugins.Plugin
}

func (p *PluginServer) Server(*plugin.MuxBroker) (interface{}, error) {
	return &PluginRPCServer{Impl: p.Impl}, nil
}

func (p *PluginServer) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &plugins.PluginRPC{Client: c}, nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "PLUGIN_MAGIC_COOKIE",
			MagicCookieValue: "hello",
		},
		Plugins: map[string]plugin.Plugin{
			"plugin": &PluginServer{Impl: &Plugin{}},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
