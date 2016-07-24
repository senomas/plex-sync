package main

import (
	"os"
	"sync"

	"code.senomas.com/go/plexapi"
	"code.senomas.com/go/util"
	log "github.com/Sirupsen/logrus"
)

func mainx() {
	log.SetOutput(os.Stderr)
	log.SetLevel(log.DebugLevel)

	log.Debug("TEST")

	api := plexapi.API{HTTP: plexapi.HTTPConfig{Timeout: 30, WorkerSize: 10}}
	api.LoadConfig("config.yaml")

	// servers, err := api.GetServers()
	// if err != nil {
	// 	log.Fatal("UNABLE TO GET SERVERS ", err)
	// }
	// for s := range servers {
	// 	log.Debug("SERVER ", util.JSONPrettyPrint(s))
	// }

	server, _ := api.GetServer("My SHIELD Android TV")
	log.Debug("SERVER ", server)

	// c, _ := server.GetContainer("/library/sections")
	// // c, _ := server.GetContainer("/library/metadata/4980")
	// log.Debug("DATA ", util.JSONPrettyPrint(c))

	// server.Perform("GET", "/library/metadata/4980")
	// server.Perform("GET", "/:/timeline?key=4980")
	// server.Perform("PUT", "/library/metadata/4980?viewOffset.value=100")

	// content, _ := server.GetContainer("/library/metadata/4980")
	// log.Debug("CONTENT ", util.JSONPrettyPrint(content))

	// content, _ = server.GetContainer("/library/metadata/4979")
	// log.Debug("CONTENT ", util.JSONPrettyPrint(content))

	// http://<CLIENT IP>:<CLIENT PORT>/player/playback/playMedia?key=%2Flibrary%2Fmetadata%2F<MEDIA ID>&offset=0&X-Plex-Client-Identifier=<CLIENT ID>&machineIdentifier=<SERVER ID>&address=<SERVER IP>&port=<SERVER PORT>&protocol=http&path=http%3A%2F%2F<SERVER IP>%3A<SERVER PORT>%2Flibrary%2Fmetadata%2F<MEDIA ID>

	var wg sync.WaitGroup
	out := make(chan interface{})

	server.GetVideos(&wg, out)

	go func() {
		for o := range out {
			switch o := o.(type) {
			case plexapi.Video:
				v := plexapi.Video(o)
				log.Debug("VIDEO ", util.JSONPrettyPrint(v))
			}
		}
	}()

	wg.Wait()
}
