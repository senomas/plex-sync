package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"

	_ "github.com/mattn/go-sqlite3"

	"code.senomas.com/go/plexapi"
	"code.senomas.com/go/util"
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
	util.Panicf("Create table guid %v", err)

	repo.findMedia, err = repo.db.Prepare("select title, addedAt, updatedAt, viewCount, lastViewedAt from media where guid = ?")
	util.Panicf("Prepare findMedia %v", err)

	repo.insertMedia, err = repo.db.Prepare("insert or replace into media(guid, title, addedAt, updatedAt, viewCount, lastViewedAt) values(?, ?, ?, ?, ?, ?)")
	util.Panicf("Prepare insertMedia %v", err)
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
	repo := Repo{}
	repo.init()
	defer repo.close()

	api := plexapi.API{}
	cfg, _ := ioutil.ReadFile("config.yaml")
	yaml.Unmarshal(cfg, &api)

	var wg sync.WaitGroup
	out := make(chan plexapi.Video)

	servers, err := api.GetServers()
	util.Panicf("GetServers failed %v", err)
	for _, server := range servers {
		server.GetVids(&wg, out)
	}

	var videos []plexapi.Video
	go func() {
		for v := range out {
			if v.GUID != "" && !strings.HasPrefix(v.GUID, "local://") {
				videos = append(videos, v)
				media := repo.getMedia(v.GUID)
				if media != nil {
					if media.LastViewedAt != "" {
						fmt.Printf("MEDIA %s %s %s SKIP\n", v.Server.Name, v.GUID, v.Title)
					} else {
						if v.LastViewedAt != "" {
							fmt.Printf("MEDIA %s %s %s UPDATE DB\n", v.Server.Name, v.GUID, v.Title)
							repo.save(v)
						} else {
							fmt.Printf("MEDIA %s %s %s SKIP\n", v.Server.Name, v.GUID, v.Title)
						}
					}
				} else {
					repo.save(v)
					fmt.Printf("MEDIA %s %s %s SAVE\n", v.Server.Name, v.GUID, v.Title)
				}
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
						fmt.Printf("MEDIA %s %s %s MARKED WATCHED\n", v.Server.Name, v.GUID, v.Title)
					}
				}
			}
		}
	}

	fmt.Println("DONE")
}
