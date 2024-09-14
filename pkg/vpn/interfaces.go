package vpn

type VPNClient interface {
	NewConfig(cfgName string) (Config, error)
}

type Config struct {
	Meta interface{}
}
