package discovery

import (
	"log"
	"strings"

	"github.com/muka/go-bluetooth/api"
	"github.com/muka/go-bluetooth/bluez/profile/adapter"
	"github.com/muka/go-bluetooth/bluez/profile/device"
)

type (
	// DiscoveredDeviceData ...
	DiscoveredDeviceData struct {
		Adapter *adapter.Adapter1
		Dd      *device.Device1
		Err     error
	}
)

// Run runs a discovery
func Run(adapterID, deviceName string) <-chan DiscoveredDeviceData {
	ddStream := make(chan DiscoveredDeviceData)
	go func() {
		defer func() {
			close(ddStream)
			log.Println("Exiting out of discovery goroutine.")
		}()
		log.Println("Running goroutine for discovery of devices ...")
		a, err := adapter.GetAdapter(adapterID)
		if err != nil {
			log.Printf("There was an error in getting the adpter from the adpterID given the error is %s", err.Error())
			ddStream <- DiscoveredDeviceData{Err: err}
			return
		}
		log.Println("Flushing the device list for the given adapter")
		err = a.FlushDevices()
		if err != nil {
			log.Printf("There was an error in flushing adpter from the adpterID given the error is %s", err.Error())
			ddStream <- DiscoveredDeviceData{Err: err}
			return
		}
		log.Println("Starting discovery ...")
		discovery, cancel, err := api.Discover(a, nil)
		if err != nil {
			log.Printf("There was an error in Discovering devices the error is %s", err.Error())
			ddStream <- DiscoveredDeviceData{Err: err}
			return
		}
		defer cancel()
		for d := range discovery {
			if d.Type == adapter.DeviceRemoved {
				continue
			}

			dev, err := device.NewDevice1(d.Path)
			if err != nil {
				log.Printf("%s: %s", d.Path, err)
				continue
			}

			if dev == nil {
				log.Printf("%s: not found", d.Path)
				continue
			}
			log.Printf("name=%s addr=%s rssi=%d", dev.Properties.Alias, dev.Properties.Address, dev.Properties.RSSI)
			ddStream <- DiscoveredDeviceData{
				Adapter: a,
				Dd:      dev,
			}
			if strings.ToLower(dev.Properties.Alias) == strings.ToLower(deviceName) {
				log.Printf("%s: found returning", deviceName)
				return
			}
		}
	}()
	return ddStream
}

// Name returns the name of the fuction.
func Name() string {
	return "discovery"
}
