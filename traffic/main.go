package main

import (
	"fmt"
	"log"

	"github.com/cilium/ebpf"
	"golang.org/x/sys/unix"
)

const iface = "eth0"

// Define the port redirection map
const keySize = uint32(2)   // size of the source port (uint16)
const valueSize = uint32(2) // size of the destination port (uint16)
const maxEntries = 2        // max number of entries in the map

func main() {
	// Load the BPF object file
	obj, err := ebpf.LoadCollection("xdp_port_redirect_with_map.o")
	if err != nil {
		log.Fatalf("Failed to load BPF object: %v", err)
	}
	defer obj.Close()

	// Access the XDP program and attach it
	prog, exists := obj.Programs["xdp"]
	if !exists {
		log.Fatalf("XDP program not found")
	}

	// Open the network interface (e.g., eth0) to attach the program
	fd, err := unix.Socket(unix.AF_PACKET, unix.SOCK_RAW, unix.ETH_P_IP) // ETH_P_IP is 0x0800
	if err != nil {
		log.Fatalf("Failed to open socket: %v", err)
	}
	defer unix.Close(fd)

	// Attach the XDP program to the interface
	// Attach the XDP program to the interface
	if err := unix.SetsockoptInt(fd, unix.SOL_XDP, unix.XDP_FLAGS_UPDATE_IF_NOEXIST, prog.FD()); err != nil {
		log.Fatalf("Failed to attach XDP program: %v", err)
	}

	// Create the port_map BPF map
	// Create the port_map BPF map
	// print maxEntries
	fmt.Println("Max entries:", maxEntries)
	portMapSpec := &ebpf.MapSpec{
		Type:       ebpf.Hash,  // This should match BPF_MAP_TYPE_HASH from C code
		KeySize:    keySize,    // 2 bytes for source port (uint16)
		ValueSize:  valueSize,  // 2 bytes for destination port (uint16)
		MaxEntries: maxEntries, // Ensure this is set correctly
	}

	fmt.Printf("Map spec: %+v\n", portMapSpec)

	portMap, err := ebpf.NewMap(portMapSpec)
	if err != nil {
		log.Fatalf("Failed to create BPF map: %v", err)
	}

	// Insert a port mapping into the map: source port 8080 -> destination port 9090
	srcPort := uint16(8080)
	dstPort := uint16(9090)
	if err := portMap.Update(&srcPort, &dstPort, ebpf.UpdateAny); err != nil {
		log.Fatalf("Failed to insert entry into port_map: %v", err)
	}

	// Print confirmation
	fmt.Printf("Successfully added port mapping: %d -> %d\n", srcPort, dstPort)

	// The program will keep running to process packets; you can implement further logic here.
	select {}
}
