package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/ameena3/blueTooth/agent"
	"github.com/ameena3/blueTooth/discovery"
	"github.com/muka/go-bluetooth/api"
)

func main() {
	defer api.Exit()
	log.Printf("Starting scanning for %s", os.Args[1])
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, os.Kill) // get notified of all OS signals

	dd := discovery.Run("hci0", os.Args[1])
	a, err := agent.NewAdapter(&agent.Params{
		AdapterID: "hci0",
	})
	if err != nil {
		log.Print(err.Error())
		return
	}
	var addressToConnect string
out:
	for {
		select {
		case sig := <-ch:
			log.Printf("Received signal [%v]; shutting down...\n", sig)
			return
		case d, ok := <-dd:
			if d.Err != nil {
				log.Printf("The error is %s", d.Err.Error())
			}
			if !ok {
				log.Println("The discovery stream is closed exiting discovery phase.")
				break out
			}
			if strings.ToLower(d.Dd.Properties.Alias) == strings.ToLower(os.Args[1]) {
				err = a.Connect(d.Dd.Properties.Address)
				if err != nil {
					log.Print(err.Error())
					return
				}
				addressToConnect = d.Dd.Properties.Address
			}
		}
	}

	// Create a ticker that tries to connect every 10 second.
	tick := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-tick.C:
			log.Println("Trying to connect again")
			err = a.Connect(addressToConnect)
			if err != nil {
				log.Print(err.Error())
				return
			}
		}
	}

}
