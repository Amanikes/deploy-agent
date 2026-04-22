package pipeline

import (
	"auto-deploy-agent/internal/logger"
	"os/exec"
)

type Pipeline struct {
	WorkDir string
}

func NewPipeline(workDir string) *Pipeline {
	return &Pipeline{
		WorkDir: workDir,
	}
}

func (p *Pipeline) RunStep(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = p.WorkDir

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return err
	}

	//Straem logs in the background so we dont block

	go logger.Stream(stdout, "STDOUT")
	go logger.Stream(stderr, "STDERR")

	return cmd.Wait()

}
