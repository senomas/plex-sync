package plexapi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"code.senomas.com/go/util"
)

// Repo struct
type Repo struct {
	db               *sql.DB
	findMedia        *sql.Stmt
	insertMedia      *sql.Stmt
	findViewStatus   *sql.Stmt
	insertViewStatus *sql.Stmt
}

// ViewStatus struct
type ViewStatus struct {
	ViewCount    int
	ViewOffset   int64
	LastViewedAt int64
}

// Open func
func (repo *Repo) Open() (err error) {

	repo.db, err = sql.Open("sqlite3", "./plex.db")
	util.Panicf("Open DB %v", err)

	_, err = repo.db.Exec("create table if not exists media(id text primary key, json text)")
	if err != nil {
		return fmt.Errorf("Failed to create table media %v", err)
	}

	_, err = repo.db.Exec("create table if not exists viewstatus(id text primary key, json text)")
	if err != nil {
		return fmt.Errorf("Failed to create table view status %v", err)
	}

	repo.findMedia, err = repo.db.Prepare("select json from media where id = ?")
	if err != nil {
		return err
	}

	repo.insertMedia, err = repo.db.Prepare("insert or replace into media(id, json) values(?, ?)")
	if err != nil {
		return err
	}

	repo.findViewStatus, err = repo.db.Prepare("select json from viewstatus where id = ?")
	if err != nil {
		return err
	}

	repo.insertViewStatus, err = repo.db.Prepare("insert or replace into viewstatus(id, json) values(?, ?)")
	return err
}

// GetMedia func
func (repo *Repo) GetMedia(id string) (vid *Video, err error) {
	rows, err := repo.findMedia.Query(id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	var jdata []byte
	err = rows.Scan(&jdata)
	if err != nil {
		return nil, err
	}
	vid = &Video{}
	err = json.Unmarshal(jdata, vid)
	return vid, err
}

// Save func
func (repo *Repo) Save(v *Video) error {
	jdata, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = repo.insertMedia.Exec(v.FID, jdata)
	if err != nil {
		return err
	}
	jdata, err = json.Marshal(v.GetStatus())
	log.Fatal("SAVE ", string(jdata))
	if err != nil {
		return err
	}
	_, err = repo.insertViewStatus(v.Server.Name+":"+v.FID, jdata)
	return err
}

// GetViewStatus func
func (repo *Repo) GetViewStatus(server string, key string) (vid *Video, err error) {
	rows, err := repo.findViewStatus.Query(server + ":" + key)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	var jdata []byte
	err = rows.Scan(&jdata)
	if err != nil {
		return nil, err
	}
	vs = &ViewStatus{}
	err = json.Unmarshal(jdata, vs)
	return vs, err
}

// Close func
func (repo *Repo) Close() {
	repo.findMedia.Close()
	repo.insertMedia.Close()
	repo.db.Close()
}
