// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build linux_bpf

package connection

import (
	"io"
	"unsafe"

	"github.com/cilium/ebpf"
	"github.com/davecgh/go-spew/spew"

	manager "github.com/DataDog/ebpf-manager"

	ddebpf "github.com/DataDog/datadog-agent/pkg/network/ebpf"
	"github.com/DataDog/datadog-agent/pkg/network/ebpf/probes"
	"github.com/DataDog/datadog-agent/pkg/network/tracer/offsetguess"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

func dumpMapsHandler(w io.Writer, _ *manager.Manager, mapName string, currentMap *ebpf.Map) {
	switch mapName {

	case "connectsock_ipv6": // maps/connectsock_ipv6 (BPF_MAP_TYPE_HASH), key C.__u64, value uintptr // C.void*
		io.WriteString(w, "Map: '"+mapName+"', key: 'C.__u64', value: 'uintptr // C.void*'\n")
		iter := currentMap.Iterate()
		var key uint64
		var value uintptr // C.void*
		for iter.Next(unsafe.Pointer(&key), unsafe.Pointer(&value)) {
			spew.Fdump(w, key, value)
		}

	case probes.TracerStatusMap: // maps/tracer_status (BPF_MAP_TYPE_HASH), key C.__u64, value tracerStatus
		io.WriteString(w, "Map: '"+mapName+"', key: 'C.__u64', value: 'tracerStatus'\n")
		iter := currentMap.Iterate()
		var key uint64
		var value offsetguess.TracerStatus
		for iter.Next(unsafe.Pointer(&key), unsafe.Pointer(&value)) {
			spew.Fdump(w, key, value)
		}

	case probes.ConntrackStatusMap: // maps/conntrack_status (BPF_MAP_TYPE_HASH), key C.__u64, value conntrackStatus
		io.WriteString(w, "Map: '"+mapName+"', key: 'C.__u64', value: 'conntrackStatus'\n")
		iter := currentMap.Iterate()
		var key uint64
		var value offsetguess.ConntrackStatus
		for iter.Next(unsafe.Pointer(&key), unsafe.Pointer(&value)) {
			spew.Fdump(w, key, value)
		}

	case probes.ConntrackMap: // maps/conntrack (BPF_MAP_TYPE_HASH), key ConnTuple, value ConnTuple
		io.WriteString(w, "Map: '"+mapName+"', key: 'ConnTuple', value: 'ConnTuple'\n")
		iter := currentMap.Iterate()
		var key ddebpf.ConnTuple
		var value ddebpf.ConnTuple
		for iter.Next(unsafe.Pointer(&key), unsafe.Pointer(&value)) {
			spew.Fdump(w, key, value)
		}

	case probes.ConntrackTelemetryMap: // maps/conntrack_telemetry (BPF_MAP_TYPE_ARRAY), key C.u32, value conntrackTelemetry
		io.WriteString(w, "Map: '"+mapName+"', key: 'C.u32', value: 'conntrackTelemetry'\n")
		var zero uint64
		telemetry := &ddebpf.ConntrackTelemetry{}
		if err := currentMap.Lookup(unsafe.Pointer(&zero), unsafe.Pointer(telemetry)); err != nil {
			log.Tracef("error retrieving the contrack telemetry struct: %s", err)
		}
		spew.Fdump(w, telemetry)

	case probes.ConnMap: // maps/conn_stats (BPF_MAP_TYPE_HASH), key ConnTuple, value ConnStatsWithTimestamp
		io.WriteString(w, "Map: '"+mapName+"', key: 'ConnTuple', value: 'ConnStatsWithTimestamp'\n")
		iter := currentMap.Iterate()
		var key ddebpf.ConnTuple
		var value ddebpf.ConnStats
		for iter.Next(unsafe.Pointer(&key), unsafe.Pointer(&value)) {
			spew.Fdump(w, key, value)
		}

	case probes.TCPStatsMap: // maps/tcp_stats (BPF_MAP_TYPE_HASH), key ConnTuple, value TCPStats
		io.WriteString(w, "Map: '"+mapName+"', key: 'ConnTuple', value: 'TCPStats'\n")
		iter := currentMap.Iterate()
		var key ddebpf.ConnTuple
		var value ddebpf.TCPStats
		for iter.Next(unsafe.Pointer(&key), unsafe.Pointer(&value)) {
			spew.Fdump(w, key, value)
		}

	case probes.TCPOngoingConnectPid: // maps/tcp_ongoing_connect_pid (BPF_MAP_TYPE_HASH), key SkpConnTuple, value u64
		io.WriteString(w, "Map: '"+mapName+"', key: 'SkpConnTuple', value: 'C.u64'\n")
		io.WriteString(w, "This map is used to store the PID of the process that initiated the connection\n")
		totalSize := 0
		info, _ := currentMap.Info()
		spew.Fdump(w, info)
		iter := currentMap.Iterate()
		var key ddebpf.SkpConn
		var value ddebpf.PidTs
		for iter.Next(unsafe.Pointer(&key), unsafe.Pointer(&value)) {
			totalSize++
			spew.Fdump(w, key.Tup, value)
		}
		io.WriteString(w, "Total entries: "+spew.Sdump(totalSize))

	case probes.ConnCloseBatchMap: // maps/conn_close_batch (BPF_MAP_TYPE_HASH), key C.__u32, value batch
		io.WriteString(w, "Map: '"+mapName+"', key: 'C.__u32', value: 'batch'\n")
		iter := currentMap.Iterate()
		var key uint32
		var value ddebpf.Batch
		for iter.Next(unsafe.Pointer(&key), unsafe.Pointer(&value)) {
			spew.Fdump(w, key, value)
		}

	case "udp_recv_sock": // maps/udp_recv_sock (BPF_MAP_TYPE_HASH), key C.__u64, value C.udp_recv_sock_t
		io.WriteString(w, "Map: '"+mapName+"', key: 'C.__u64', value: 'C.udp_recv_sock_t'\n")
		iter := currentMap.Iterate()
		var key uint64
		var value ddebpf.UDPRecvSock
		for iter.Next(unsafe.Pointer(&key), unsafe.Pointer(&value)) {
			spew.Fdump(w, key, value)
		}

	case "udpv6_recv_sock": // maps/udpv6_recv_sock (BPF_MAP_TYPE_HASH), key C.__u64, value C.udp_recv_sock_t
		io.WriteString(w, "Map: '"+mapName+"', key: 'C.__u64', value: 'C.udp_recv_sock_t'\n")
		iter := currentMap.Iterate()
		var key uint64
		var value ddebpf.UDPRecvSock
		for iter.Next(unsafe.Pointer(&key), unsafe.Pointer(&value)) {
			spew.Fdump(w, key, value)
		}

	case probes.PortBindingsMap: // maps/port_bindings (BPF_MAP_TYPE_HASH), key portBindingTuple, value C.__u8
		io.WriteString(w, "Map: '"+mapName+"', key: 'portBindingTuple', value: 'C.__u8'\n")
		iter := currentMap.Iterate()
		var key ddebpf.PortBinding
		var value uint8
		for iter.Next(unsafe.Pointer(&key), unsafe.Pointer(&value)) {
			spew.Fdump(w, key, value)
		}

	case probes.UDPPortBindingsMap: // maps/udp_port_bindings (BPF_MAP_TYPE_HASH), key portBindingTuple, value C.__u8
		io.WriteString(w, "Map: '"+mapName+"', key: 'portBindingTuple', value: 'C.__u8'\n")
		iter := currentMap.Iterate()
		var key ddebpf.PortBinding
		var value uint8
		for iter.Next(unsafe.Pointer(&key), unsafe.Pointer(&value)) {
			spew.Fdump(w, key, value)
		}

	case "pending_bind": // maps/pending_bind (BPF_MAP_TYPE_HASH), key C.__u64, value C.bind_syscall_args_t
		io.WriteString(w, "Map: '"+mapName+"', key: 'C.__u64', value: 'C.bind_syscall_args_t'\n")
		iter := currentMap.Iterate()
		var key uint64
		var value ddebpf.BindSyscallArgs
		for iter.Next(unsafe.Pointer(&key), unsafe.Pointer(&value)) {
			spew.Fdump(w, key, value)
		}

	case probes.TelemetryMap: // maps/telemetry (BPF_MAP_TYPE_ARRAY), key C.u32, value kernelTelemetry
		io.WriteString(w, "Map: '"+mapName+"', key: 'C.u32', value: 'kernelTelemetry'\n")
		var zero uint64
		telemetry := &ddebpf.Telemetry{}
		if err := currentMap.Lookup(unsafe.Pointer(&zero), unsafe.Pointer(telemetry)); err != nil {
			// This can happen if we haven't initialized the telemetry object yet
			// so let's just use a trace log
			log.Tracef("error retrieving the telemetry struct: %s", err)
		}
		spew.Fdump(w, telemetry)
	}
}
