package wgvpn

import (
	"bytes"
	"fmt"
	"os"

	"github.com/PQTATOS/WireTgBot/pkg/vpn"
)


type ConfigRepo interface {
	SaveServerConfig(srv *Server) error
	AddNewPeer(srv *Server, cfgData *ConfigData) error
	SaveClientConfig(srv *Server, cfgData *ConfigData) (*vpn.Config, error)
}


type ConfigLocalRepo struct {
	directory string
}


func CreateConfigLocalRepo(path string) *ConfigLocalRepo {
	return &ConfigLocalRepo{directory: path}
}


func (repo *ConfigLocalRepo) SaveClientConfig(srv *Server, cfgData *ConfigData) (*vpn.Config, error) {
	configFile, err := os.Create(fmt.Sprintf("%s.conf", cfgData.name))
    if err != nil {
        return nil, fmt.Errorf("SaveClientConfig: %w", err)
    }
    defer configFile.Close()

	fileData := bytes.Buffer{}
	fileData.Grow(266)

    fileData.WriteString("[Interface]\n")
    fileData.WriteString(fmt.Sprintf("PrivateKey = %s", cfgData.privateKey))
    fileData.WriteString(fmt.Sprintf("Address = %s/32\n", srv.nextAllowedIp))
    fileData.WriteString("DNS = 8.8.8.8\n")
    fileData.WriteString("\n[Peer]\n")
    fileData.WriteString(fmt.Sprintf("PublicKey = %s\n", srv.publicKey))
    fileData.WriteString(fmt.Sprintf("Endpoint = %s:%d\n", srv.ipOutgoingAddress, srv.port))
    fileData.WriteString("AllowedIPs = 0.0.0.0/0\n")
    fileData.WriteString("PersistentKeepalive = 20\n")

	configFile.Write(fileData.Bytes())
    configFile.Sync()

    return &vpn.Config{FilePath: fmt.Sprintf("%s.conf", cfgData.name), Data: fileData.Bytes()}, nil
}


func (repo *ConfigLocalRepo) AddNewPeer(srv *Server, cfgData *ConfigData) error {
    file, err := os.OpenFile(fmt.Sprintf("%s/wg0.conf", repo.directory), os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        return fmt.Errorf("AddNewPeer: %w", err)
    }
    defer file.Close()
    file.WriteString("\n[Peer]\n")
    file.WriteString(fmt.Sprintf("PublicKey = %s\n", string(cfgData.publicKey)))
    file.WriteString(fmt.Sprintf("AllowedIPs = %s\n", cfgData.ipAddress))
    file.Sync()
    return nil
}


func (repo *ConfigLocalRepo) SaveServerConfig(srv *Server) error {
	if err := repo.saveKey("server_private", srv.privateKey); err != nil {
		return fmt.Errorf("SaveServerConfig error with saving private key: %w", err)
	}
	if err := repo.saveKey("server_public", srv.publicKey); err != nil {
		return fmt.Errorf("SaveServerConfig error with saving public key: %w", err)
	}
	configFile, err := os.Create(fmt.Sprintf("%s/wg0.conf", repo.directory))
	if err != nil {
		return fmt.Errorf("SaveServerConfig error with creating file: %w", err)
	}
	defer configFile.Close()

	configFile.WriteString("[Interface]\n")
	configFile.WriteString(fmt.Sprintf("PrivateKey = %s", srv.privateKey))
	configFile.WriteString(fmt.Sprintf("Address = %s\n", srv.ipLocalAddress))
	configFile.WriteString(fmt.Sprintf("ListenPort = %d\n", srv.port))
	configFile.WriteString(fmt.Sprintf("PostUp = iptables -A FORWARD -i %si -j ACCEPT; iptables -t nat -A POSTROUTING -o %s -j MASQUERADE\n", "%", srv.ipInterface))
	configFile.WriteString(fmt.Sprintf("PostDown = iptables -D FORWARD -i %si -j ACCEPT; iptables -t nat -D POSTROUTING -o %s -j MASQUERADE\n", "%", srv.ipInterface))
	configFile.Sync()
	return nil
}

func (repo *ConfigLocalRepo) saveKey(path string, key []byte) error {
	file, err := os.Create(fmt.Sprintf("%s/%s", repo.directory, path))
    if err != nil {
        return fmt.Errorf("saveKey error with creating file: %w", err)
    }
    defer file.Close()
    file.Write(key)
    return nil
}