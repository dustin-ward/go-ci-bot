package tasks

import (
	"fmt"
	"log"
	"net"
	"time"
)

const (
	// MaxTasks is the maximum number of simultaneous tasks *per machine*
	MaxTasks = 1
)

var (
	zosMachines = []string{
	}
)

func getZosMachine() (string, bool) {
	for _, candidate := range zosMachines {
		// Are any other workers using this machine?
		count := 0
		for _, worker := range workerPool {
			if worker.CurMachine == candidate {
				count++ // We can probably have a hashmap for each machine to track this instead
			}
		}
		if count >= MaxTasks {
			// log.Printf("Candidate (%s) count=%d exceeds max=%d\n", candidate, count, MaxTasks)
			continue
		}

		// Is this machine accepting ssh connections?
		if ok, err := sshAvailable(candidate); ok && err == nil {
			return candidate, true
		} else if err != nil {
			log.Printf("Error testing %s for ssh connection: %v", candidate, err)
			continue
		}
	}

	return "", false
}

// Lifted from 'merge-n-test.go` in 'go-build-zos-automation'
func sshAvailable(host string) (bool, error) {
	port := "22"
	timeout := time.Second

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
	if err != nil {
		return false, err
	}

	if conn != nil {
		defer conn.Close()
		err := conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		recvBuf := make([]byte, 1024)

		_, err = conn.Read(recvBuf[:])
		if err != nil {
			return false, err
		} else {
			if string(recvBuf[0:3]) != "SSH" {
				return false, fmt.Errorf("Not SSH response")
			}
		}
	}

	return true, nil
}
