package vpn

type VPNServer interface {
	NewConfig(cfgName string) (*Config, error)
}

type Config struct {
	FilePath string
	Data []byte
}
