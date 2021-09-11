package template

type Task interface {
	Run()
	BindParameters(map[string]interface{})
	GetName() string
}
