package template

type Task interface {
	Run() error
	BindParameters(map[string]interface{})
	GetName() string
}
