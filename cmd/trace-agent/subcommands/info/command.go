// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// Package info contains the 'info' subcommand for the 'trace-agent' command.
package info

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/DataDog/datadog-agent/cmd/trace-agent/subcommands"
	"github.com/DataDog/datadog-agent/comp/api/authtoken/fetchonlyimpl"
	coreconfig "github.com/DataDog/datadog-agent/comp/core/config"
	log "github.com/DataDog/datadog-agent/comp/core/log/def"
	logfx "github.com/DataDog/datadog-agent/comp/core/log/fx"
	"github.com/DataDog/datadog-agent/comp/core/secrets"
	"github.com/DataDog/datadog-agent/comp/core/secrets/secretsimpl"
	nooptagger "github.com/DataDog/datadog-agent/comp/core/tagger/fx-noop"
	"github.com/DataDog/datadog-agent/comp/trace/config"
	"github.com/DataDog/datadog-agent/pkg/trace/info"
	"github.com/DataDog/datadog-agent/pkg/util/fxutil"
	"github.com/DataDog/datadog-agent/pkg/util/option"
)

// MakeCommand returns the start subcommand for the 'trace-agent' command.
func MakeCommand(globalParamsGetter func() *subcommands.GlobalParams) *cobra.Command {
	infoCmd := &cobra.Command{
		Use:   "info",
		Short: "Gather Datadog trace-agent information.",
		Long:  `Use this to gather information on the running trace-agent`,
		RunE: func(*cobra.Command, []string) error {
			return runTraceAgentInfoFct(globalParamsGetter(), agentInfo)
		},
	}

	return infoCmd
}

func runTraceAgentInfoFct(params *subcommands.GlobalParams, fct interface{}) error {
	return fxutil.OneShot(fct,
		config.Module(),
		fx.Supply(coreconfig.NewAgentParams(params.ConfPath, coreconfig.WithFleetPoliciesDirPath(params.FleetPoliciesDirPath))),
		fx.Supply(log.ForOneShot(params.LoggerName, "off", true)),
		fx.Supply(option.None[secrets.Component]()),
		fx.Supply(secrets.NewEnabledParams()),
		coreconfig.Module(),
		secretsimpl.Module(),
		nooptagger.Module(),
		fetchonlyimpl.Module(),
		logfx.Module(),
	)
}

func agentInfo(config config.Component) error {
	tracecfg := config.Object()
	if tracecfg == nil {
		return fmt.Errorf("Unable to successfully parse config")
	}
	if err := info.InitInfo(tracecfg); err != nil {
		return err
	}
	return info.Info(os.Stdout, tracecfg)
}
