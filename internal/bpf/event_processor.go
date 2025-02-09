package bpf

import (
	"log/slog"

	"github.com/nullswan/bpfsnitch/internal/metrics"
	"github.com/nullswan/bpfsnitch/pkg/network"
)

func ProcessNetworkEvent(
	event *NetworkEvent,
	container string,
	log *slog.Logger,
) {
	// Adjust endianness if necessary
	event.Saddr = network.Ntohl(event.Saddr)
	event.Daddr = network.Ntohl(event.Daddr)
	event.Sport = network.Ntohs(event.Sport)
	event.Dport = network.Ntohs(event.Dport)

	// Convert IP addresses to net.IP
	saddr := network.IntToIP(event.Saddr)
	daddr := network.IntToIP(event.Daddr)

	log.With("pid", event.Pid).
		With("cgroup_id", event.CgroupID).
		With("container", container).
		With("saddr", saddr).
		With("daddr", daddr).
		With("sport", event.Sport).
		With("dport", event.Dport).
		With("size", event.Size).
		Debug("Received network event")

	if event.Protocol == 17 && event.Direction == 0 && event.Dport == 53 {
		metrics.DNSQueryCounter.WithLabelValues(container).Inc()
	}

	daddrStr := daddr.String()
	if event.Direction == 0 {
		metrics.NetworkSentBytesCounter.WithLabelValues(container, daddrStr).
			Add(float64(event.Size))
		metrics.NetworkSentPacketsCounter.WithLabelValues(container, daddrStr).
			Inc()
	} else {
		metrics.NetworkReceivedBytesCounter.WithLabelValues(container, daddrStr).Add(float64(event.Size))
		metrics.NetworkReceivedPacketsCounter.WithLabelValues(container, daddrStr).Inc()
	}
}

func ProcessSyscallEvent(
	event *SyscallEvent,
	container string,
	log *slog.Logger,
) {
	log.With("syscall", event.GetSyscallName()).
		With("pid", event.Pid).
		With("cgroup_id", event.CgroupID).
		With("container", container).
		Debug("Received syscall event")

	metrics.SyscallCounter.WithLabelValues(event.GetSyscallName(), container).
		Inc()
}
