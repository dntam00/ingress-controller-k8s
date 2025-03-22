package model

import "crypto/tls"

type IngressRules struct {
	Host  string
	Cert  *tls.Certificate
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
	Ingress map[string]*IngressRules
}
