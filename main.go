package main

import (
	"cloudflare-ip-updater/services"
	"flag"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	zoneIdentifier string
	filter         string
	authKey        string
	previousIP     string
	ipLock         sync.Mutex
)

func main() {
	// Define flags with custom error messages
	flag.StringVar(&zoneIdentifier, "zone_identifier", "", "Cloudflare Zone Identifier (required)")
	flag.StringVar(&filter, "filter", "", "Cloudflare A record filter (required)")
	flag.StringVar(&authKey, "auth_key", "", "Cloudflare Authentication Key (required)")

	// Parse the command line arguments
	flag.Parse()

	// Check if required flags are provided
	if err := checkRequiredFlags(); err != nil {
		fmt.Println("Error:", err)
		flag.PrintDefaults()
		return
	}

	cloudflareService := services.NewCloudflareService(zoneIdentifier, authKey, filter)

	// Set up a Goroutine to run every 2 minutes
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for ; true; <-ticker.C {
		go func() {
			fmt.Println("Stating IP Update Process")
			// Get the current IP address
			currentIP, err := services.GetIPAddress()
			if err != nil {
				fmt.Println("Error getting IP address:", err)
				return
			}

			// Compare the current IP with the previous IP
			ipLock.Lock()
			if currentIP != previousIP {
				fmt.Printf("IP has changed. Updating Cloudflare DNS records, ip: %s\n", currentIP)

				records, err := cloudflareService.GetDnsRecords()
				if err != nil {
					fmt.Println("Error Getting A Records:", err)
					return
				}

				err = cloudflareService.UpdateDnsRecord(records, currentIP)
				if err != nil {
					fmt.Println("Error Updating A Records:", err)
					return
				}
				// Update the previous IP
				previousIP = currentIP
			}
			ipLock.Unlock()
		}()
	}

	// Run indefinitely
	select {}
}

func checkRequiredFlags() error {
	var missingFlags []string

	if zoneIdentifier == "" {
		missingFlags = append(missingFlags, "-zone_identifier")
	}
	if filter == "" {
		missingFlags = append(missingFlags, "-filter")
	}
	if authKey == "" {
		missingFlags = append(missingFlags, "-auth_key")
	}

	if len(missingFlags) > 0 {
		return fmt.Errorf("missing or empty required flags: %s", strings.Join(missingFlags, ", "))
	}

	return nil
}
