package runner

func init() {
	Register(phpRunner{})
}

type phpRunner struct{}

func (phpRunner) Name() string { return "php" }

func (phpRunner) Detect(dir string) bool {
	return fileExists(dir, "composer.json") || fileExists(dir, "phpunit.xml") || fileExists(dir, "phpunit.xml.dist")
}

func (phpRunner) Test(dir string) error {
	if fileExists(dir, "vendor/bin/phpunit") || fileExists(dir, "vendor\\bin\\phpunit") {
		return run(dir, "vendor/bin/phpunit")
	}
	return run(dir, "composer", "test")
}
