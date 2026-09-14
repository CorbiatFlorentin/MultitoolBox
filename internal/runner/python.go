package runner

func init() {
	Register(pythonRunner{})
}

type pythonRunner struct{}

func (pythonRunner) Name() string { return "python" }

func (pythonRunner) Detect(dir string) bool {
	return fileExists(dir, "pyproject.toml") || fileExists(dir, "requirements.txt") || fileExists(dir, "setup.py")
}

func (pythonRunner) Test(dir string) error {
	return run(dir, "python", "-m", "pytest")
}
