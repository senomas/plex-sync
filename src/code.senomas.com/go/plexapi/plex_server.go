package plexapi

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"code.senomas.com/go/util"
	log "github.com/Sirupsen/logrus"
)

// Server struct
type Server struct {
	api               *API
	host              string
	XMLName           xml.Name `xml:"Server"`
	AccessToken       string   `xml:"accessToken,attr"`
	Name              string   `xml:"name,attr"`
	Address           string   `xml:"address,attr"`
	Port              int      `xml:"port,attr"`
	Version           string   `xml:"version,attr"`
	Scheme            string   `xml:"scheme,attr"`
	Host              string   `xml:"host,attr"`
	LocalAddresses    string   `xml:"localAddresses,attr"`
	MachineIdentifier string   `xml:"machineIdentifier,attr"`
	CreatedAt         int      `xml:"createdAt,attr"`
	UpdatedAt         int      `xml:"updatedAt,attr"`
	Owned             int      `xml:"owned,attr"`
	Synced            int      `xml:"synced,attr"`
}

func (server *Server) setHeader(req *http.Request) {
	req.Header.Add("X-Plex-Product", "plex-sync")
	req.Header.Add("X-Plex-Version", "1.0.0")
	req.Header.Add("X-Plex-Client-Identifier", "donkey")
	if server.api.userInfo.Token != "" {
		req.Header.Add("X-Plex-Token", server.api.userInfo.Token)
	}
}

func (server *Server) getContainer(url string) (container MediaContainer, err error) {
	log.WithField("url", url).Debugf("GET")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.WithField("url", url).WithError(err).Errorf("http.GET")
		return container, err
	}
	server.setHeader(req)

	resp, err := server.api.client.Do(req)
	if err != nil {
		return container, err
	}
	defer resp.Body.Close()

	log.WithField("url", url).WithField("status", resp.Status).Debugf("RESP")

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.WithField("url", url).WithError(err).Errorf("RESP")
		return container, err
	}

	err = xml.Unmarshal(body, &container)
	return container, err
}

// Check func
func (server *Server) Check() {
	var urls []string
	urls = append(urls, fmt.Sprintf("https://%s:32400", server.Address))
	urls = append(urls, fmt.Sprintf("https://%s:32400", server.Host))
	for _, la := range strings.Split(server.LocalAddresses, ",") {
		urls = append(urls, fmt.Sprintf("https://%s:32400", la))
	}
	log.Debugf("CHECK %s", util.JSONPrettyPrint(urls))

	var checkOnce util.CheckOnce
	var wg sync.WaitGroup

	out := make(chan string)
	defer close(out)

	for _, pu := range urls {
		url := pu
		wg.Add(1)
		go func() error {
			defer wg.Done()

			if !checkOnce.IsDone() {
				log.Debugf("CHECK %s %s", server.Name, url)
				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					log.Debugf("ERROR GET %s '%s': %v", server.Name, url, err)
					return err
				}
				server.setHeader(req)
				resp, err := server.api.client.Do(req)
				if err != nil {
					log.Debugf("ERROR GET %s '%s': %v", server.Name, url, err)
					return err
				}
				defer resp.Body.Close()

				log.Debugf("RESP %s '%s' %s", server.Name, url, resp.Status)

				if resp.StatusCode == 200 {
					checkOnce.Done(func() {
						out <- url
					})
				}
			}
			return nil
		}()
	}

	go func() {
		wg.Wait()
		checkOnce.Done(func() {
			out <- ""
		})
	}()

	server.host = <-out
	log.Debugf("URL %s  %s", server.Name, server.host)
}

// Perform func
func (server *Server) Perform(cmd, path string) error {
	if server.host != "" {
		url := server.host + path
		log.WithField("url", url).Debugf(cmd)
		req, err := http.NewRequest(cmd, url, nil)
		if err != nil {
			log.WithField("url", url).WithError(err).Errorf("http.%s", cmd)
			return err
		}
		server.setHeader(req)

		resp, err := server.api.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		log.WithField("url", url).WithField("status", resp.Status).Debugf("RESP")

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.WithField("url", url).WithError(err).Errorf("RESP")
			return err
		}

		log.Debugf("RESP BODY\n%s", body)
		return err
	}
	return fmt.Errorf("NO HOST")
}

// GetContainer func
func (server *Server) GetContainer(path string) (container MediaContainer, err error) {
	if server.host != "" {
		container, err = server.getContainer(server.host + path)
		return container, err
	}
	return container, fmt.Errorf("NO HOST")
}

// GetDirectories func
func (server *Server) GetDirectories() (directories []Directory, err error) {
	container, err := server.GetContainer("/library/sections")
	if err != nil {
		return directories, err
	}
	for _, d := range container.Directories {
		nd := d
		// nd.server = server
		directories = append(directories, nd)
	}
	return directories, err
}

// GetVideos func
func (server *Server) GetVideos(wg *sync.WaitGroup, out chan<- interface{}) {
	cs := make(chan MediaContainer)
	dirs := make(chan Directory)

	wg.Add(1)
	go func() {
		container, err := server.GetContainer("/library/sections")
		if err == nil {
			cs <- container
		} else {
			wg.Done()
		}
	}()
	for i, il := 0, server.api.HTTP.WorkerSize; i < il; i++ {
		go func() {
			for c := range cs {
				func() {
					defer wg.Done()
					for _, d := range c.Directories {
						if d.Type == "movie" {
							wg.Add(1)
							d.Paths = c.Paths
							d.Keys = c.Keys
							d.Keys = append(d.Keys, KeyInfo{Key: d.Key, Type: d.Type, Title: d.Title})
							dirs <- d
						}
					}
					for _, v := range c.Videos {
						// if v.GUID == "" {
						// 	meta, err := server.GetMeta(v)
						// 	if err != nil {
						// 		log.Fatal("GetMeta ", err)
						// 	}
						// 	v = meta
						// }
						v.server = server
						v.Paths = c.Paths
						v.Keys = c.Keys
						var idx []string
						for _, px := range v.Media.Parts {
							for _, kk := range v.Paths {
								if strings.HasPrefix(px.File, kk) {
									idx = append(idx, px.File[len(kk):])
								}
							}
						}
						if len(idx) > 0 {
							v.FID = strings.Join(idx, ":")
							v.FID = strings.Replace(v.FID, "\\", "/", -1)
						} else {
							v.FID = v.GUID
						}
						wg.Add(1)
						out <- v
					}
				}()
			}
		}()
		go func() {
			for d := range dirs {
				func() {
					defer wg.Done()

					if strings.HasPrefix(d.Key, "/library/") {
						cc, err := server.GetContainer(d.Key)
						if err == nil {
							wg.Add(1)
							cc.Keys = d.Keys
							cc.Paths = d.Paths
							for _, l := range d.Locations {
								if l.Path != "" {
									cc.Paths = append(cc.Paths, l.Path)
								}
							}
							cs <- cc
						}
					} else {
						cc, err := server.GetContainer(fmt.Sprintf("/library/sections/%v/all", d.Key))
						if err == nil {
							wg.Add(1)
							cc.Keys = d.Keys
							cc.Paths = d.Paths
							for _, l := range d.Locations {
								if l.Path != "" {
									cc.Paths = append(cc.Paths, l.Path)
								}
							}
							cs <- cc
						}
					}
				}()
			}
		}()
	}
}

// GetMeta func
func (server *Server) GetMeta(video Video) (meta Video, err error) {
	if video.GUID != "" {
		return video, nil
	}
	var mc MediaContainer
	mc, err = server.GetContainer(video.Key)
	if err != nil {
		return video, err
	}
	return mc.Videos[0], nil
}

// MarkWatched func
func (server *Server) MarkWatched(key string) error {
	url := server.host + "/:/scrobble?identifier=com.plexapp.plugins.library&key=" + key
	log.WithField("url", url).WithField("server", server.Name).Debug("MarkWatched.GET")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	server.setHeader(req)

	resp, err := server.api.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.WithField("url", url).WithField("server", server.Name).Debugf("MarkWatched.RESULT\n%s", body)

	return err
}

// MarkUnwatched func
func (server *Server) MarkUnwatched(key string) error {
	url := server.host + "/:/unscrobble?identifier=com.plexapp.plugins.library&key=" + key
	log.WithField("url", url).WithField("server", server.Name).Debug("MarkUnwatched.GET")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	server.setHeader(req)

	resp, err := server.api.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.WithField("url", url).WithField("server", server.Name).Debugf("MarkUnwatched.RESULT\n%s", body)

	return err
}

// SetViewOffset func
func (server *Server) SetViewOffset(key, offset string) error {
	url := server.host + "/:/progress?key=" + key + "&identifier=com.plexapp.plugins.library&time=" + offset
	log.WithField("url", url).WithField("server", server.Name).Debug("SetViewOffset.GET")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	server.setHeader(req)

	resp, err := server.api.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.WithField("url", url).WithField("server", server.Name).Debugf("SetViewOffset.RESULT\n%s", body)

	return err
}
