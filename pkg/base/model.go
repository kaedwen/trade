package base

type Quote struct {
	Symbol string
	Value  float64
}

type Data struct {
	Name   string
	Tags   map[string]string
	Fields map[string]interface{}
}
