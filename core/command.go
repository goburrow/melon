package core

// Command is a basic CLI command
type Command interface {
	Name() string
	Description() string
	Run(bootstrap *Bootstrap) error
}
