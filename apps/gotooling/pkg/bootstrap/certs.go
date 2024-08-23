package bootstrap

type CertPair struct {
	Name         string `json:"name"`
	CertPath     string `json:"cert_path"`
	CertKeyPath  string `json:"cert_key_path"`
	ConsulKVPath string `json:"consul_kv_path"`
}
