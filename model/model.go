package model

type IngressRules struct {
	Host  string
	Rules []IngressRule
}

type IngressRule struct {
	Path    string
	Service string
	Port    int32
}
