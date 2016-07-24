package plexapi

import (
	"database/sql"
	"fmt"

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

// Open func
func (repo *Repo) Open() (err error) {

	repo.db, err = sql.Open("sqlite3", "./plex.db")
	util.Panicf("Open DB %v", err)

	_, err = repo.db.Exec("create table if not exists media(id text primary key, guid text, title text, addedAt int, updatedAt int, viewCount int, viewOffset int, lastViewedAt int)")
	if err != nil {
		return fmt.Errorf("Failed to create table media %v", err)
	}

	repo.findMedia, err = repo.db.Prepare("select guid, title, addedAt, updatedAt, viewCount, viewOffset, lastViewedAt from media where id = ?")
	if err != nil {
		return err
	}

	repo.insertMedia, err = repo.db.Prepare("insert or replace into media(id, guid, title, addedAt, updatedAt, viewCount, viewOffset, lastViewedAt) values(?, ?, ?, ?, ?, ?, ?, ?)")
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

	vid = &Video{FID: id}
	err = rows.Scan(vid.GUID, vid.Title, vid.AddedAt, vid.UpdatedAt, vid.ViewCount, vid.ViewOffset, vid.LastViewedAt)
	return vid, err
}

// Save func
func (repo *Repo) Save(v *Video) error {
	_, err := repo.insertMedia.Exec(v.FID, v.GUID, v.Title, v.AddedAt, v.UpdatedAt, v.ViewCount, v.ViewOffset, v.LastViewedAt)
	return err
}

// Close func
func (repo *Repo) Close() {
	repo.findMedia.Close()
	repo.insertMedia.Close()
	repo.db.Close()
}
