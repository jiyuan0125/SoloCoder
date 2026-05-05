package main

import (
	"sort"
	"strings"

	"env-switcher/protocol"
)

type EnvManager struct {
	envs  *EnvsFile
	state *StateFile
}

func NewEnvManager() (*EnvManager, error) {
	envs, err := loadEnvsFile()
	if err != nil {
		return nil, err
	}
	
	state, err := loadStateFile()
	if err != nil {
		return nil, err
	}
	
	return &EnvManager{
		envs:  envs,
		state: state,
	}, nil
}

func (em *EnvManager) reload() error {
	envs, err := loadEnvsFile()
	if err != nil {
		return err
	}
	
	state, err := loadStateFile()
	if err != nil {
		return err
	}
	
	em.envs = envs
	em.state = state
	return nil
}

func (em *EnvManager) listEnvs() *protocol.Response {
	var envNames []string
	for name := range em.envs.Envs {
		envNames = append(envNames, name)
	}
	sort.Strings(envNames)
	
	return &protocol.Response{
		Success: true,
		Envs:    envNames,
		Current: em.state.Current,
	}
}

func (em *EnvManager) showEnv() *protocol.Response {
	envConfig, ok := em.envs.Envs[em.state.Current]
	if !ok {
		return &protocol.Response{
			Success: false,
			Message: "Current environment not found: " + em.state.Current,
		}
	}
	
	return &protocol.Response{
		Success: true,
		Vars:    envConfig.Vars,
		Current: em.state.Current,
	}
}

func (em *EnvManager) switchEnv(envName string) *protocol.Response {
	envConfig, ok := em.envs.Envs[envName]
	if !ok {
		var availableEnvs []string
		for name := range em.envs.Envs {
			availableEnvs = append(availableEnvs, name)
		}
		sort.Strings(availableEnvs)
		
		return &protocol.Response{
			Success: false,
			Message: "Environment not found: " + envName + "\nAvailable environments: " + strings.Join(availableEnvs, ", "),
			Envs:    availableEnvs,
		}
	}
	
	var warnings []string
	if len(envConfig.Vars) == 0 {
		warnings = append(warnings, "Environment '"+envName+"' has no variables defined")
	}
	
	em.state.Current = envName
	if err := saveStateFile(em.state); err != nil {
		return &protocol.Response{
			Success: false,
			Message: "Failed to save state: " + err.Error(),
		}
	}
	
	return &protocol.Response{
		Success:  true,
		Message:  "Switched to environment: " + envName,
		Warnings: warnings,
	}
}

func (em *EnvManager) exportEnv() *protocol.Response {
	envConfig, ok := em.envs.Envs[em.state.Current]
	if !ok {
		return &protocol.Response{
			Success: false,
			Message: "Current environment not found: " + em.state.Current,
		}
	}
	
	return &protocol.Response{
		Success: true,
		Vars:    envConfig.Vars,
		Current: em.state.Current,
	}
}

func (em *EnvManager) addEnv(envName string, vars map[string]string) *protocol.Response {
	if _, exists := em.envs.Envs[envName]; exists {
		return &protocol.Response{
			Success: false,
			Message: "Environment '" + envName + "' already exists",
		}
	}
	
	if vars == nil {
		vars = make(map[string]string)
	}
	
	em.envs.Envs[envName] = EnvConfig{Vars: vars}
	
	if err := saveEnvsFile(em.envs); err != nil {
		return &protocol.Response{
			Success: false,
			Message: "Failed to save environments: " + err.Error(),
		}
	}
	
	return &protocol.Response{
		Success: true,
		Message: "Added environment: " + envName,
	}
}

func (em *EnvManager) delEnv(envName string) *protocol.Response {
	if envName == em.state.Current {
		return &protocol.Response{
			Success: false,
			Message: "Cannot delete current active environment: " + envName,
		}
	}
	
	if _, exists := em.envs.Envs[envName]; !exists {
		return &protocol.Response{
			Success: false,
			Message: "Environment '" + envName + "' does not exist",
		}
	}
	
	delete(em.envs.Envs, envName)
	
	if err := saveEnvsFile(em.envs); err != nil {
		return &protocol.Response{
			Success: false,
			Message: "Failed to save environments: " + err.Error(),
		}
	}
	
	return &protocol.Response{
		Success: true,
		Message: "Deleted environment: " + envName,
	}
}

func (em *EnvManager) HandleRequest(req *protocol.Request) *protocol.Response {
	switch req.Cmd {
	case protocol.CmdList:
		return em.listEnvs()
	case protocol.CmdShow:
		return em.showEnv()
	case protocol.CmdSwitch:
		return em.switchEnv(req.Env)
	case protocol.CmdExport:
		return em.exportEnv()
	case protocol.CmdAdd:
		return em.addEnv(req.Env, req.Vars)
	case protocol.CmdDel:
		return em.delEnv(req.Env)
	default:
		return &protocol.Response{
			Success: false,
			Message: "Unknown command: " + string(req.Cmd),
		}
	}
}
