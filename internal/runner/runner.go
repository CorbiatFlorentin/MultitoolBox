package runner

// Runner is implemented by each language plugin. Detect reports whether dir
// looks like a project of that kind; Test runs its test command in dir,
// streaming output to stdout/stderr.
type Runner interface {
	Name() string
	Detect(dir string) bool
	Test(dir string) error
}

var registry = map[string]Runner{}

func Register(r Runner) {
	registry[r.Name()] = r
}

func Get(name string) (Runner, bool) {
	r, ok := registry[name]
	return r, ok
}

func All() map[string]Runner {
	return registry
}
