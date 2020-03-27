package main

import (
	"github.com/itmo-eve/eden/pkg/cloud"
	"github.com/itmo-eve/eden/pkg/device"
	uuid "github.com/satori/go.uuid"
	"log"
)

func main() {
	cloudCxt := &cloud.Ctx{}
	deviceCtx := device.CreateWithBaseConfig(uuid.NewV4(), cloudCxt)
	b, err := deviceCtx.GenerateJSONBytes()
	if err != nil {
		log.Fatal(err)
	}
	log.Print(string(b))
}
