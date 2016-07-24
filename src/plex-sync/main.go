package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"

	"code.senomas.com/go/plexapi"
	"code.senomas.com/go/util"
	log "github.com/Sirupsen/logrus"
)

func atoi(v string) int {
	if v == "" {
		return 0
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		log.Fatalf("Unable to parse '%s'", v)
	}
	return i
}

func main() {
	log.SetOutput(os.Stderr)
	// log.SetLevel(log.InfoLevel)
	log.SetLevel(log.DebugLevel)

	repo := plexapi.Repo{}
	err := repo.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer repo.Close()

	api := plexapi.API{HTTP: plexapi.HTTPConfig{Timeout: 30, WorkerSize: 10}}
	api.LoadConfig("config.yaml")

	var wg sync.WaitGroup
	out := make(chan interface{})

	servers, err := api.GetServers()
	util.Panicf("GetServers failed %v", err)
	for _, server := range servers {
		server.GetVideos(&wg, out)
	}

	var videos []plexapi.Video
	go func() {
		for o := range out {
			switch o := o.(type) {
			case plexapi.Video:
				v := plexapi.Video(o)
				if v.FID != "" && !strings.HasPrefix(v.FID, "local://") {
					// log.WithField("server", v.Server.Name).WithField("guid", v.GUID).WithField("title", v.Title).WithField("viewCount", v.ViewCount).WithField("lastViewedAt", v.LastViewedAt).Info("MEDIA")

					videos = append(videos, v)
					_, err := repo.GetViewStatus(v.Server.Name, v.FID)
					if err != nil {
						log.Warn("Error GetMedia ", v.FID, " ", err)
					}
					repo.Save(&v)
					// if media != nil {
					// 	if atoi(v.UpdatedAt) > atoi(media.UpdatedAt) {
					// 		repo.Save(&v)
					// 		log.WithField("server", v.Server.Name).WithField("key", v.Key).WithField("id", v.FID).Info("UPDATE")
					// 	} else {
					// 		log.WithField("server", v.Server.Name).WithField("key", v.Key).WithField("id", v.FID).Info("SKIP")
					// 	}
					// } else {
					// 	repo.Save(&v)
					// 	log.WithField("server", v.Server.Name).WithField("key", v.Key).WithField("id", v.FID).Info("SAVE")
					// }
				}
			default:
				fmt.Printf("Type of o is %T. Value %v", o, o)
			}
		}
	}()

	wg.Wait()
	fmt.Println()

	for _, v := range videos {
		if v.FID != "" && !strings.HasPrefix(v.FID, "local://") {
			media, err := repo.GetMedia(v.FID)
			if err != nil {
				log.Warn("Error GetMedia ", v.FID, " ", err)
			}
			if media != nil {
				if atoi(v.UpdatedAt) < atoi(media.UpdatedAt) {
					mvc, vvc := atoi(media.ViewCount), atoi(v.ViewCount)
					if mvc > 0 {
						if mvc != vvc {
							log.WithField("id", v.FID).WithField("title", v.Title).Infof("NEED UPDATE VIEW-COUNT %v %v %s", mvc, vvc, v.Server.Name)
						} else {
							log.WithField("id", v.FID).WithField("title", v.Title).Infof("VIEW-COUNT OK %s", v.Server.Name)
						}
					} else {
						log.WithField("id", v.FID).WithField("title", v.Title).Infof("NEED UPDATE VIEW-OFFSET %v %s", atoi(media.ViewOffset), v.Server.Name)
					}
					// if media.LastViewedAt != "" && v.LastViewedAt == "" {
					// 	log.WithField("server", v.Server.Name).WithField("id", v.FID).WithField("title", v.Title).Info("MARKED WATCHED")
					// 	// v.Server.MarkWatched(v)
					// 	// } else if media.LastViewedAt == "" && v.LastViewedAt != "" {
					// 	// 	log.WithField("server", v.Server.Name).WithField("guid", v.GUID).WithField("title", v.Title).Info("MARKED UNWATCHED")
					// 	// 	v.Server.MarkUnwatched(v)
					// }
				}
			} else {
				log.Fatal("NO MEDIA ", util.JSONPrettyPrint(v))
			}
		}
	}
}
