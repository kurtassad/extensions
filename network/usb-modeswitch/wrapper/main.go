package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	modeswitchBin = "/usr/local/bin/usb_modeswitch"
	configDir     = "/etc/usb_modeswitch.d"
	sysUSBDevices = "/sys/bus/usb/devices"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("usb-modeswitch: ")

	configs, err := loadConfigs(configDir)
	if err != nil {
		log.Fatalf("read configs: %v", err)
	}

	devices, err := listUSBDevices(sysUSBDevices)
	if err != nil {
		log.Fatalf("scan usb devices: %v", err)
	}

	var matched, failed int
	for id, configPath := range configs {
		devs, ok := devices[id]
		if !ok {
			continue
		}

		vendor, product, _ := strings.Cut(id, ":")
		for _, dev := range devs {
			matched++
			log.Printf("switching %s (bus %s dev %s) using %s", id, dev.bus, dev.dev, configPath)

			cmd := exec.Command(
				modeswitchBin,
				"-v", vendor,
				"-p", product,
				"-b", dev.bus,
				"-g", dev.dev,
				"-c", configPath,
			)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				log.Printf("switch failed for %s: %v", id, err)
				failed++
			}
		}
	}

	if failed > 0 {
		os.Exit(1)
	}

	if matched == 0 {
		log.Printf("no storage-mode devices present; nothing to do")
		return
	}

	log.Printf("switched %d device(s)", matched)
}

func loadConfigs(dir string) (map[string]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	configs := make(map[string]string)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.Count(name, ":") != 1 {
			continue
		}
		configs[strings.ToLower(name)] = filepath.Join(dir, name)
	}

	return configs, nil
}

type usbDev struct {
	bus string
	dev string
}

func listUSBDevices(dir string) (map[string][]usbDev, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	devices := make(map[string][]usbDev)
	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		vendor, err := readTrimmed(filepath.Join(path, "idVendor"))
		if err != nil {
			continue
		}
		product, err := readTrimmed(filepath.Join(path, "idProduct"))
		if err != nil {
			continue
		}
		bus, err := readTrimmed(filepath.Join(path, "busnum"))
		if err != nil {
			continue
		}
		dev, err := readTrimmed(filepath.Join(path, "devnum"))
		if err != nil {
			continue
		}

		id := fmt.Sprintf("%s:%s", vendor, product)
		devices[id] = append(devices[id], usbDev{bus: bus, dev: dev})
	}

	return devices, nil
}

func readTrimmed(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.ToLower(strings.TrimSpace(string(data))), nil
}
