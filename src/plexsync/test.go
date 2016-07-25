package main

import (
	"os"
	"testing"

	"code.senomas.com/go/plexapi"
	log "github.com/Sirupsen/logrus"
	_ "github.com/mattn/go-sqlite3"
)

// TestSimple func
func TestSimple(t *testing.T) {
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

	// server, _ := api.GetServer("senomas")
	server, _ := api.GetServer("My SHIELD Android TV")
	log.Debug("SERVER ", server)

	// c, _ := server.GetContainer("/library/sections")
	// // c, _ := server.GetContainer("/library/metadata/4980")
	// log.Debug("DATA ", util.JSONPrettyPrint(c))

	// server.Perform("GET", "/library/metadata/4980")
	// server.Perform("GET", "/:/progress?key=4980&identifier=com.plexapp.plugins.library&time=100000")
	// server.Perform("PUT", "/library/metadata/4980?viewOffset.value=100")

	// server.MarkUnwatched("4980")
	//
	// content, _ := server.GetContainer("/library/metadata/4980")
	// v := content.Videos[0]
	//
	// m, _ := repo.GetMedia("/Batman v Superman Dawn of Justice (2016)/Batman v Superman Dawn of Justice.mkv")
	//
	// log.Debugf("UPDATE AT       %v : %v  %v", m.UpdatedAt, v.UpdatedAt, m.UpdatedAt == v.UpdatedAt)
	// log.Debugf("LAST VIEWED AT  %v : %v", m.LastViewedAt, v.LastViewedAt)
	// log.Debugf("VIEW OFFSET     %v : %v", m.ViewOffset, v.ViewOffset)

	// content, _ = server.GetContainer("/library/metadata/4979")
	// log.Debug("CONTENT ", util.JSONPrettyPrint(content))

	// http://<CLIENT IP>:<CLIENT PORT>/player/playback/playMedia?key=%2Flibrary%2Fmetadata%2F<MEDIA ID>&offset=0&X-Plex-Client-Identifier=<CLIENT ID>&machineIdentifier=<SERVER ID>&address=<SERVER IP>&port=<SERVER PORT>&protocol=http&path=http%3A%2F%2F<SERVER IP>%3A<SERVER PORT>%2Flibrary%2Fmetadata%2F<MEDIA ID>

	// var wg sync.WaitGroup
	// out := make(chan interface{})
	//
	// server.GetVideos(&wg, out)
	//
	// go func() {
	// 	for o := range out {
	// 		switch o := o.(type) {
	// 		case plexapi.Video:
	// 			v := plexapi.Video(o)
	// 			log.Debug("VIDEO ", util.JSONPrettyPrint(v))
	// 		}
	// 	}
	// }()
	//
	// wg.Wait()
}
