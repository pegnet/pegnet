package controlPanel

import (
	"net/http"
	"encoding/json"
	"github.com/pegnet/pegnet/common"
	"github.com/alexandrevicenzi/go-sse"
	log "github.com/sirupsen/logrus"
)

func ServeControlPanel(monitor *common.Monitor) {
	log.Info("Starting control panel on localhost:8080")

	alert := monitor.NewListener()

	// Create the server.
	s := sse.NewServer(&sse.Options{
		// Print debug info
		Logger: nil,
	})
	defer s.Shutdown()

	// Register with /events endpoint.
	http.Handle("/events/", s)

	// Dispatch messages to channel-1.
	go func () {
		for {
			select {
			case f := <- alert:
				data, _ := json.Marshal(f)
				s.SendMessage("/events/common", sse.SimpleMessage(string(data)))
			}
		}
	}()

	http.Handle("/", http.FileServer(http.Dir("./controlPanel/static")))
	http.ListenAndServe(":8080", nil)
	
}