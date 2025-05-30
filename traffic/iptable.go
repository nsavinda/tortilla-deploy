package traffic

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/coreos/go-iptables/iptables"
)

func UpdateIPTables(sourcePort, newPort, prvPort int) error {
	ipt, err := iptables.New()
	if err != nil {
		return fmt.Errorf("failed to initialize iptables: %v", err)
	}

	sourcePortStr := fmt.Sprintf("%d", sourcePort)
	newPortStr := fmt.Sprintf("127.0.0.1:%d", newPort)
	prvPortStr := fmt.Sprintf("127.0.0.1:%d", prvPort)

	// List all OUTPUT nat rules
	existingRules, err := ipt.List("nat", "OUTPUT")
	if err != nil {
		return fmt.Errorf("failed to list iptables rules: %v", err)
	}

	// Add the new rule
	err = ipt.AppendUnique("nat", "OUTPUT", "-p", "tcp",
		"--dport", sourcePortStr,
		"-j", "DNAT",
		"--to-destination", newPortStr)

	if err != nil {
		return fmt.Errorf("failed to append iptables rule: %v", err)
	}

	fmt.Println("Checking for existing rules...")

	// Regex to extract the --to-destination
	r := regexp.MustCompile(`--to-destination\s+([\d\.]+:\d+)`)

	for _, rule := range existingRules {
		if strings.Contains(rule, "--dport "+sourcePortStr) {
			fmt.Println("Found existing rule:", rule)

			matches := r.FindStringSubmatch(rule)
			if len(matches) > 1 {
				// toDestination := matches[1]
				if matches[1] == prvPortStr {

					// Construct the same rule string to delete it
					err := ipt.Delete("nat", "OUTPUT", "-p", "tcp",
						"--dport", sourcePortStr,
						"-j", "DNAT",
						"--to-destination", prvPortStr)
					if err != nil {
						fmt.Printf("Failed to delete rule for port %s → %s: %v\n", prvPortStr, newPortStr, err)
					} else {
						fmt.Printf("Deleted existing rule for port %s → %s\n", prvPortStr, newPortStr)
					}
				}
			} else {
				fmt.Println("Could not extract --to-destination from rule:", rule)
			}
		}
	}

	fmt.Printf("iptables updated: %d → %d\n", sourcePort, newPort)
	return nil
}
