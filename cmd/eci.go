package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/lf-edge/eden/pkg/utils"
	edgeRegistry "github.com/lf-edge/edge-containers/pkg/registry"
	"github.com/lf-edge/edge-containers/pkg/resolver"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	kernelFile   string
	initrdFile   string
	rootFile     string
	arch         string
	remoteTarget resolver.ResolverCloser
)

// convert a "path:type" to a Disk struct
func diskToStruct(path string) (*edgeRegistry.Disk, error) {
	parts := strings.SplitN(path, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("expected structure <path>:<type>")
	}
	// get the disk type
	diskType, ok := edgeRegistry.NameToType[parts[1]]
	if !ok {
		return nil, fmt.Errorf("unknown disk type: %s", parts[1])
	}
	return &edgeRegistry.Disk{
		Source: &edgeRegistry.FileSource{Path: parts[0]},
		Type:   diskType,
	}, nil
}

//podPublishCmd is a command to publish files into edge-container
var podPublishCmd = &cobra.Command{
	Use:   "publish <image>",
	Short: "Publish pod files into image",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		assignCobraToViper(cmd)
		_, err := utils.LoadConfigFile(configFile)
		if err != nil {
			return fmt.Errorf("error reading config: %s", err.Error())
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]
		var (
			rootDisk     *edgeRegistry.Disk
			kernelSource *edgeRegistry.FileSource
			initrdSource *edgeRegistry.FileSource
			err          error
		)
		ctx := context.TODO()
		_, remoteTarget, err = resolver.NewRegistry(ctx)
		if err != nil {
			log.Fatalf("unexpected error when created NewRegistry resolver: %v", err)
		}
		if rootFile != "" {
			rootDisk, err = diskToStruct(rootFile)
			if err != nil {
				log.Fatalf("unable to read root disk %s: %v", rootFile, err)
			}
		}
		if kernelFile != "" {
			kernelSource = &edgeRegistry.FileSource{Path: kernelFile}
		}
		if initrdFile != "" {
			initrdSource = &edgeRegistry.FileSource{Path: initrdFile}
		}
		artifact := &edgeRegistry.Artifact{
			Kernel: kernelSource,
			Initrd: initrdSource,
			Root:   rootDisk,
		}
		if kernelFile == "" {
			artifact.Kernel = nil
		}
		if initrdFile == "" {
			artifact.Initrd = nil
		}
		pusher := edgeRegistry.Pusher{
			Artifact: artifact,
			Image:    appName,
		}
		hash, err := pusher.Push(edgeRegistry.FormatArtifacts, true, os.Stdout, edgeRegistry.ConfigOpts{
			Author:       edgeRegistry.DefaultAuthor,
			OS:           edgeRegistry.DefaultOS,
			Architecture: arch,
		}, remoteTarget)
		if err != nil {
			log.Fatalf("error pushing to registry: %v", err)
		}
		fmt.Printf("Pushed image %s with digest %s\n", image, hash)
	},
}

func eciInit() {
	podCmd.AddCommand(podPublishCmd)
	podPublishCmd.Flags().StringVar(&kernelFile, "kernel", "", "path to kernel file, optional")
	podPublishCmd.Flags().StringVar(&initrdFile, "initrd", "", "path to initrd file, optional")
	podPublishCmd.Flags().StringVar(&rootFile, "root", "", "path to root disk file and format (for example: image.img:qcow2)")
	podPublishCmd.Flags().StringVarP(&arch, "arch", "", edgeRegistry.DefaultArch, "arch to deploy")
}
