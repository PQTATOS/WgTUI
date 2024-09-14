package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
)

func main() {
    if err := InstallWireguard(); err == nil {
        fmt.Printf("Wireguard installed\n")
    } else {
        fmt.Printf("%s\n", err)
        return
    }
    if err := CreateWgConfig(); err == nil {

    } else {

    }
}

func CreateWgConfig() error {
    workingDir, _ := os.Getwd()
    if err := os.Chdir("/etc/wireguard/"); err != nil {
        fmt.Errorf("Cant move to /etc/wireguard/: %s\n", err)
    }
    privateKey, err := genPrivateKey()
    if err != nil {
        return fmt.Errorf("Creting config: %w\n", err)
    }
    _, err = genPublicKey(privateKey)
    if err != nil {
        return fmt.Errorf("Creting config: %w\n", err)
    }

    fmt.Println(workingDir)
    return nil
}

func createConfigFile(prvKey []byte) error {
    configFile, err := os.Create("wg0.conf")
    if err != nil {
        return fmt.Errorf("Error with creating file: %w\n", err)
    }
    configFile.WriteString("[Interface]\n")
    configFile.WriteString(fmt.Sprintf("PostUp = iptables -A FORWARD -i %i -j ACCEPT; iptables -t nat -A POSTROUTING -o %s -j MASQUERADE"))
    configFile.WriteString(fmt.Sprintf("PostDown = iptables -D FORWARD -i %i -j ACCEPT; iptables -t nat -D POSTROUTING -o %s -j MASQUERADE\n",))

    return nil
}

func getOutboundInterface()

func genPrivateKey() ([]byte, error) {
    privateKey, err := exec.Command("wg", "genkey").Output()
    if err != nil {
        return nil, fmt.Errorf("Error with generating a private key: %w\n", err)
    }
    prvFile, err := os.Create("private_key")
    if err != nil {
        return nil, fmt.Errorf("Error with creating file: %w\n", err)
    }
    defer prvFile.Close()
    prvFile.Write(privateKey)
    return privateKey, nil
}

func genPublicKey(prvKey []byte) ([]byte, error) {
    cmd := exec.Command("wg", "genkey")
    cmdStdin, _ := cmd.StdinPipe()
    io.WriteString(cmdStdin, string(prvKey))
    publicKey, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("Error with generating a public key: %w\n", err)
    }
    pubFile, err := os.Create("public_key")
    if err != nil {
        return nil, fmt.Errorf("Error with creating file: %w\n", err)
    }
    defer pubFile.Close()
    pubFile.Write(publicKey)
    return publicKey, nil
}

func getIp() (string, error) {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        return "", err
    }
    defer conn.Close()
    localAddr := conn.LocalAddr().(*net.UDPAddr)
    return localAddr.IP.String(), nil
} 

func InstallWireguard() error {
	cmd := exec.Command("wg", "version")
    cmd.Stderr = os.Stdout
    err := cmd.Run()
    if err == nil {
        return fmt.Errorf("Wireguard alredy install\n")
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
        fmt.Errorf("Wireguard fail to install\n")
    }
    return nil
}