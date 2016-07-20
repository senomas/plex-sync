package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"

	_ "github.com/mattn/go-sqlite3"

	"code.senomas.com/go/plexapi"
	"code.senomas.com/go/util"
	log "github.com/Sirupsen/logrus"
)

// Repo struct
type Repo struct {
	db          *sql.DB
	findGUID    *sql.Stmt
	insertGUID  *sql.Stmt
	findMedia   *sql.Stmt
	insertMedia *sql.Stmt
}

func (repo *Repo) init() {
	var err error

	repo.db, err = sql.Open("sqlite3", "./plex.db")
	util.Panicf("Open DB %v", err)

	_, err = repo.db.Exec("create table if not exists media(guid text primary key, title text, addedAt int, updatedAt int, viewCount int, lastViewedAt int)")
	if err != nil {
		log.Fatal("Failed to create table media", err)
	}

	repo.findMedia, err = repo.db.Prepare("select title, addedAt, updatedAt, viewCount, lastViewedAt from media where guid = ?")
	if err != nil {
		log.Fatal(err)
	}

	repo.insertMedia, err = repo.db.Prepare("insert or replace into media(guid, title, addedAt, updatedAt, viewCount, lastViewedAt) values(?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
}

func (repo *Repo) getMedia(guid string) *plexapi.Video {
	rows, err := repo.findMedia.Query(guid)
	util.Panicf("getMedia %v", err)
	defer rows.Close()

	if !rows.Next() {
		return nil
	}

	var video plexapi.Video
	video.GUID = guid
	rows.Scan(&video.Title, &video.AddedAt, &video.UpdatedAt, &video.ViewCount, &video.LastViewedAt)
	return &video
}

func (repo *Repo) save(v plexapi.Video) {
	_, err := repo.insertMedia.Exec(v.GUID, v.Title, v.AddedAt, v.UpdatedAt, v.ViewCount, v.LastViewedAt)
	util.Panicf("InsertMedia failed %v", err)
}

func (repo *Repo) close() {
	repo.findMedia.Close()
	repo.insertMedia.Close()
	repo.db.Close()
}

func main() {
	log.SetOutput(os.Stderr)
	log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	repo := Repo{}
	repo.init()
	defer repo.close()

	api := plexapi.API{HTTP: plexapi.HTTPConfig{Timeout: 30, WorkerSize: 10}}
	cfg, _ := ioutil.ReadFile("config.yaml")
	yaml.Unmarshal(cfg, &api)

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
				if v.GUID != "" && !strings.HasPrefix(v.GUID, "local://") {
					// log.WithField("server", v.Server.Name).WithField("guid", v.GUID).WithField("title", v.Title).WithField("viewCount", v.ViewCount).WithField("lastViewedAt", v.LastViewedAt).Info("MEDIA")

					videos = append(videos, v)
					media := repo.getMedia(v.GUID)
					if media != nil {
						if media.LastViewedAt != "" {
							log.WithField("server", v.Server.Name).WithField("guid", v.GUID).WithField("title", v.Title).Info("SKIP")
						} else {
							if v.LastViewedAt != "" {
								log.WithField("server", v.Server.Name).WithField("guid", v.GUID).WithField("title", v.Title).Info("UPDATE DB")
								repo.save(v)
							} else {
								log.WithField("server", v.Server.Name).WithField("guid", v.GUID).WithField("title", v.Title).Info("SKIP")
							}
						}
					} else {
						repo.save(v)
						log.WithField("server", v.Server.Name).WithField("guid", v.GUID).WithField("title", v.Title).Info("SAVE")
					}
				}
			default:
				fmt.Printf("Type of o is %T. Value %v", o, o)
			}
		}
	}()

	wg.Wait()

	for _, v := range videos {
		if v.GUID != "" && !strings.HasPrefix(v.GUID, "local://") {
			media := repo.getMedia(v.GUID)
			if media != nil {
				if media.LastViewedAt != "" {
					if v.LastViewedAt == "" {
						v.Server.MarkWatched(v)
						log.WithField("server", v.Server.Name).WithField("guid", v.GUID).WithField("title", v.Title).Info("MARKED WATCHED")
					}
				}
			}
		}
	}
}
