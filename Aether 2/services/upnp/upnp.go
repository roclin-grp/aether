// Services > UPNP
// This module provides UPNP port mapping functionality for routers, so that a node that is behind a router can still be accessed by other nodes.

package upnp

import (
	"aether-core/services/globals"
	"aether-core/services/logging"
	"fmt"
	extUpnp "github.com/NebulousLabs/go-upnp"
	// "time"
)

// var router *extUpnp.IGD
// var err error

func MapPort() {
	// External port verification has moved to the start of the main routine, so we do no longer need to do that here. We can assume that at this point, the external port is verified.
	router, err := extUpnp.Discover()
	if err != nil {
		// Either could not be found, or connected to the internet directly.
		logging.Log(1, fmt.Sprintf("A router to port map could not be found. This computer could be directly connected to the Internet without a router. Error: %s", err.Error()))
		return
	}
	extIp, err2 := router.ExternalIP()
	if err2 != nil {
		// External IP finding failed.
		logging.Log(1, fmt.Sprintf("External IP of this machine could not be determined. Error: %s", err2.Error()))
	} else {
		globals.BackendConfig.SetExternalIp(extIp)
		logging.Log(1, fmt.Sprintf("This computer's external IP is %s", globals.BackendConfig.GetExternalIp()))
	}
	err3 := router.Forward(globals.BackendConfig.GetExternalPort(), "Aether")
	if err3 != nil {
		// Router is there, but port mapping failed.
		logging.Log(1, fmt.Sprintf("In an attempt to port map, the router was found, but the port mapping failed. Error: %s", err3.Error()))
	}
	logging.Log(1, fmt.Sprintf("Port mapping was successful. We mapped port %d to this computer.", globals.BackendConfig.GetExternalPort()))
}

// func UnmapPort() {
// 	// Attempt to remove the port mapping on quit.
// 	_ = router.Clear(globals.AddressPort)
// }
