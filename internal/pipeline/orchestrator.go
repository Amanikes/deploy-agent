package pipeline

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type ProgressUpdate struct {
	Step    string
	Message string
	Meta    map[string]string
}

type ProgressFunc func(update ProgressUpdate)

// Context-aware version of Execute
func ExecuteWithContext(ctx context.Context, action string, payload map[string]string, onProgress ProgressFunc) (map[string]string, error) {
	notify := func(step string, message string, meta map[string]string) {
		if onProgress != nil {
			onProgress(ProgressUpdate{Step: step, Message: message, Meta: meta})
		}
	}

	switch action {
	case "deploy":
		return executeDeployWithContext(ctx, payload, notify)
	case "restart":
		return executeRestartWithContext(ctx, payload, notify)
	case "status":
		return executeStatusWithContext(ctx, payload, notify)
	default:
		return nil, fmt.Errorf("unsupported action: %s", action)
	}
}

// Context-aware deploy
func executeDeployWithContext(ctx context.Context, payload map[string]string, notify func(step string, message string, meta map[string]string)) (map[string]string, error) {
	workDir := strings.TrimSpace(payload["repo_dir"])
	repoURL := strings.TrimSpace(payload["repo_url"])
	cleanup := func() {}

	if workDir == "" && repoURL != "" {
		notify("clone_repo", "cloning repository", map[string]string{"repo_url": repoURL})
		clonedDir, closeFn, err := cloneRepo(repoURL)
		if err != nil {
			return nil, err
		}
		workDir = clonedDir
		cleanup = closeFn
	}
	defer cleanup()

	if workDir != "" && repoURL == "" {
		notify("git_pull", "running git pull", nil)
		if output, err := runCommandContext(ctx, "", "git", "-C", workDir, "pull", "--ff-only"); err != nil {
			return nil, fmt.Errorf("git pull failed: %w (%s)", err, output)
		}
	}

	testCmd := strings.TrimSpace(payload["test_cmd"])
	if testCmd == "" && strings.EqualFold(strings.TrimSpace(payload["run_tests"]), "true") {
		testCmd = "go test ./..."
	}

	if testCmd != "" {
		notify("run_tests", "running test command", map[string]string{"test_cmd": testCmd})
		if output, err := runShellContext(ctx, testCmd, workDir); err != nil {
			return nil, fmt.Errorf("test command failed: %w (%s)", err, output)
		}
	}

	if deployCmd := strings.TrimSpace(payload["deploy_cmd"]); deployCmd != "" {
		notify("deploy_cmd", "running deploy command", map[string]string{"deploy_cmd": deployCmd})
		if output, err := runShellContext(ctx, deployCmd, workDir); err != nil {
			return nil, fmt.Errorf("deploy command failed: %w (%s)", err, output)
		}
		return nil, nil
	}

	composeFile := strings.TrimSpace(payload["compose_file"])
	service := strings.TrimSpace(payload["service"])
	if composeFile != "" {
		notify("compose_up", "running docker compose up", nil)
		args := []string{"compose", "-f", composeFile, "up", "-d"}
		if service != "" {
			args = append(args, service)
		}
		if output, err := runCommandContext(ctx, workDir, "docker", args...); err != nil {
			return nil, fmt.Errorf("docker compose up failed: %w (%s)", err, output)
		}
		return nil, nil
	}

	return nil, fmt.Errorf("deploy requires deploy_cmd or compose_file")
}

// Context-aware restart
func executeRestartWithContext(ctx context.Context, payload map[string]string, notify func(step string, message string, meta map[string]string)) (map[string]string, error) {
	repoDir := strings.TrimSpace(payload["repo_dir"])

	if restartCmd := strings.TrimSpace(payload["restart_cmd"]); restartCmd != "" {
		notify("restart_cmd", "running restart command", nil)
		if output, err := runShellContext(ctx, restartCmd, repoDir); err != nil {
			return nil, fmt.Errorf("restart command failed: %w (%s)", err, output)
		}
		return nil, nil
	}

	if container := strings.TrimSpace(payload["container"]); container != "" {
		notify("docker_restart", "restarting container", map[string]string{"container": container})
		if output, err := runCommandContext(ctx, "", "docker", "restart", container); err != nil {
			return nil, fmt.Errorf("docker restart failed: %w (%s)", err, output)
		}
		return nil, nil
	}

	composeFile := strings.TrimSpace(payload["compose_file"])
	service := strings.TrimSpace(payload["service"])
	if composeFile != "" && service != "" {
		notify("compose_restart", "restarting compose service", map[string]string{"service": service})
		if output, err := runCommandContext(ctx, repoDir, "docker", "compose", "-f", composeFile, "restart", service); err != nil {
			return nil, fmt.Errorf("docker compose restart failed: %w (%s)", err, output)
		}
		return nil, nil
	}

	return nil, fmt.Errorf("restart requires restart_cmd, container, or compose_file+service")
}

// Context-aware status
func executeStatusWithContext(ctx context.Context, payload map[string]string, notify func(step string, message string, meta map[string]string)) (map[string]string, error) {
	repoDir := strings.TrimSpace(payload["repo_dir"])

	if statusCmd := strings.TrimSpace(payload["status_cmd"]); statusCmd != "" {
		notify("status_cmd", "running status command", nil)
		output, err := runShellContext(ctx, statusCmd, repoDir)
		if err != nil {
			return nil, fmt.Errorf("status command failed: %w (%s)", err, output)
		}
		return map[string]string{"output": trimOutput(output)}, nil
	}

	if container := strings.TrimSpace(payload["container"]); container != "" {
		notify("docker_ps", "reading container status", map[string]string{"container": container})
		output, err := runCommandContext(ctx, "", "docker", "ps", "--filter", "name="+container, "--format", "{{.Names}} {{.Status}}")
		if err != nil {
			return nil, fmt.Errorf("docker ps failed: %w (%s)", err, output)
		}
		return map[string]string{"output": trimOutput(output)}, nil
	}

	return nil, fmt.Errorf("status requires status_cmd or container")
}

// Context-aware shell/command runners
func runShellContext(ctx context.Context, command string, workDir string) (string, error) {
	return runCommandContext(ctx, workDir, "sh", "-c", command)
}

func runCommandContext(ctx context.Context, workDir string, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if workDir != "" {
		cmd.Dir = workDir
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	return out.String(), err
}

// Backward-compatible: create a background context and call context-aware version
func Execute(action string, payload map[string]string, onProgress ProgressFunc) (map[string]string, error) {
	return ExecuteWithContext(context.Background(), action, payload, onProgress)
}

func ExecuteBuild(repoURL string, imageName string) error {
	//Create a temporary directory for the build
	tmpDir, _ := os.MkdirTemp("", "build-*")
	defer os.RemoveAll(tmpDir)

	p := NewPipeline(tmpDir)

	//Clone the repo

	fmt.Println("---- Step 1: Cloning Repository ----")
	if err := p.RunStep("git", "clone", repoURL, "."); err != nil {
		return fmt.Errorf("failed to clone repo: %w", err)
	}

	//Build the Docker image
	fmt.Println("---- Step 2: Building Docker Image ----")
	if err := p.RunStep("docker", "build", "-t", imageName, "."); err != nil {
		return fmt.Errorf("failed to build docker image: %w", err)
	}

	fmt.Println("---- Build completed successfully! ----")
	return nil

}

func cloneRepo(repoURL string) (string, func(), error) {
	tmpDir, err := os.MkdirTemp("", "deploy-agent-repo-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	if output, err := runCommand("", "git", "clone", repoURL, tmpDir); err != nil {
		_ = os.RemoveAll(tmpDir)
		return "", nil, fmt.Errorf("git clone failed: %w (%s)", err, output)
	}

	cleanup := func() {
		_ = os.RemoveAll(tmpDir)
	}
	return tmpDir, cleanup, nil
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
