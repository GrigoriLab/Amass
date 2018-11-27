// Copyright 2017 Jeff Foley. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

package amass

import (
	"fmt"

	"github.com/OWASP/Amass/amass/utils"
)

// Netcraft is the AmassService that handles access to the Netcraft data source.
type Netcraft struct {
	BaseAmassService

	SourceType string
}

// Netcraft returns he object initialized, but not yet started.
func NewNetcraft(e *Enumeration) *Netcraft {
	n := &Netcraft{SourceType: SCRAPE}

	n.BaseAmassService = *NewBaseAmassService(e, "Netcraft", n)
	return n
}

// OnStart implements the AmassService interface
func (n *Netcraft) OnStart() error {
	n.BaseAmassService.OnStart()

	go n.startRootDomains()
	go n.processRequests()
	return nil
}

// OnStop implements the AmassService interface
func (n *Netcraft) OnStop() error {
	n.BaseAmassService.OnStop()
	return nil
}

func (n *Netcraft) processRequests() {
	for {
		select {
		case <-n.PauseChan():
			<-n.ResumeChan()
		case <-n.Quit():
			return
		case <-n.RequestChan():
			// This data source just throws away the checked DNS names
			n.SetActive()
		}
	}
}

func (n *Netcraft) startRootDomains() {
	// Look at each domain provided by the config
	for _, domain := range n.Enum().Config.Domains() {
		n.executeQuery(domain)
	}
}

func (n *Netcraft) executeQuery(domain string) {
	url := n.getURL(domain)
	page, err := utils.RequestWebPage(url, nil, nil, "", "", n.Enum().Proxy)
	if err != nil {
		n.Enum().Log.Printf("%s: %s, %v", n.String(), url, err)
		return
	}

	n.SetActive()
	re := n.Enum().Config.DomainRegex(domain)
	for _, sd := range re.FindAllString(page, -1) {
		n.Enum().NewNameEvent(&AmassRequest{
			Name:   cleanName(sd),
			Domain: domain,
			Tag:    n.SourceType,
			Source: n.String(),
		})
	}
}

func (n *Netcraft) getURL(domain string) string {
	format := "https://searchdns.netcraft.com/?restriction=site+ends+with&host=%s"

	return fmt.Sprintf(format, domain)
}
