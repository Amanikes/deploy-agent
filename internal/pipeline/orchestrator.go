package pipeline

import (
	"fmt"
	"os"
)

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
