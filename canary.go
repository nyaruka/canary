package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"
)

// TunnelTest is our config struct.. we expect to receive a file that looks something like:
// [
//   { "host": "192.168.0.1:900", "tunnel": "vpn-to-foo" }
// ]
type TunnelTest struct {
	Host   string `json:"host"`
	Tunnel string `json:"tunnel"`
}

func main() {
	log.SetFlags(log.Ldate | log.Lmicroseconds)

	if len(os.Args) != 2 {
		log.Fatalf("usage: canary <canary file.json>")
	}

	file, e := ioutil.ReadFile(os.Args[1])
	if e != nil {
		log.Fatalf("error reading file: %v\n", e)
	}

	tunnels := make([]TunnelTest, 0)
	err := json.Unmarshal(file, &tunnels)
	if err != nil {
		log.Fatalf("error reading file: %v\n", e)
	}

	pid := os.Getpid()
	pidFileName := "/run/canary.pid"
	_, err = os.Stat(pidFileName)
	if !os.IsNotExist(err) {

		pidFile, e := ioutil.ReadFile(pidFileName)
		if e != nil {
			log.Fatalf("error reading PID file: %v\n", e)
		}

		if string(pidFile) != strconv.Itoa(pid) {
			cmd := exec.Command("kill", "-9", string(pidFile))
			stdoutStderr, err := cmd.CombinedOutput()
			log.Printf("%s\n", stdoutStderr)
			if err == nil {
				log.Printf("killed previous process with PID %s", string(pidFile))
			}

		}
	}

	err = ioutil.WriteFile(pidFileName, []byte(strconv.Itoa(pid)), 0644)

	for _, tunnel := range tunnels {
		conn, err := net.DialTimeout("tcp", tunnel.Host, time.Second*10)

		if err != nil {
			log.Printf("!! %s not healthy, bouncing\n", tunnel.Tunnel)
			log.Printf("%s", err)

			cmd := exec.Command("/usr/sbin/ipsec", "auto", "--down", tunnel.Tunnel)
			stdoutStderr, err := cmd.CombinedOutput()
			log.Printf("%s\n", stdoutStderr)

			cmd = exec.Command("/usr/sbin/ipsec", "auto", "--up", tunnel.Tunnel)
			stdoutStderr, err = cmd.CombinedOutput()
			if err != nil {
				log.Printf("!! error bringing tunnel %s up: %v\n", tunnel.Tunnel, err)
			}
			log.Printf("%s\n", stdoutStderr)
			log.Printf("%s restarted\n", tunnel.Tunnel)
		} else {
			err = conn.Close()
			if err != nil {
				log.Printf("!! error closing connection to %s: %v\n", tunnel.Host, err)
			}
			log.Printf("%s healthy\n", tunnel.Tunnel)
		}
	}
}
