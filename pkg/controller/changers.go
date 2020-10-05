package controller

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/lf-edge/eden/pkg/controller/adam"
	"github.com/lf-edge/eden/pkg/device"
	"github.com/lf-edge/eve/api/go/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"strings"
)

type ConfigChanger interface {
	GetControllerAndDev() (Cloud, *device.Ctx, error)
	SetControllerAndDev(Cloud, *device.Ctx) error
}

type FileChanger struct {
	fileConfig string
	oldHash    [32]byte
}

func GetFileChanger(fileConfig string) *FileChanger {
	return &FileChanger{fileConfig: fileConfig}
}

func (ctx *FileChanger) GetControllerAndDev() (Cloud, *device.Ctx, error) {
	if ctx.fileConfig == "" {
		return nil, nil, fmt.Errorf("cannot use empty url for file")
	}
	if _, err := os.Lstat(ctx.fileConfig); os.IsNotExist(err) {
		return nil, nil, err
	}
	var ctrl Cloud = &CloudCtx{Controller: &adam.Ctx{}}
	data, err := ioutil.ReadFile(ctx.fileConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("file reading error: %s", err)
	}
	var deviceConfig config.EdgeDevConfig
	err = json.Unmarshal(data, &deviceConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("unmarshal error: %s", err)
	}
	dev, err := ctrl.ConfigParse(&deviceConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("configParse error: %s", err)
	}
	res, err := ctrl.GetConfigBytes(dev, false)
	if err != nil {
		return nil, nil, fmt.Errorf("GetConfigBytes error: %s", err)
	}
	ctx.oldHash = sha256.Sum256(res)
	return ctrl, dev, nil
}

func (ctx *FileChanger) SetControllerAndDev(ctrl Cloud, dev *device.Ctx) error {
	res, err := ctrl.GetConfigBytes(dev, false)
	if err != nil {
		return fmt.Errorf("GetConfigBytes error: %s", err)
	}
	newHash := sha256.Sum256(res)
	if ctx.oldHash == newHash {
		log.Debug("config not modified")
		return nil
	}
	if res, err = VersionIncrement(res); err != nil {
		return fmt.Errorf("VersionIncrement error: %s", err)
	}
	if err = ioutil.WriteFile(ctx.fileConfig, res, 0755); err != nil {
		return fmt.Errorf("WriteFile error: %s", err)
	}
	log.Debug("config modification done")
	return nil
}

type AdamChanger struct {
	adamUrl string
}

func GetAdamChanger(url ...string) *AdamChanger {
	if len(url) > 0 {
		return &AdamChanger{adamUrl: url[0]}
	}
	return &AdamChanger{}
}

func (ctx *AdamChanger) getController() (Cloud, error) {
	if ctx.adamUrl != "" { //overwrite config only if url defined
		ipPort := strings.Split(ctx.adamUrl, ":")
		ip := ipPort[0]
		if ip == "" {
			return nil, fmt.Errorf("cannot get ip/hostname from %s", ctx.adamUrl)
		}
		port := "80"
		if len(ipPort) > 1 {
			port = ipPort[1]
		}
		viper.Set("adam.ip", ip)
		viper.Set("adam.port", port)
	}
	ctrl, err := CloudPrepare()
	if err != nil {
		return nil, fmt.Errorf("CloudPrepare error: %s", err)
	}
	return ctrl, nil
}

func (ctx *AdamChanger) GetControllerAndDev() (Cloud, *device.Ctx, error) {
	ctrl, err := ctx.getController()
	if err != nil {
		return nil, nil, fmt.Errorf("getController error: %s", err)
	}
	devFirst, err := ctrl.GetDeviceCurrent()
	if err != nil {
		return nil, nil, fmt.Errorf("GetDeviceCurrent error: %s", err)
	}
	return ctrl, devFirst, nil
}

func (ctx *AdamChanger) SetControllerAndDev(ctrl Cloud, dev *device.Ctx) error {
	if err := ctrl.ConfigSync(dev); err != nil {
		return fmt.Errorf("configSync error: %s", err)
	}
	return nil
}
