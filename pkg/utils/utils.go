package utils

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

// Verbosity levels for PrintVerbose
const (
	CRITICAL    = -1 // Always print
	DEBUG       = 3  // Print if verbose 3
	INFORMATION = 2  // Print if verbose 2
	VERBOSE     = 1  // Peint if verbose 1
)

func Check(e error, verbosity int, message ...string) {
	if e != nil {
		fmt.Println(strings.Join(message, " "))
		if verbosity >= DEBUG {
			panic(e)
		} else {
			os.Exit(-1)
		}
	}
}

func PrintVerbose(verbosity, priority int, message ...interface{}) {
	if verbosity < priority {
		return
	}

	var sb strings.Builder
	switch priority {
	case CRITICAL:
		sb.WriteString("[CRITICAL]: ")
	case DEBUG:
		sb.WriteString("[DEBUG]: ")
	case INFORMATION:
		sb.WriteString("[INFORMATION]: ")
	case VERBOSE:
		sb.WriteString("[VERBOSE]: ")
	default:
		sb.WriteString("[VERBOSE]: ")
	}
	for _, m := range message {
		sb.WriteString(fmt.Sprintf("%v", m))
	}
	fmt.Println(sb.String())
}

// GenerateRandomString generates a random string of length n
func GenerateRandomString(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = charset[random.Intn(len(charset))]
	}
	return string(b)
}

// GetIP returns the IP address for the default route in this host
func GetDefaultRouteIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

// GetInterfaceIP returns the IP address of a network interface
func GetInterfaceIP(name string) (string, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return "", err
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		switch v := addr.(type) {
		case *net.IPNet:
			if v.IP.To4() != nil {
				return v.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no IPv4 address found for interface %s", name)
}

func Min(values []int) int {
	if len(values) == 0 {
		return -1
	}
	min := values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
	}
	return min
}

func ContainsString(array []string, item string) bool {
	for _, v := range array {
		if v == item {
			return true
		}
	}
	return false
}

func RandomChoiceString(items []string) (string, int) {
	idx := rand.Intn(len(items))
	return items[idx], idx
}

func RandomChoiceInt(items []int) (int, int) {
	idx := rand.Intn(len(items))
	return items[idx], idx
}

func RandomPercentChance(percent float64) bool {
	return rand.Float64() <= percent
}
