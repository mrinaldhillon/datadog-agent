// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build linux

package runtime

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/DataDog/datadog-agent/cmd/system-probe/command"
	"github.com/DataDog/datadog-agent/pkg/security/proto/api"
)

// Commands returns the config commands
func Commands(globalParams *command.GlobalParams) []*cobra.Command {
	runtimeCmd := &cobra.Command{
		Use:   "runtime",
		Short: "Runtime Security Agent (CWS) utility commands",
	}

	runtimeCmd.AddCommand(commonPolicyCommands(globalParams)...)
	runtimeCmd.AddCommand(selfTestCommands(globalParams)...)
	runtimeCmd.AddCommand(activityDumpCommands(globalParams)...)
	runtimeCmd.AddCommand(securityProfileCommands(globalParams)...)
	runtimeCmd.AddCommand(processCacheCommands(globalParams)...)
	runtimeCmd.AddCommand(networkNamespaceCommands(globalParams)...)
	runtimeCmd.AddCommand(discardersCommands(globalParams)...)

	return []*cobra.Command{runtimeCmd}
}

func printSecurityActivityDumpMessage(prefix string, msg *api.ActivityDumpMessage) {
	fmt.Printf("%s- name: %s\n", prefix, msg.GetMetadata().GetName())
	fmt.Printf("%s  start: %s\n", prefix, msg.GetMetadata().GetStart())
	fmt.Printf("%s  timeout: %s\n", prefix, msg.GetMetadata().GetTimeout())
	if len(msg.GetMetadata().GetContainerID()) > 0 {
		fmt.Printf("%s  container ID: %s\n", prefix, msg.GetMetadata().GetContainerID())
	}
	if len(msg.GetMetadata().GetCGroupID()) > 0 {
		fmt.Printf("%s  cgroup ID: %s\n", prefix, msg.GetMetadata().GetCGroupID())
	}
	if len(msg.GetTags()) > 0 {
		fmt.Printf("%s  tags: %s\n", prefix, strings.Join(msg.GetTags(), ", "))
	}
	fmt.Printf("%s  differentiate args: %v\n", prefix, msg.GetMetadata().GetDifferentiateArgs())
	printActivityTreeStats(prefix, msg.GetStats())
	if len(msg.GetStorage()) > 0 {
		fmt.Printf("%s  storage:\n", prefix)
		for _, storage := range msg.GetStorage() {
			printStorageRequestMessage(prefix+"\t", storage)
		}
	}
}
