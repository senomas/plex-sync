package main

import (
	"database/sql"
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

	_, err = repo.db.Exec("create table if not exists media(id text primary key, guid text, title text, addedAt int, updatedAt int, viewCount int, viewOffset int, lastViewedAt int)")
	if err != nil {
		log.Fatal("Failed to create table media", err)
	}

	repo.findMedia, err = repo.db.Prepare("select guid, title, addedAt, updatedAt, viewCount, viewOffset, lastViewedAt from media where id = ?")
	if err != nil {
		log.Fatal(err)
	}

	repo.insertMedia, err = repo.db.Prepare("insert or replace into media(id, guid, title, addedAt, updatedAt, viewCount, viewOffset, lastViewedAt) values(?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
}

func (repo *Repo) getMedia(id string) *plexapi.Video {
	rows, err := repo.findMedia.Query(id)
	util.Panicf("getMedia %v", err)
	defer rows.Close()

	if !rows.Next() {
		return nil
	}

	var video plexapi.Video
	video.FID = id
	rows.Scan(&video.GUID, &video.Title, &video.AddedAt, &video.UpdatedAt, &video.ViewCount, &video.ViewOffset, &video.LastViewedAt)
	return &video
}

func (repo *Repo) save(v plexapi.Video) {
	_, err := repo.insertMedia.Exec(v.FID, v.GUID, v.Title, v.AddedAt, v.UpdatedAt, v.ViewCount, v.ViewOffset, v.LastViewedAt)
	util.Panicf("InsertMedia failed %v", err)
}

func (repo *Repo) close() {
	repo.findMedia.Close()
	repo.insertMedia.Close()
	repo.db.Close()
}

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
	log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	repo := Repo{}
	repo.init()
	defer repo.close()

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
					media := repo.getMedia(v.FID)
					if media != nil {
						if atoi(v.UpdatedAt) > atoi(media.UpdatedAt) {
							repo.save(v)
							log.WithField("server", v.Server.Name).WithField("key", v.Key).WithField("id", v.FID).Info("UPDATE")
						} else {
							log.WithField("server", v.Server.Name).WithField("key", v.Key).WithField("id", v.FID).Info("SKIP")
						}
					} else {
						repo.save(v)
						log.WithField("server", v.Server.Name).WithField("key", v.Key).WithField("id", v.FID).Info("SAVE")
					}
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
			media := repo.getMedia(v.FID)
			if media != nil {
				if atoi(v.UpdatedAt) < atoi(media.UpdatedAt) {
					if media.LastViewedAt != "" && v.LastViewedAt == "" {
						log.WithField("server", v.Server.Name).WithField("id", v.FID).WithField("title", v.Title).Info("MARKED WATCHED")
						v.Server.MarkWatched(v)
						// } else if media.LastViewedAt == "" && v.LastViewedAt != "" {
						// 	log.WithField("server", v.Server.Name).WithField("guid", v.GUID).WithField("title", v.Title).Info("MARKED UNWATCHED")
						// 	v.Server.MarkUnwatched(v)
					}
				}
			}
		}
	}
}
