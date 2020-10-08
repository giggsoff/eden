package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/lf-edge/adam/pkg/server"
	"github.com/lf-edge/eden/pkg/utils"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var exportCmd = &cobra.Command{
	Use:   "export <filename>",
	Short: "export harness",
	Long:  `Export harness.`,
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		assignCobraToViper(cmd)
		viperLoaded, err := utils.LoadConfigFile(configFile)
		if err != nil {
			return fmt.Errorf("error reading config: %s", err.Error())
		}
		if viperLoaded {
			certsDir = utils.ResolveAbsPath(viper.GetString("eden.certs-dist"))
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		changer := &adamChanger{}
		ctrl, dev, err := changer.getControllerAndDev()
		if err != nil {
			log.Fatalf("getControllerAndDev: %s", err)
		}
		deviceCert, err := ctrl.GetDeviceCert(dev)
		if err != nil {
			log.Warn(err)
		} else {
			if err = ioutil.WriteFile(ctrl.GetVars().EveDeviceCert, deviceCert.Cert, 0777); err != nil {
				log.Warn(err)
			}
		}
		edenDir, err := utils.DefaultEdenDir()
		if err != nil {
			log.Fatal(err)
		}
		tarFile := args[0]
		files := []utils.FileToSave{
			{Location: certsDir, Destination: "certs"},
			{Location: edenDir, Destination: "eden"},
		}
		if err := utils.CreateTar(tarFile, files); err != nil {
			log.Fatal(err)
		}
		log.Infof("Export Eden done")
	},
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "import harness",
	Long:  `Import harness.`,
	Run: func(cmd *cobra.Command, args []string) {
		changer := &adamChanger{}
		ctrl, err := changer.getController()
		if err != nil {
			log.Fatal(err)
		}
		devUUID, err := ctrl.DeviceGetByOnboard(ctrl.GetVars().EveCert)
		if err != nil {
			log.Debug(err)
		}
		if devUUID == uuid.Nil {
			if _, err := os.Stat(ctrl.GetVars().EveDeviceCert); os.IsNotExist(err) {
				log.Fatalf("No device cert in %s, you need to run 'eden export' first", ctrl.GetVars().EveDeviceCert)
			}
			if _, err := os.Stat(ctrl.GetVars().EveCert); os.IsNotExist(err) {
				log.Fatalf("No onboard cert in %s, you need to run 'eden setup' first", ctrl.GetVars().EveCert)
			}
			deviceCert, err := ioutil.ReadFile(ctrl.GetVars().EveDeviceCert)
			if err != nil {
				log.Fatal(err)
			}
			onboardCert, err := ioutil.ReadFile(ctrl.GetVars().EveCert)
			if err != nil {
				log.Warn(err)
			}
			dc := server.DeviceCert{
				Cert:   deviceCert,
				Serial: ctrl.GetVars().EveSerial,
			}
			if onboardCert != nil {
				dc.Onboard = onboardCert
			}
			err = ctrl.UploadDeviceCert(dc)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Info("Device already exists")
		}
		log.Infof("Import Eden done")
	},
}

func exportImportInit() {
}
