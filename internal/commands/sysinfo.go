package commands

import (
	"fmt"
	"os"
)

// os.Hostname()
func handleHostname(args []string, ctx *Context) bool {
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println("Error getting hostname:", err)
		return false
	}

	fmt.Println("Hostname:", hostname)
	return true
}

// os.UserHomeDir()
func handleHomedir(args []string, ctx *Context) bool {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		return false
	}

	fmt.Println("Home:", home)
	return true
}

// os.UserCacheDir()
func handleCachedir(args []string, ctx *Context) bool {
	cache, err := os.UserCacheDir()
	if err != nil {
		fmt.Println("Error getting cache directory:", err)
		return false
	}

	fmt.Println("Cache:", cache)
	return true
}

// os.UserConfigDir()
func handleConfigdir(args []string, ctx *Context) bool {
	config, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("Error getting config directory:", err)
		return false
	}

	fmt.Println("Config:", config)
	return true
}

// show all system info at once
func handleSysinfo(args []string, ctx *Context) bool {
	hostname, _ := os.Hostname()
	home, _ := os.UserHomeDir()
	cache, _ := os.UserCacheDir()
	config, _ := os.UserConfigDir()

	fmt.Println("Hostname :", hostname)
	fmt.Println("Home     :", home)
	fmt.Println("Cache    :", cache)
	fmt.Println("Config   :", config)
	return true
}