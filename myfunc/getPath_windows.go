//go:build windows

package myfunc

import (
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// GetInstalledResourcesByName lấy danh sách ứng dụng đã cài đặt từ Windows Registry
func GetInstalledResourcesByName() (map[string]string, error) {
	resources := make(map[string]string)
	registryKeys := []string{
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`,
		`SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`,
	}

	for _, regKey := range registryKeys {
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, regKey, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)
		if err != nil {
			continue
		}
		defer k.Close()

		subKeys, err := k.ReadSubKeyNames(0)
		if err != nil {
			continue
		}

		for _, subKey := range subKeys {
			sub, err := registry.OpenKey(k, subKey, registry.QUERY_VALUE)
			if err != nil {
				continue
			}

			displayName, _, _ := sub.GetStringValue("DisplayName")
			exePath, _, _ := sub.GetStringValue("DisplayIcon")
			if exePath == "" {
				exePath, _, _ = sub.GetStringValue("InstallLocation")
				if exePath != "" {
					exePath = filepath.Join(exePath, subKey+".exe")
				}
			}

			if exePath != "" && displayName != "" {
				exePath = strings.Split(strings.Trim(exePath, `"`), ",")[0]
				resources[strings.ToLower(displayName)] = exePath
			}
			sub.Close()
		}
	}
	return resources, nil
}
