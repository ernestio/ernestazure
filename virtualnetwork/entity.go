/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package vritualnetwork

import (
	"errors"
	"log"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/arm/network"
	"github.com/ernestio/ernestazure"
	"github.com/ernestio/ernestazure/credentials"
)

// Event : ...
type Event struct {
	ernestazure.Base
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	AddressSpace   []string `json:"address_space"`
	DNSServerNames []string `json:"dns_server_names"`
	Subnets        []subnet `json:"subnets"`
	Location       string   `json:"location"`
	ClientID       string   `json:"azure_subscription_id"`
	ClientSecret   string   `json:"azure_subscription_id"`
	TenantID       string   `json:"azure_subscription_id"`
	SubscriptionID string   `json:"azure_subscription_id"`

	ResourceGroupName string             `json:"resource_group_name"`
	Tags              map[string]*string `json:"tags"`

	ErrorMessage string `json:"error,omitempty"`
	Subject      string `json:"-"`
	Body         []byte `json:"-"`
	CryptoKey    string `json:"-"`
}

type subnet struct {
	Name          string `json:"name"`
	AddressPrefix string `json:"address_prefix"`
	SecurityGroup string `json:"security_group"`
}

// New : Constructor
func New(subject string, body []byte, cryptoKey string) ernestazure.Event {
	n := Event{Subject: subject, Body: body, CryptoKey: cryptoKey}

	return &n
}

// Azure virtual network client
func (ev *Event) client() *network.VirtualNetworksClient {
	client, _ := credentials.Client(ev.ClientID, ev.ClientSecret, ev.TenantID, ev.SubscriptionID, ev.CryptoKey)
	return &client.VirtualNetworkClient
}

// Validate checks if all criteria are met
func (ev *Event) Validate() error {
	// TOOD : Add validation rules
	return nil
}

// Find : Find an object on azure
func (ev *Event) Find() error {
	return errors.New(ev.Subject + " not supported")
}

// Create : Creates a nat object on azure
func (ev *Event) Create() error {
	c := ev.client()

	log.Printf("[INFO] preparing arguments for Azure ARM virtual network creation.")

	resGroup := ev.ResourceGroupName

	vnet := network.VirtualNetwork{
		Name:     &ev.Name,
		Location: &ev.Location,
		Tags:     &ev.Tags,
	}

	_, err := c.CreateOrUpdate(resGroup, ev.Name, vnet, make(chan struct{}))
	if err != nil {
		return err
	}

	read, err := c.Get(resGroup, ev.Name, "")
	if err != nil {
		return err
	}
	if read.ID == nil {
		return errors.New("Cannot read Virtual Network " + ev.Name + " (resource group " + resGroup + ") ID")
	}

	ev.ID = *read.ID

	return ev.Get()
}

// Update : Updates a nat object on azure
func (ev *Event) Update() error {
	return ev.Create()
}

// Delete : Deletes a nat object on azure
func (ev *Event) Delete() (err error) {
	resGroup := ev.ResourceGroupName
	name := ev.Name

	_, err = ev.client().Delete(resGroup, name, make(chan struct{}))

	return err
}

// Get : Gets a nat object on azure
func (ev *Event) Get() error {
	resGroup := ev.ResourceGroupName
	name := ev.Name

	resp, err := ev.client().Get(resGroup, name, "")
	if err != nil {
		return errors.New("Error making Read request on Azure virtual network " + name + ": " + err.Error())
	}
	if resp.StatusCode == http.StatusNotFound {
		ev.ID = ""
		return nil
	}

	// update appropriate values
	ev.Name = *resp.Name
	ev.Location = *resp.Location
	ev.AddressSpace = *resp.AddressSpace.AddressPrefixes
	// ev.Subnets = *resp.Subnets

	dnses := []string{}
	for _, dns := range *resp.DhcpOptions.DNSServers {
		dnses = append(dnses, dns)
	}
	ev.DNSServerNames = dnses
	ev.Tags = *resp.Tags

	return nil
}
