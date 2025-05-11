package traffic

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/coreos/go-iptables/iptables"
)

// UpdateIPTables updates the iptables rule to redirect a given port to a new destination port.
// It first removes any existing rules for the given `prevPort` and then adds the new rule.
func UpdateIPTables(prevPort, newPort int) error {
	ipt, err := iptables.New()
	if err != nil {
		return fmt.Errorf("failed to initialize iptables: %v", err)
	}

	// Format the port numbers
	prevPortStr := fmt.Sprintf("%d", prevPort)
	newPortStr := fmt.Sprintf("127.0.0.1:%d", newPort)

	// 🔍 Find and delete existing rules
	existingRules, err := ipt.List("nat", "OUTPUT")
	if err != nil {
		return fmt.Errorf("failed to list iptables rules: %v", err)
	}

	fmt.Println("Checking for existing rules...")

	// Regex to extract the `--to-destination` part
	r := regexp.MustCompile(`--to-destination\s+([\d\.]+:\d+)`)

	for _, rule := range existingRules {
		if strings.Contains(rule, "--dport "+prevPortStr) {
			fmt.Println("Found existing rule: ", rule)

			// Extract the `--to-destination` value
			matches := r.FindStringSubmatch(rule)
			if len(matches) > 1 {
				toDestination := matches[1]

				// 🗑️ Delete the rule properly with the full address
				err := ipt.Delete("nat", "OUTPUT", "-p", "tcp",
					"--dport", prevPortStr,
					"-j", "DNAT",
					"--to-destination", toDestination)
				if err != nil {
					fmt.Printf("Failed to remove old iptables rule: %v\n", err)
				} else {
					fmt.Printf("Removed existing rule for port %s → %s\n", prevPortStr, toDestination)
				}
			} else {
				fmt.Println("Could not parse --to-destination from rule. Skipping.")
			}
		}
	}

	// ✅ Append the new rule to the NAT table
	err = ipt.AppendUnique("nat", "OUTPUT", "-p", "tcp",
		"--dport", prevPortStr,
		"-j", "DNAT",
		"--to-destination", newPortStr)

	if err != nil {
		return fmt.Errorf("failed to update iptables: %v", err)
	}

	fmt.Printf("iptables updated: %d → %d\n", prevPort, newPort)
	return nil
}
