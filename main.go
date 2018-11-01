package main

import (
	"errors"
	"os"

	restore "github.com/maorfr/helm-restore/pkg"
	"github.com/spf13/cobra"
)

var (
	releaseName     string
	tillerNamespace string
	label           string
)

func main() {
	cmd := &cobra.Command{
		Use:   "restore [flags] RELEASE_NAME",
		Short: "restore last deployed release to original state",
		RunE:  run,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("RELEASE_NAME is required")
			}
			return nil
		},
	}

	f := cmd.Flags()
	f.StringVar(&tillerNamespace, "tiller-namespace", "kube-system", "namespace of Tiller")
	f.StringVarP(&label, "label", "l", "OWNER=TILLER,STATUS=DEPLOYED", "label to select tiller resources by")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	releaseName := args[0]
	if err := restore.Restore(releaseName, tillerNamespace, label); err != nil {
		return err
	}
	return nil
}
