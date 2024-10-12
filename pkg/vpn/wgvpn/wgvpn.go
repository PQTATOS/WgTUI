package wgvpn

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/PQTATOS/WireTgBot/pkg/vpn"
)

type Server struct {
	configRepo ConfigRepo

    publicKey []byte
    privateKey []byte

    ipInterface string
    ipLocalAddress string
    ipOutgoingAddress string
    port uint16

    nextAllowedIp net.IP
}


type ConfigData struct {
    name string

    privateKey []byte
    publicKey []byte

    ipAddress net.IP
}


func New(configRepo ConfigRepo) (*Server, error) {
    if err := checkReqPackages(); err != nil {
        return nil, fmt.Errorf("%e", err)
    }
	return &Server{configRepo: configRepo}, nil
}


func checkReqPackages() error {
    if err := checkPackage("wg"); err != nil {
        return err
    }
    if err := checkPackage("ufw"); err != nil {
        return err
    }
    return nil
} 


func checkPackage(name string) error {
    cmd := exec.Command(name, "version")
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("%s not installed", name)
    }
    return nil
}


func (srv *Server) Init() error {
    if err := srv.configure(); err != nil {
        return fmt.Errorf("Init: %w", err)
    }
    if err := srv.configRepo.SaveServerConfig(srv); err != nil {
        return fmt.Errorf("Init: %w", err)
    }
    if err := srv.startWgService(); err != nil {
        return fmt.Errorf("Init: %w", err)
    }
	return nil
}


func (srv *Server) configure() error {
    if err := srv.assignPort(); err != nil {
        return fmt.Errorf("configure: %w", err)
    }
    if err := srv.getIpInterface(); err != nil {
        return fmt.Errorf("configure: %w", err)
    }
    if err := srv.assignIPLocalAddress(); err != nil {
        return fmt.Errorf("configure: %w", err)
    }
    var keyError error
    srv.privateKey, keyError = srv.genPrivateKey()
    if keyError != nil {
        return fmt.Errorf("configure: %w", keyError)
    }
    srv.publicKey, keyError = srv.genPublicKey(srv.privateKey)
    if keyError != nil {
        return fmt.Errorf("configure: %w", keyError)
    }
    return nil
}


func (clt *Server) assignPort() error {
    clt.port = 51830
    if err := exec.Command("ufw", "allow", strconv.Itoa(int(clt.port))).Run(); err != nil {
        return fmt.Errorf("assignPort allow: %w", err)
    }
    if err := exec.Command("ufw", "enable").Run(); err != nil {
        return fmt.Errorf("assignPort enable: %w", err)
    }
    return nil
}


func (clt *Server) assignIPLocalAddress() error {
    clt.ipLocalAddress = "10.101.0.1/24"
    clt.nextAllowedIp = net.IPv4(10,101,0,2)
    return nil
}


func (srv *Server) getIpInterface() error {
    if err := srv.getOutgoingIp(); err != nil {
        return fmt.Errorf("getIpInterface: %w", err)
    }
    inter, err := net.Interfaces()
    if err != nil {
        return fmt.Errorf("getIpInterface interfaces: %w", err)
    }
    for i := range inter {
        addrs, err := inter[i].Addrs()
        if err != nil {
            return fmt.Errorf("getIpInterface addres: %w", err)
        }
        for j := range addrs {
            if ipnet, ok := addrs[j].(*net.IPNet); ok {
                if ipnet.IP.Mask(ipnet.Mask).Equal(net.ParseIP(srv.ipOutgoingAddress).Mask(ipnet.Mask)) {
                    srv.ipInterface = inter[i].Name
                    return nil
                }
            }
        }
    }
    return fmt.Errorf("can`t find interface")
}


func (srv *Server) getOutgoingIp() error {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        return fmt.Errorf("getOutgoingIp: %w", err)
    }
    defer conn.Close()
    srv.ipOutgoingAddress = strings.Split(conn.LocalAddr().String(), ":")[0]
    return nil
} 


func (clt *Server) startWgService() error {
    file, err := os.OpenFile("/etc/sysctl.conf", os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        return fmt.Errorf("startWgService opening file: %w", err)
    }
    defer file.Close()
    file.WriteString("\nnet.ipv4.ip_forward=1\n")

    if err := exec.Command("systemctl", "enable", "wg-quick@wg0.service").Run(); err != nil {
        return fmt.Errorf("startWgService enabling: %w", err)
    }

    if err := exec.Command("systemctl", "start", "wg-quick@wg0.service").Run(); err != nil {
        return fmt.Errorf("startWgService starting: %w", err)
    }
    status, err := exec.Command("systemctl", "status", "wg-quick@wg0.service").Output()
    if  err != nil {
        return fmt.Errorf("startWgService statusing: %w", err)
    }
    fmt.Printf("%s\n", status)
    return nil
}


func (srv *Server) NewConfig(cfgName string) (*vpn.Config, error) {
	privateKey, err := srv.genPrivateKey()
	if err != nil {
		return nil, err
	}
	publicKey, err := srv.genPublicKey(privateKey)
	if err != nil {
		return nil, err
	}
    cfgData := ConfigData{
        name: cfgName,
        privateKey: privateKey,
        publicKey: publicKey,
        ipAddress: srv.nextAllowedIp,
    }

    cfg, err := srv.configRepo.SaveClientConfig(srv, &cfgData)
	if err != nil {
		return nil, err
	}
    if err := srv.configRepo.AddNewPeer(srv, &cfgData); err != nil {
		return nil, err
	}
    if err := srv.incrementIP(); err != nil {
        return nil, fmt.Errorf("NewConfig: %w", err)
    }
    if err := exec.Command("systemctl", "restart", "wg-quick@wg0.service").Run(); err != nil {
        return nil, fmt.Errorf("NewConfig restarting: %w", err)
    }
	return cfg, nil
}


func (srv *Server) incrementIP() error {
    ip := srv.nextAllowedIp
	ip[len(ip) - 1]++
	if ip[len(ip) - 1] != 0 {
        srv.nextAllowedIp = ip
		return nil
	}
	return fmt.Errorf("incrementIP: hit the limit %v", ip)
}


func (clt *Server) genPrivateKey() ([]byte, error) {
	privateKey, err := exec.Command("wg", "genkey").Output()
	if err != nil {
		return nil, fmt.Errorf("genPrivateKey executicng command: %w", err)
	}
	prvFile, err := os.Create("private_key")
	if err != nil {
		return nil, fmt.Errorf("genPrivateKey creating: %w", err)
	}
	defer prvFile.Close()
    if _, err := prvFile.Write(privateKey[:len(privateKey)-1]); err != nil {
        return nil, fmt.Errorf("genPrivateKey writeing: %w", err)
    }
	return privateKey, nil
}


func (clt *Server) genPublicKey(prvKey []byte) ([]byte, error) {
    cmd := exec.Command("wg", "pubkey")
    cmdStdin, _ := cmd.StdinPipe()
    go func() {
        defer cmdStdin.Close()
        io.WriteString(cmdStdin, string(prvKey))
    }()
    publicKey, err := cmd.CombinedOutput()
    if err != nil {
        return nil, fmt.Errorf("genPublicKey executiing: %w", err)
    }
    publicKey = publicKey[:len(publicKey)-1]
    pubFile, err := os.Create("public_key")
    if err != nil {
        return nil, fmt.Errorf("genPublicKey creating file: %w", err)
    }
    defer pubFile.Close()
    if _, err := pubFile.Write(publicKey); err != nil {
        return nil, fmt.Errorf("genPublicKey writing: %w", err)
    }
    return publicKey, nil
}