package runner

func init() {
	Register(typescriptRunner{})
}

type typescriptRunner struct{}

func (typescriptRunner) Name() string { return "typescript" }

func (typescriptRunner) Detect(dir string) bool {
	return fileExists(dir, "tsconfig.json")
}

func (typescriptRunner) Test(dir string) error {
	if err := run(dir, "npx", "tsc", "--noEmit"); err != nil {
		return err
	}
	return run(dir, "npm", "test", "--silent")
}
