package port

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/samber/lo"
)

// 获取可用端口
func GetAvailablePort(t string) (int, error) {
	address := fmt.Sprintf("%s:0", "0.0.0.0")
	if t == "udp" {
		add, err := net.ResolveUDPAddr(t, address)
		if err != nil {
			return 0, err
		}

		listener, err := net.ListenUDP(t, add)
		if err != nil {
			return 0, err
		}

		defer listener.Close()
		return listener.LocalAddr().(*net.UDPAddr).Port, nil
	}

	add, err := net.ResolveTCPAddr(t, address)
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP(t, add)
	if err != nil {
		return 0, err
	}

	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

// 判断端口是否可以（未被占用）
// param t tcp/udp
func IsPortAvailable(port int, t string) bool {
	address := fmt.Sprintf("%s:%d", "0.0.0.0", port)

	switch t {
	case "tcp":
		listener, err := net.Listen(t, address)
		if err != nil {
			// log.Infof("port %s is taken: %s", address, err)
			return false
		}
		defer listener.Close()

	case "udp":
		sadd, err := net.ResolveUDPAddr(t, address)
		if err != nil {
			fmt.Println(err.Error())
			return false
		}

		uc, err := net.ListenUDP(t, sadd)
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
		defer uc.Close()

	default:
		return false
	}

	return true
}

func ListPortsInUse() ([]int, []int, error) {
	usedPorts := map[string]map[int]struct{}{
		"tcp":  {},
		"udp":  {},
		"tcp6": {},
		"udp6": {},
	}

	for _, protocol := range lo.Keys(usedPorts) {
		filename := fmt.Sprintf("/proc/net/%s", protocol)

		file, err := os.Open(filename)
		if err != nil {
			return nil, nil, errors.New("Failed to open " + filename)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}

			localAddress := fields[1]
			addressParts := strings.Split(localAddress, ":")
			if len(addressParts) < 2 {
				continue
			}

			portHex := addressParts[1]
			port, err := strconv.ParseInt(portHex, 16, 0)
			if err != nil {
				continue
			}

			usedPorts[protocol][int(port)] = struct{}{}
		}

		if err := scanner.Err(); err != nil {
			return nil, nil, errors.New("Error reading from " + filename)
		}
	}

	return lo.Union(
			lo.Keys(usedPorts["tcp"]), lo.Keys(usedPorts["tcp6"]),
		),
		lo.Union(
			lo.Keys(usedPorts["udp"]), lo.Keys(usedPorts["udp6"]),
		),
		nil
}
