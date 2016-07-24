package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"code.senomas.com/go/plexapi"
	"code.senomas.com/go/util"
	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
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
	log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	db, err := bolt.Open("./plex.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	api := plexapi.API{HTTP: plexapi.HTTPConfig{Timeout: 30, WorkerSize: 10}}
	api.LoadConfig("config.yaml")

	now := time.Now().Unix()

	tx, err := db.Begin(true)
	if err != nil {
		log.Fatal("DB failed ", err)
	}
	defer tx.Rollback()

	var wg sync.WaitGroup
	out := make(chan interface{})
	counts := make(map[string]int)

	servers, err := api.GetServers()
	util.Panicf("GetServers failed %v", err)
	for _, server := range servers {
		server.GetVideos(&wg, out)
		counts["Video '"+server.Name+"'"] = 0
	}

	bvid, err := tx.CreateBucketIfNotExists([]byte("Media"))
	if err != nil {
		log.Fatal("NO Bucket ", err)
	}

	var videos []plexapi.Video
	go func() {
		for o := range out {
			func() {
				defer wg.Done()

				switch o := o.(type) {
				case plexapi.Video:
					v := plexapi.Video(o)
					vsn := v.GetServer().Name
					if v.FID != "" && !strings.HasPrefix(v.FID, "local://") {
						counts["Video '"+vsn+"'"]++
						// log.WithField("server", v.GetServer().Name).WithField("guid", v.GUID).WithField("title", v.Title).WithField("viewCount", v.ViewCount).WithField("lastViewedAt", v.LastViewedAt).Info("MEDIA")
						videos = append(videos, v)
						data := &plexapi.Data{
							Videos:    make(map[string]plexapi.Video),
							UpdatedAt: make(map[string]int64),
						}

						bb := bvid.Get([]byte(v.FID))
						if bb != nil {
							json.Unmarshal(bb, data)
						} else {
							log.Warn("NO DATA ", v.FID)
						}

						vx, ok := data.Videos[vsn]
						update := true
						var vnow int64
						if ok {
							bx1, err := json.Marshal(vx)
							if err != nil {
								log.Fatal("Marshal ", err)
							}
							bx2, err := json.Marshal(v)
							if err != nil {
								log.Fatal("Marshal ", err)
							}
							if bytes.Equal(bx1, bx2) {
								update = false
							} else if v.LastViewedAt != vx.LastViewedAt || v.ViewOffset != vx.ViewOffset {
								vnow = now
							} else {
								if vn, ok := data.UpdatedAt[vsn]; ok {
									vnow = vn
								} else {
									log.Fatal("NO PREVIOUS UPDATED???")
								}
							}
						}
						if update {
							data.Videos[vsn] = v
							data.UpdatedAt[vsn] = vnow
							bb, err = json.Marshal(data)
							if err != nil {
								log.Fatal("Marshal ", err)
							}
							err := bvid.Put([]byte(v.FID), bb)
							if err != nil {
								log.Fatal("Bucket put ", err)
							}
							log.Infof("UPDATE '%s'   %v   %s", v.GetServer().Name, len(data.Videos), v.FID)
						} else {
							log.Infof("SKIP '%s'   %s", v.GetServer().Name, v.FID)
						}
					}
				default:
					fmt.Printf("Type of o is %T. Value %v", o, o)
				}
			}()
		}
	}()

	wg.Wait()
	fmt.Print("\n\n\n\n")

	for _, v := range videos {
		if v.FID != "" && !strings.HasPrefix(v.FID, "local://") {
			vsn := v.GetServer().Name

			data := &plexapi.Data{
				Videos:    make(map[string]plexapi.Video),
				UpdatedAt: make(map[string]int64),
			}

			bb := bvid.Get([]byte(v.FID))
			if bb == nil {
				log.Fatal("No data ", v.FID)
			}
			json.Unmarshal(bb, data)

			sn := vsn
			su := data.UpdatedAt[vsn]
			for kn, ku := range data.UpdatedAt {
				if ku > su {
					sn = kn
					su = ku
				} else if ku == su {
					kv := data.Videos[kn]
					kvi := atoi(kv.LastViewedAt)
					svi := atoi(v.LastViewedAt)
					if kvi > svi {
						sn = kn
						su = ku
					} else if kvi == svi && atoi(kv.ViewCount) > atoi(v.ViewCount) {
						sn = kn
						su = ku
					}
				}
			}

			if vsn != sn {
				update := false

				vn := data.Videos[sn]
				// log.Info("DATA FINAL ", sn, "  ", vn.FID, "  LV ", atoi(vn.LastViewedAt), "  VC ", atoi(vn.ViewCount), "   VO ", vn.ViewOffset)
				if atoi(vn.ViewCount) > 0 {
					if atoi(v.ViewCount) == 0 {
						v.GetServer().MarkWatched(v.RatingKey)
						log.Infof("MARK WATCHED '%s' <<%s>> %s", vsn, sn, v.FID)
						update = true
					}
				} else if atoi(vn.ViewOffset) > 0 {
					if atoi(v.ViewCount) > 0 {
						v.GetServer().MarkUnwatched(v.RatingKey)
						log.Infof("MARK UNWATCHED '%s' <<%s>> %s", vsn, sn, v.FID)
						update = true
					}
					if v.ViewOffset != vn.ViewOffset && vn.ViewOffset != "" {
						v.GetServer().SetViewOffset(v.RatingKey, vn.ViewOffset)
						log.Infof("UPDATE VIEW-OFFSET '%s' <<%s>> %s", vsn, sn, v.FID)
						update = true
					}
				} else {
					if atoi(v.ViewCount) > 0 {
						v.GetServer().MarkUnwatched(v.RatingKey)
						log.Infof("MARK UNWATCHED '%s' <<%s>> %s", vsn, sn, v.FID)
						update = true
					}
				}

				if update {
					data.UpdatedAt[vsn] = data.UpdatedAt[sn] - 1
					bb, err = json.Marshal(data)
					if err != nil {
						log.Fatal("Marshal ", err)
					}
					err := bvid.Put([]byte(v.FID), bb)
					if err != nil {
						log.Fatal("Bucket put ", err)
					}
				}
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal("DB Commit ", err)
	}

	fmt.Println("COUNTS: ", util.JSONPrettyPrint(counts))
}
