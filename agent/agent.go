package agent

import (
	"errors"
	"fmt"
	"log"

	"github.com/godbus/dbus/v5"
	"github.com/muka/go-bluetooth/bluez/profile/adapter"
	bagent "github.com/muka/go-bluetooth/bluez/profile/agent"
)

type (
	agent struct {
		adapter         *adapter.Adapter1
		agentRegistered bool
	}

	//Params are the parameters required for creating a new adapter
	Params struct {
		AdapterID string
	}
)

// NewAdapter eturns a new adapter for connection
func NewAdapter(p *Params) (*agent, error) {
	if p.AdapterID == "" {
		return nil, errors.New("adapter is required in params")
	}
	ad, err := adapter.GetAdapter(p.AdapterID)
	if err != nil {
		return nil, err
	}
	return &agent{
		adapter: ad,
	}, nil
}

// Connect this particular device
func (a *agent) Connect(deviceAddress string) (err error) {
	//Connect DBus System bus
	conn, err := dbus.SystemBus()
	if err != nil {
		return err
	}
	if !a.agentRegistered {
		ag := bagent.NewSimpleAgent()
		err = bagent.ExposeAgent(conn, ag, bagent.CapKeyboardDisplay, true)
		if err != nil {
			return fmt.Errorf("SimpleAgent: %s", err)
		}
		a.agentRegistered = true
	}

	devices, err := a.adapter.GetDevices()
	if err != nil {
		return fmt.Errorf("GetDevices: %s", err)
	}
	aID, err := a.adapter.GetAdapterID()
	found := false
	for _, dev := range devices {

		if dev.Properties.Address != deviceAddress {
			continue
		}

		if dev.Properties.Paired {
			found = true
			continue
		}

		found = true
		// log.Info(i, v.Path)
		log.Printf("Pairing with %s", dev.Properties.Address)

		err := dev.Pair()
		if err != nil {
			return fmt.Errorf("Pair failed: %s", err)
		}

		log.Printf("Pair succeed, connecting...")

		if err != nil {
			return fmt.Errorf("Get AdapterID failed: %s", err)
		}
		bagent.SetTrusted(aID, dev.Path())

		err = dev.Connect()
		if err != nil {
			return fmt.Errorf("Connect failed: %s", err)
		}

	}

	if !found {
		return fmt.Errorf("No device found that need to be paired on %s", aID)
	}
	return
}
