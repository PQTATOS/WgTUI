package testvpn

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/PQTATOS/WireTgBot/pkg/vpn"
)

type TestVpn struct {
	configsDir string
	filenameSep string
}


func New(configDir string) *TestVpn {
	if err := os.Mkdir(configDir, 0777); err != nil && !os.IsExist(err) {
		panic(err)
	}
	slog.Info("Creating vpnClient")
	return &TestVpn{configsDir: configDir, filenameSep: "-"}
}

func (vpnClient *TestVpn) NewConfig(cfgName string) (vpn.Config, error) {
	slog.Info("Creating new vpn config")
	file, fileErr := vpnClient.newFile(cfgName)
	if fileErr != nil {
		panic(fileErr)
	}
	defer file.Close()
	file.WriteString("Hello")

	return vpn.Config{Meta: file.Name()}, nil
}

func (vpnClient *TestVpn) newFile(cfgName string) (*os.File, error) {
	filename, nameErr := vpnClient.newFilename(cfgName)
	if nameErr != nil {
		return nil, fmt.Errorf("error in newFile with new filename: %w", nameErr)
	}
	filepath := fmt.Sprintf("%s\\%s", vpnClient.configsDir, filename)
	file, fileErr := os.Create(filepath)
	if fileErr != nil {
		panic(fileErr)
	}
	return file, nil
}

func (vpnClient *TestVpn) newFilename(cfgName string) (string, error) {
	files, err := os.ReadDir(vpnClient.configsDir)
	if err != nil {
		return "", fmt.Errorf("error with in newFilename while reading dirs: %w", err)
	}
	slog.Debug(fmt.Sprintf("newFilename: files %v", files))
	index, inRange := findFile(files, fmt.Sprintf("%s%s99.conf",cfgName, vpnClient.filenameSep), vpnClient.filenameSep)
	if !inRange {
		return fmt.Sprintf("%s%s01.conf", cfgName, vpnClient.filenameSep), nil
	}

	fileWithoutType := strings.Split(files[index].Name(), ".")
	nameAndID := strings.Split(fileWithoutType[0], vpnClient.filenameSep)
	slog.Debug(fmt.Sprintf("newFilename: found file name and id %v", nameAndID))
	if nameAndID[0] != cfgName {
		return fmt.Sprintf("%s%s01.conf", cfgName, vpnClient.filenameSep), nil
	}

	id, parseErr := strconv.ParseUint(nameAndID[1], 10, 32)
	if parseErr != nil {
		return "", fmt.Errorf("error with in newFilename while parsing: %w", parseErr)
	}
	id++

	return fmt.Sprintf("%s%s%02d.conf", cfgName, vpnClient.filenameSep, id), nil
}


func findFile(files []fs.DirEntry, key, sep string) (index int, inRange bool) {
	left := -1
	right := len(files)
	for left + 1 < right {
		mid := (left + right) / 2
		if comparePrefix(files[mid].Name(), key, sep) > 0 {
			right = mid
		} else {
			left = mid
		}
		slog.Debug(fmt.Sprintf("left=%d right=%d", left, right))
	}
	return left, left > -1 && left < len(files)
}

func comparePrefix(a, b, sep string) (out int) {
	defer func() { slog.Debug(fmt.Sprintf("%s  <%d>  %s", a, out, b)) } ()
	minlen := min(len(a), len(b))
	for i := 0; i < minlen; i++ {
		if a[i] < b[i] {
			return -1
		} else if a[i] > b[i] {
			return 1
		}
		if a[i] == sep[0] {
			break
		}
	}
	return 0
}