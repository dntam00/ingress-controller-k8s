package model

type IngressRules struct {
	Host  string
	Rules []IngressRule
}

type IngressRule struct {
	Namespace string
	Path      string
	Service   string
	Port      int32
}

type IngressEvent struct {
	Type    string
	Ingress map[string]IngressRules
}
