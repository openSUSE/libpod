package main

import (
	"fmt"

	"github.com/containers/libpod/cmd/podman/cliconfig"
	"github.com/containers/libpod/pkg/adapter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	podRestartCommand     cliconfig.PodRestartValues
	podRestartDescription = `Restarts one or more pods. The pod ID or name can be used.`
	_podRestartCommand    = &cobra.Command{
		Use:   "restart [flags] POD [POD...]",
		Short: "Restart one or more pods",
		Long:  podRestartDescription,
		RunE: func(cmd *cobra.Command, args []string) error {
			podRestartCommand.InputArgs = args
			podRestartCommand.GlobalFlags = MainGlobalOpts
			return podRestartCmd(&podRestartCommand)
		},
		Args: func(cmd *cobra.Command, args []string) error {
			return checkAllAndLatest(cmd, args, false)
		},
		Example: `podman pod restart podID1 podID2
  podman pod restart --latest
  podman pod restart --all`,
	}
)

func init() {
	podRestartCommand.Command = _podRestartCommand
	podRestartCommand.SetUsageTemplate(UsageTemplate())
	flags := podRestartCommand.Flags()
	flags.BoolVarP(&podRestartCommand.All, "all", "a", false, "Restart all running pods")
	flags.BoolVarP(&podRestartCommand.Latest, "latest", "l", false, "Restart the latest pod podman is aware of")

	markFlagHiddenForRemoteClient("latest", flags)
}

func podRestartCmd(c *cliconfig.PodRestartValues) error {
	var lastError error
	runtime, err := adapter.GetRuntime(&c.PodmanCommand)
	if err != nil {
		return errors.Wrapf(err, "could not get runtime")
	}
	defer runtime.Shutdown(false)

	restartIDs, conErrors, restartErrors := runtime.RestartPods(getContext(), c)

	for _, p := range restartIDs {
		fmt.Println(p)
	}
	if conErrors != nil && len(conErrors) > 0 {
		for ctr, err := range conErrors {
			if lastError != nil {
				logrus.Errorf("%q", lastError)
			}
			lastError = errors.Wrapf(err, "unable to pause container %s", ctr)
		}
	}
	if len(restartErrors) > 0 {
		lastError = restartErrors[len(restartErrors)-1]
		// Remove the last error from the error slice
		restartErrors = restartErrors[:len(restartErrors)-1]
	}
	for _, err := range restartErrors {
		logrus.Errorf("%q", err)
	}
	return lastError
}
