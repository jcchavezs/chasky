package types

type Output struct {
	EnvVars []string
	Closer  func() error
}
