package application

// ports
type LineSource interface {
	Load() ([]string, error)
}

// use cases
// .....
