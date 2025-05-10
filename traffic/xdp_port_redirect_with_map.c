#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/tcp.h>
#include <linux/if_packet.h>
#include <linux/in.h>

#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

// Define the BPF map in a modern style

// struct bpf_map_def SEC("maps") port_map = {
//     .type = BPF_MAP_TYPE_HASH,
//     .key_size = sizeof(__u16),   // Source port (TCP)
//     .value_size = sizeof(__u16), // Destination port (TCP)
//     .max_entries = 2,
// };

struct  {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(key_size, sizeof(__u16));   // Source port (TCP)
    __uint(value_size, sizeof(__u16)); // Destination port (TCP)
    __uint(max_entries, 2);
} port_map SEC(".maps");


// Helper to get pointer to Ethernet header
static __always_inline struct ethhdr *parse_eth(struct __sk_buff *skb) {
    return (struct ethhdr *)(long)skb->data;
}

// Helper to get pointer to IP header
static __always_inline struct iphdr *parse_ip(struct __sk_buff *skb) {
    struct ethhdr *eth = parse_eth(skb);
    return (struct iphdr *)(eth + 1);
}

// Helper to get pointer to TCP header
static __always_inline struct tcphdr *parse_tcp(struct __sk_buff *skb) {
    struct iphdr *ip = parse_ip(skb);
    return (struct tcphdr *)(ip + 1);
}

SEC("xdp")
int redirect_ports(struct __sk_buff *skb) {
    struct ethhdr *eth = parse_eth(skb);
    struct iphdr *ip = parse_ip(skb);

    // Ensure this is an IPv4 packet
    if (ip->version != 4) {
        return XDP_PASS;
    }

    // Ensure this is a TCP packet
    if (ip->protocol != IPPROTO_TCP) {
        return XDP_PASS;
    }

    struct tcphdr *tcp = parse_tcp(skb);

    // Get the source port (TCP)
    __u16 src_port = bpf_ntohs(tcp->source);

    bpf_printk("Looking up port: %d\n", src_port);

    // Look up the destination port for redirection
    __u16 *dst_port = bpf_map_lookup_elem(&port_map, &src_port);
    if (!dst_port) {
        return XDP_PASS;
    }

    // Modify the TCP packet destination port
    tcp->dest = bpf_htons(*dst_port);

    // Return XDP_TX to send the modified packet
    return XDP_TX;
}

char _license[] SEC("license") = "GPL";
