package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
)

var pubKey []byte

func main() {
    if err := InstallWireguard(); err != nil {
        fmt.Printf("%s\n", err)
        //return
    }
    fmt.Printf("Wireguard installed\n")

    if err := CreateWgConfig(); err != nil {
        fmt.Printf("%s\n", err)
        return
    }
    fmt.Printf("Config crated\n")

    if err := StartWgService(); err != nil {
        fmt.Printf("%s\n", err)
        return
    }
    fmt.Printf("Service started\n")

    if err := makeClientConfig(); err != nil {
        fmt.Printf("%s\n", err)
        return
    }
    if err := exec.Command("systemctl", "restart", "wg-quick@wg0.service").Run(); err != nil {
        fmt.Printf("%s\n", err)
        return
    }
}


func makeClientConfig() error {
    privateKey, err := genPrivateKey()
    if err != nil {
        return err
    }
    pubKeyClt, err := genPublicKey(privateKey)
    if err != nil {
        return err
    }

    if err := addNewPeer(pubKeyClt); err != nil {
        return err
    }

    if err := createClientConfigFile(privateKey); err != nil {
        return err
    }
    return nil
}


func addNewPeer(pubKeyClt []byte) error {
    file, err := os.OpenFile("/etc/wireguard/wg0.conf", os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer file.Close()
    file.WriteString("\n[Peer]\n")
    file.WriteString(fmt.Sprintf("PublicKey = %s\n", string(pubKeyClt)))
    file.WriteString("AllowedIPs = 10.101.0.2\n")
    file.Sync()
    return nil
}

func createClientConfigFile(prvKey []byte) error {
    configFile, err := os.Create("test.conf")
    if err != nil {
        return fmt.Errorf("error with creating file: %w", err)
    }
    defer configFile.Close()

    configFile.WriteString("[Interface]\n")
    configFile.WriteString(fmt.Sprintf("PrivateKey = %s\n", string(prvKey)))
    configFile.WriteString("Address = 10.101.0.2/24\n")
    configFile.WriteString("DNS = 8.8.8.8\n")

    ip, err := getIp()
    if err != nil {
        return  err
    }
    configFile.WriteString("\n[Peer]\n")
    configFile.WriteString(fmt.Sprintf("PublicKey = %s\n", string(pubKey)))
    configFile.WriteString(fmt.Sprintf("Endpoint = %s:51830\n", ip))
    configFile.WriteString("AllowedIPs = 0.0.0.0/0\n")
    configFile.WriteString("PersistentKeepalive = 20\n")

    configFile.Sync()

    return nil
}


func StartWgService() error {
    file, err := os.OpenFile("/etc/sysctl.conf", os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer file.Close()
    file.WriteString("\nnet.ipv4.ip_forward=1\n")

    if err := exec.Command("systemctl", "enable", "wg-quick@wg0.service").Run(); err != nil {
        return err
    }

    if err := exec.Command("systemctl", "start", "wg-quick@wg0.service").Run(); err != nil {
        return err
    }
    status, err := exec.Command("systemctl", "start", "wg-quick@wg0.service").Output()
    if  err != nil {
        return err
    }
    fmt.Printf("%s\n", status)
    return nil
}


func CreateWgConfig() error {
    workingDir, _ := os.Getwd()
    if err := os.Chdir("/etc/wireguard/"); err != nil {
        return fmt.Errorf("cant move to /etc/wireguard/: %w", err)
    }
    privateKey, err := genPrivateKey()
    if err != nil {
        return fmt.Errorf("creting config: %w", err)
    }
    pubKey, err = genPublicKey(privateKey)
    if err != nil {
        return fmt.Errorf("creting config: %w", err)
    }
    if err := createConfigFile(privateKey); err != nil {
        return err
    }
    if err := os.Chdir(workingDir); err != nil {
        return fmt.Errorf("cant move to %s: %w", workingDir, err)
    }
    return nil
}

func createConfigFile(prvKey []byte) error {
    configFile, err := os.Create("wg0.conf")
    if err != nil {
        return fmt.Errorf("error with creating file: %w", err)
    }
    defer configFile.Close()

    configFile.WriteString("[Interface]\n")
    configFile.WriteString(fmt.Sprintf("PrivateKey = %s\n", string(prvKey)))
    configFile.WriteString("Address = 10.101.0.1/24\n")
    if err := exec.Command("ufw", "allow", "51830").Run(); err != nil {
        return err
    }
    configFile.WriteString("ListenPort = 51830\n")
    eth, err := getIpInterface()
    if err != nil {
        return  err
    }
    configFile.WriteString(fmt.Sprintf("PostUp = iptables -A FORWARD -i %si -j ACCEPT; iptables -t nat -A POSTROUTING -o %s -j MASQUERADE\n", "%", eth))
    configFile.WriteString(fmt.Sprintf("PostDown = iptables -D FORWARD -i %si -j ACCEPT; iptables -t nat -D POSTROUTING -o %s -j MASQUERADE\n", "%", eth))
    configFile.Sync()
    return nil
}

func genPrivateKey() ([]byte, error) {
    privateKey, err := exec.Command("wg", "genkey").Output()
    if err != nil {
        return nil, fmt.Errorf("error with generating a private key: %w", err)
    }
    prvFile, err := os.Create("private_key")
    if err != nil {
        return nil, fmt.Errorf("error with creating file: %w", err)
    }
    defer prvFile.Close()
    prvFile.Write(privateKey[:len(privateKey)-1])
    return privateKey, nil
}

func genPublicKey(prvKey []byte) ([]byte, error) {
    cmd := exec.Command("wg", "pubkey")
    cmdStdin, _ := cmd.StdinPipe()
    go func() {
        defer cmdStdin.Close()
        io.WriteString(cmdStdin, string(prvKey))
    }()
    publicKey, err := cmd.CombinedOutput()
    publicKey = publicKey[:len(publicKey)-1]
    if err != nil {
        return nil, fmt.Errorf("error with generating a public key: %w", err)
    }
    pubFile, err := os.Create("public_key")
    if err != nil {
        return nil, fmt.Errorf("error with creating file: %w", err)
    }
    defer pubFile.Close()
    pubFile.Write(publicKey)
    return publicKey, nil
}


func getIpInterface() (string, error) {
    outIP, err := getIp()
    if err != nil {
        return "", err
    }
    inter, err := net.Interfaces()
    if err != nil {
        return "", err
    }
    for i := range inter {
        addrs, err := inter[i].Addrs()
        if err != nil {
            return "", err
        }
        for j := range addrs {
            if ipnet, ok := addrs[j].(*net.IPNet); ok {
                if ipnet.IP.Mask(ipnet.Mask).Equal(net.ParseIP(outIP).Mask(ipnet.Mask)) {
                    return inter[i].Name, nil
                }
            }
        }
    }
    return "", fmt.Errorf("can`t find interface")
}


func getIp() (string, error) {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        return "", err
    }
    defer conn.Close()
    return strings.Split(conn.LocalAddr().String(), ":")[0], nil
} 

func InstallWireguard() error {
	cmd := exec.Command("wg", "version")
    cmd.Stderr = os.Stdout
    err := cmd.Run()
    if err == nil {
        return fmt.Errorf("wireguard alredy install")
    }
    cmd = exec.Command("apt", "install", "wireguard")
    var stdErrBuf bytes.Buffer
    cmd.Stdout = os.Stdout
    cmd.Stderr = io.MultiWriter(os.Stderr, &stdErrBuf)
    cmdStdin, _ := cmd.StdinPipe()
    go func() {
        defer cmdStdin.Close()
        buf := bufio.NewReader(os.Stdin)
        str, _ := buf.ReadString('\n')
        io.WriteString(cmdStdin, str)
    } ()
    err = cmd.Start()
    if err != nil {
        fmt.Printf("Cant execute command: %s\n", err)
    }
    if err := cmd.Wait(); err != nil {
        if exiterr, ok := err.(*exec.ExitError); ok {
            fmt.Printf("Exit Status: %d\n", exiterr.ExitCode())
        } else {
            fmt.Printf("cmd.Wait: %v\n", err)
        }
    }
    if len(stdErrBuf.Bytes()) != 0 {
        return fmt.Errorf("wireguard fail to install")
    }
    return nil
}