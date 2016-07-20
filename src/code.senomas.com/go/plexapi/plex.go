package plexapi

import (
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"
)

// API struct
type API struct {
	User     string     `yaml:"user"`
	Password string     `yaml:"password"`
	HTTP     HTTPConfig `yaml:"http"`
	client   *http.Client
	userInfo UserInfo
	servers  map[string]*Server
}

// HTTPConfig struct
type HTTPConfig struct {
	Timeout    time.Duration `yaml:"timeout"`
	WorkerSize int           `yaml:"workerSize"`
}

// UserInfo struct
type UserInfo struct {
	XMLName  xml.Name `xml:"user"`
	UserName string   `xml:"username"`
	Token    string   `xml:"authentication-token"`
}

// MediaContainer struct
type MediaContainer struct {
	XMLName                       xml.Name    `xml:"MediaContainer"`
	Servers                       []Server    `xml:"Server"`
	Directories                   []Directory `xml:"Directory"`
	Videos                        []Video     `xml:"Video"`
	Size                          int         `xml:"size,attr"`
	AllowCameraUpload             int         `xml:"allowCameraUpload,attr"`
	AllowSync                     int         `xml:"allowSync,attr"`
	AllowChannelAccess            int         `xml:"allowChannelAccess,attr"`
	RequestParametersInCookie     int         `xml:"requestParametersInCookie,attr"`
	Sync                          int         `xml:"sync,attr"`
	TranscoderActiveVideoSessions int         `xml:"transcoderActiveVideoSessions,attr"`
	TranscoderAudio               int         `xml:"transcoderAudio,attr"`
	TranscoderVideo               int         `xml:"transcoderVideo,attr"`
	TranscoderVideoBitrates       string      `xml:"transcoderVideoBitrates,attr"`
	TranscoderVideoQualities      string      `xml:"transcoderVideoQualities,attr"`
	TranscoderVideoResolutions    string      `xml:"transcoderVideoResolutions,attr"`
	FriendlyName                  string      `xml:"friendlyName,attr"`
	MachineIdentifier             string      `xml:"machineIdentifier,attr"`
}

// Progress struct
type Progress struct {
	Server  string
	Command string
	Delta   int
}

func (api *API) setHeader(req *http.Request) {
	req.Header.Add("X-Plex-Product", "plex-sync")
	req.Header.Add("X-Plex-Version", "1.0.0")
	req.Header.Add("X-Plex-Client-Identifier", "donkey")
	if api.userInfo.Token != "" {
		req.Header.Add("X-Plex-Token", api.userInfo.Token)
	}
}

// Login func
func (api *API) login() error {
	if api.client == nil {
		cookieJar, _ := cookiejar.New(nil)
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		api.client = &http.Client{
			Timeout:   time.Duration(time.Second * api.HTTP.Timeout),
			Transport: tr,
			Jar:       cookieJar,
		}
	}

	reqBody := fmt.Sprintf("user[login]=%s&user[password]=%s", api.User, api.Password)
	req, err := http.NewRequest("POST", "https://plex.tv/users/sign_in.xml", strings.NewReader(reqBody))
	api.setHeader(req)
	if err != nil {
		return err
	}

	resp, err := api.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return xml.Unmarshal(body, &api.userInfo)
}

// GetServers func
func (api *API) GetServers() (servers map[string]*Server, err error) {
	if api.userInfo.Token == "" {
		err := api.login()
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest("GET", "https://plex.tv/pms/servers.xml", nil)
	if err != nil {
		return nil, err
	}
	api.setHeader(req)

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data MediaContainer
	err = xml.Unmarshal(body, &data)
	api.servers = make(map[string]*Server)
	for _, s := range data.Servers {
		ns := s
		ns.api = api
		api.servers[s.Name] = &ns
	}
	return api.servers, err
}

// GetServer func
func (api *API) GetServer(name string) (server *Server, err error) {
	if api.servers == nil {
		_, err := api.GetServers()
		if err != nil {
			return nil, err
		}
	}

	s, ok := api.servers[name]
	if ok {
		return s, nil
	}
	return nil, fmt.Errorf("Server '%s' not found", name)
}
