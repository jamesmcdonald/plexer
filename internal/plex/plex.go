package plex

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Plex struct {
	Endpoint string
	Token    string
}

func New(endpoint string, token string) *Plex {
	return &Plex{
		Endpoint: endpoint,
		Token:    token,
	}
}

func (p *Plex) query(target string) (*PlexResponse, error) {
	url := p.Endpoint + target + "?X-Plex-Token=" + p.Token
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transCfg}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get response: %s", res.Status)
	}
	var plexResponse PlexResponse
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &plexResponse)
	if err != nil {
		return nil, err
	}
	return &plexResponse, nil

}

func (p *Plex) GetLibraries() ([]string, error) {
	res, err := p.query("/library/sections")
	if err != nil {
		return nil, err
	}
	libraries := make([]string, len(res.MediaContainer.Directory))
	for i, library := range res.MediaContainer.Directory {
		libraries[i] = library.Title
	}
	return libraries, nil
}
