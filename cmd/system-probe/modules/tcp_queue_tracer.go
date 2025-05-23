// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build linux && linux_bpf

package modules

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/DataDog/datadog-agent/pkg/collector/corechecks/ebpf/probe/tcpqueuelength"
	"github.com/DataDog/datadog-agent/pkg/ebpf"
	"github.com/DataDog/datadog-agent/pkg/system-probe/api/module"
	"github.com/DataDog/datadog-agent/pkg/system-probe/config"
	sysconfigtypes "github.com/DataDog/datadog-agent/pkg/system-probe/config/types"
	"github.com/DataDog/datadog-agent/pkg/system-probe/utils"
)

func init() { registerModule(TCPQueueLength) }

// TCPQueueLength Factory
var TCPQueueLength = &module.Factory{
	Name:             config.TCPQueueLengthTracerModule,
	ConfigNamespaces: []string{},
	Fn: func(_ *sysconfigtypes.Config, _ module.FactoryDependencies) (module.Module, error) {
		t, err := tcpqueuelength.NewTracer(ebpf.NewConfig())
		if err != nil {
			return nil, fmt.Errorf("unable to start the TCP queue length tracer: %w", err)
		}

		return &tcpQueueLengthModule{
			Tracer: t,
		}, nil
	},
	NeedsEBPF: func() bool {
		return true
	},
}

var _ module.Module = &tcpQueueLengthModule{}

type tcpQueueLengthModule struct {
	*tcpqueuelength.Tracer
	lastCheck atomic.Int64
}

func (t *tcpQueueLengthModule) Register(httpMux *module.Router) error {
	httpMux.HandleFunc("/check", func(w http.ResponseWriter, _ *http.Request) {
		t.lastCheck.Store(time.Now().Unix())
		stats := t.Tracer.GetAndFlush()
		utils.WriteAsJSON(w, stats)
	})

	return nil
}

func (t *tcpQueueLengthModule) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"last_check": t.lastCheck.Load(),
	}
}
