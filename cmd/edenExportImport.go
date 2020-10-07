package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lf-edge/eden/pkg/defaults"
	"github.com/lf-edge/eden/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "export harness",
	Long:  `Export harness.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		assignCobraToViper(cmd)
		viperLoaded, err := utils.LoadConfigFile(configFile)
		if err != nil {
			return fmt.Errorf("error reading config: %s", err.Error())
		}
		if viperLoaded {
			adamTag = viper.GetString("adam.tag")
			adamPort = viper.GetInt("adam.port")
			adamDist = utils.ResolveAbsPath(viper.GetString("adam.dist"))
			adamRemoteRedisURL = viper.GetString("adam.redis.adam")
			adamRemoteRedis = viper.GetBool("adam.remote.redis")
			redisTag = viper.GetString("redis.tag")
			redisPort = viper.GetInt("redis.port")
			redisDist = utils.ResolveAbsPath(viper.GetString("redis.dist"))
			eveImageFile = utils.ResolveAbsPath(viper.GetString("eve.image-file"))
			evePidFile = utils.ResolveAbsPath(viper.GetString("eve.pid"))
			eveDist = utils.ResolveAbsPath(viper.GetString("eve.dist"))
			adamDist = utils.ResolveAbsPath(viper.GetString("adam.dist"))
			certsDir = utils.ResolveAbsPath(viper.GetString("eden.certs-dist"))
			eserverImageDist = utils.ResolveAbsPath(viper.GetString("eden.images.dist"))
			qemuFileToSave = utils.ResolveAbsPath(viper.GetString("eve.qemu-config"))
			redisDist = utils.ResolveAbsPath(viper.GetString("redis.dist"))
			context, err := utils.ContextLoad()
			if err != nil {
				log.Fatalf("Load context error: %s", err)
			}
			configSaved = utils.ResolveAbsPath(fmt.Sprintf("%s-%s", context.Current, defaults.DefaultConfigSaved))
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		changer := &adamChanger{}
		ctrl, dev, err := changer.getControllerAndDev()
		if err != nil {
			log.Fatalf("getControllerAndDev: %s", err)
		}
		if err := ctrl.SaveDeviceCert(dev); err != nil {
			log.Warn(err)
		}
		log.Infof("Export Eden done")
	},
}

func exportImportInit() {
	currentPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	configDist, err := utils.DefaultEdenDir()
	if err != nil {
		log.Fatal(err)
	}
	exportCmd.Flags().StringVarP(&eveDist, "eve-dist", "", filepath.Join(currentPath, defaults.DefaultDist, defaults.DefaultEVEDist), "directory to save EVE")
	exportCmd.Flags().StringVarP(&redisDist, "redis-dist", "", "", "redis dist")
	exportCmd.Flags().StringVarP(&qemuFileToSave, "qemu-config", "", "", "file to save qemu config")
	exportCmd.Flags().StringVarP(&adamDist, "adam-dist", "", "", "adam dist to start (required)")
	exportCmd.Flags().StringVarP(&eserverImageDist, "image-dist", "", "", "image dist for eserver")

	exportCmd.Flags().StringVarP(&certsDir, "certs-dist", "o", filepath.Join(currentPath, defaults.DefaultDist, defaults.DefaultCertsDist), "directory with certs")
	exportCmd.Flags().StringVarP(&configDir, "config-dist", "", configDist, "directory for config")
	exportCmd.Flags().BoolVar(&currentContext, "current-context", true, "clean only current context")
}
