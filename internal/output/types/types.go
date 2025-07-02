package types

type Output struct {
	WelcomeMsg string
	EnvVars    []string
	Closer     func() error
}
