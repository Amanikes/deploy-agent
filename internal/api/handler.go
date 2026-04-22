package api

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func (c *AgentClient) DispatchCommand(cmd Command) {
	action := normalizeAction(cmd.Action)
	cmd.Action = action

	log.Printf("Dispatching command [%s] action=%s project=%s", cmd.ID, action, cmd.Project)
	_ = c.SendAck(NewAck(cmd, "received", "command received", nil))
	_ = c.SendAck(NewAck(cmd, "started", "worker started", nil))

	var err error
	switch action {
	case "deploy":
		err = c.handleDeploy(cmd)
	case "restart":
		err = c.handleRestart(cmd)
	case "status":
		err = c.handleStatus(cmd)
	default:
		err = fmt.Errorf("unsupported action: %s", action)
	}

	if err != nil {
		log.Printf("Command [%s] failed: %v", cmd.ID, err)
		_ = c.SendAck(NewAck(cmd, "failed", err.Error(), nil))
		return
	}

	_ = c.SendAck(NewAck(cmd, "completed", "command completed", nil))
}

func (c *AgentClient) handleDeploy(cmd Command) error {
	repoDir := cmd.Payload["repo_dir"]

	if repoDir != "" {
		_ = c.SendAck(NewAck(cmd, "progress", "running git pull", map[string]string{"step": "git_pull"}))
		if output, err := runCommand("", "git", "-C", repoDir, "pull", "--ff-only"); err != nil {
			return fmt.Errorf("git pull failed: %w (%s)", err, output)
		}
	}

	if deployCmd := cmd.Payload["deploy_cmd"]; deployCmd != "" {
		_ = c.SendAck(NewAck(cmd, "progress", "running deploy command", map[string]string{"step": "deploy_cmd"}))
		if output, err := runShell(deployCmd, repoDir); err != nil {
			return fmt.Errorf("deploy command failed: %w (%s)", err, output)
		}
		return nil
	}

	composeFile := cmd.Payload["compose_file"]
	service := cmd.Payload["service"]
	if composeFile != "" {
		_ = c.SendAck(NewAck(cmd, "progress", "running docker compose up", map[string]string{"step": "compose_up"}))
		args := []string{"compose", "-f", composeFile, "up", "-d"}
		if service != "" {
			args = append(args, service)
		}
		if output, err := runCommand(repoDir, "docker", args...); err != nil {
			return fmt.Errorf("docker compose up failed: %w (%s)", err, output)
		}
		return nil
	}

	return fmt.Errorf("deploy requires repo_dir+deploy_cmd, or compose_file")
}

func (c *AgentClient) handleRestart(cmd Command) error {
	repoDir := cmd.Payload["repo_dir"]

	if restartCmd := cmd.Payload["restart_cmd"]; restartCmd != "" {
		_ = c.SendAck(NewAck(cmd, "progress", "running restart command", map[string]string{"step": "restart_cmd"}))
		if output, err := runShell(restartCmd, repoDir); err != nil {
			return fmt.Errorf("restart command failed: %w (%s)", err, output)
		}
		return nil
	}

	if container := cmd.Payload["container"]; container != "" {
		_ = c.SendAck(NewAck(cmd, "progress", "restarting container", map[string]string{"step": "docker_restart"}))
		if output, err := runCommand("", "docker", "restart", container); err != nil {
			return fmt.Errorf("docker restart failed: %w (%s)", err, output)
		}
		return nil
	}

	composeFile := cmd.Payload["compose_file"]
	service := cmd.Payload["service"]
	if composeFile != "" && service != "" {
		_ = c.SendAck(NewAck(cmd, "progress", "restarting compose service", map[string]string{"step": "compose_restart"}))
		if output, err := runCommand(repoDir, "docker", "compose", "-f", composeFile, "restart", service); err != nil {
			return fmt.Errorf("docker compose restart failed: %w (%s)", err, output)
		}
		return nil
	}

	return fmt.Errorf("restart requires restart_cmd, container, or compose_file+service")
}

func (c *AgentClient) handleStatus(cmd Command) error {
	repoDir := cmd.Payload["repo_dir"]

	if statusCmd := cmd.Payload["status_cmd"]; statusCmd != "" {
		_ = c.SendAck(NewAck(cmd, "progress", "running status command", map[string]string{"step": "status_cmd"}))
		output, err := runShell(statusCmd, repoDir)
		if err != nil {
			return fmt.Errorf("status command failed: %w (%s)", err, output)
		}
		_ = c.SendAck(NewAck(cmd, "progress", "status collected", map[string]string{"output": trimOutput(output)}))
		return nil
	}

	if container := cmd.Payload["container"]; container != "" {
		_ = c.SendAck(NewAck(cmd, "progress", "reading container status", map[string]string{"step": "docker_ps"}))
		output, err := runCommand("", "docker", "ps", "--filter", "name="+container, "--format", "{{.Names}} {{.Status}}")
		if err != nil {
			return fmt.Errorf("docker ps failed: %w (%s)", err, output)
		}
		_ = c.SendAck(NewAck(cmd, "progress", "status collected", map[string]string{"output": trimOutput(output)}))
		return nil
	}

	return fmt.Errorf("status requires status_cmd or container")
}

func runShell(command string, workDir string) (string, error) {
	return runCommand(workDir, "sh", "-c", command)
}

func runCommand(workDir string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	if workDir != "" {
		cmd.Dir = workDir
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	return out.String(), err
}

func trimOutput(output string) string {
	trimmed := strings.TrimSpace(output)
	if len(trimmed) > 500 {
		return trimmed[:500]
	}
	return trimmed
}
