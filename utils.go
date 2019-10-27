
package main

import (
	"net"
	"net/http"
	"io"
	"os"
	"errors"
	"bytes"
)

// downloadFile will download a url to a local file
func downloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// externalIP gets the IPv4 of the system
func externalIP() (string, error) {
	ifaces, err := net.Interfaces()

	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {

		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}

		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			var ip net.IP

			switch v := addr.(type) {

			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}

			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}

// macUint64 gets the local MAC address as a uint64
func macUint64() (uint64, error) {
    interfaces, err := net.Interfaces()
    if err != nil {
        return uint64(0), err
    }

    for _, i := range interfaces {
        if i.Flags&net.FlagUp != 0 && bytes.Compare(i.HardwareAddr, nil) != 0 {

            // Skip locally administered addresses
            if i.HardwareAddr[0]&2 == 2 {
                continue
            }

            var mac uint64
            for j, b := range i.HardwareAddr {
                if j >= 8 {
                    break
                }
                mac <<= 8
                mac += uint64(b)
            }

            return mac, nil
        }
    }

    return uint64(0), errors.New("couldn't get MAC address")
}

func getCampus() (string, error) {
	return "", nil
}

func getBuilding() (string, error) {
	return "", nil
}

func getRoom() (string, error) {
	return "", nil
}