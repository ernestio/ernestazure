/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package vritualnetwork

import (
	"errors"
	"sync"

	"github.com/ernestio/ernestazure"

	"github.com/Azure/azure-sdk-for-go/management"
	"github.com/Azure/azure-sdk-for-go/management/networksecuritygroup"
	"github.com/Azure/azure-sdk-for-go/management/virtualnetwork"
)

// Event : ...
type Event struct {
	ernestazure.Base
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	AddressSpaces  []string `json:"address_spaces"`
	DNSServerNames []string `json:"dns_server_names"`
	Subnets        []subnet `json:"subnets"`
	Location       string   `json:"location"`
	ErrorMessage   string   `json:"error,omitempty"`
	Subject        string   `json:"-"`
	Body           []byte   `json:"-"`
	CryptoKey      string   `json:"-"`
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

// Validate checks if all criteria are met
func (ev *Event) Validate() error {
	return nil
}

// Find : Find an object on azure
func (ev *Event) Find() error {
	return errors.New(ev.Subject + " not supported")
}

func (ev *Event) client() (management.Client, error) {
	subscriptionID := "my subscription id"
	settings := []byte("bla")

	return management.ClientFromPublishSettingsData(settings, subscriptionID)
}

// Create : Creates a nat object on azure
func (ev *Event) Create() error {
	mt := sync.Mutex{}

	mc, err := ev.client()
	vnetClient := virtualnetwork.NewClient(mc)

	name := "Input name"

	// Lock the client just before we get the virtual network configuration and immediately
	// set an defer to unlock the client again whenever this function exits
	mt.Lock()
	defer mt.Unlock()

	nc, err := vnetClient.GetVirtualNetworkConfiguration()
	if err != nil {
		if management.IsResourceNotFoundError(err) {
			// if no network config exists yet; create a new one now:
			nc = virtualnetwork.NetworkConfiguration{}
		} else {
			return errors.New("Error retrieving Virtual Network Configuration: " + err.Error())
		}
	}

	for _, n := range nc.Configuration.VirtualNetworkSites {
		if n.Name == name {
			return errors.New("Virtual Network " + name + " already exists")
		}
	}

	network := ev.createVirtualNetwork()
	nc.Configuration.VirtualNetworkSites = append(nc.Configuration.VirtualNetworkSites, network)

	req, err := vnetClient.SetVirtualNetworkConfiguration(nc)
	if err != nil {
		return errors.New("Error creating Virtual Network " + name + ":" + err.Error())
	}

	// Wait until the virtual network is created
	if err := mc.WaitForOperation(req, nil); err != nil {
		return errors.New("Error waiting for Virtual Network " + name + " to be created: " + err.Error())
	}

	ev.ID = name

	if err := ev.associateSecurityGroups(); err != nil {
		return err
	}

	return ev.Get()
}

// Update : Updates a nat object on azure
func (ev *Event) Update() error {
	return errors.New(ev.Subject + " not supported")
}

// Delete : Deletes a nat object on azure
func (ev *Event) Delete() error {
	return errors.New(ev.Subject + " not supported")
}

// Get : Gets a nat object on azure
func (ev *Event) Get() error {
	return errors.New(ev.Subject + " not supported")
}

func (ev *Event) createVirtualNetwork() virtualnetwork.VirtualNetworkSite {
	// fetch address spaces:
	var dnsRefs []virtualnetwork.DNSServerRef
	var subnets []virtualnetwork.Subnet

	// fetch DNS references:
	for _, dns := range ev.DNSServerNames {
		dnsRefs = append(dnsRefs, virtualnetwork.DNSServerRef{
			Name: dns,
		})
	}

	// Add all subnets that are configured
	for _, subnet := range ev.Subnets {
		subnets = append(subnets, virtualnetwork.Subnet{
			Name:          subnet.Name,
			AddressPrefix: subnet.AddressPrefix,
		})
	}

	return virtualnetwork.VirtualNetworkSite{
		Name:     ev.Name,
		Location: ev.Location,
		AddressSpace: virtualnetwork.AddressSpace{
			AddressPrefix: ev.AddressSpaces,
		},
		DNSServersRef: dnsRefs,
		Subnets:       subnets,
	}
}

func (ev *Event) associateSecurityGroups() error {
	mc, _ := ev.client()
	secGroupClient := networksecuritygroup.NewClient(mc)

	for _, subnet := range ev.Subnets {
		securityGroup := subnet.SecurityGroup
		subnetName := subnet.Name

		// Get the associated (if any) security group
		sg, err := secGroupClient.GetNetworkSecurityGroupForSubnet(subnetName, ev.ID)
		if err != nil && !management.IsResourceNotFoundError(err) {
			return errors.New("Error retrieving Network Security Group associations of subnet " + subnetName + ": " + err.Error())
		}

		// If the desired and actual security group are the same, were done so can just continue
		if sg.Name == securityGroup {
			continue
		}

		// If there is an associated security group, make sure we first remove it from the subnet
		if sg.Name != "" {
			req, err := secGroupClient.RemoveNetworkSecurityGroupFromSubnet(sg.Name, subnetName, ev.Name)
			if err != nil {
				return errors.New("Error removing Network Security Group " + securityGroup + " from subnet " + subnetName + ": " + err.Error())
			}

			// Wait until the security group is associated
			if err := mc.WaitForOperation(req, nil); err != nil {
				return errors.New("Error waiting for Network Security Group " + securityGroup + " to be removed from subnet " + subnetName + ": " + err.Error())
			}
		}

		// If the desired security group is not empty, assign the security group to the subnet
		if securityGroup != "" {
			req, err := secGroupClient.AddNetworkSecurityToSubnet(securityGroup, subnetName, ev.Name)
			if err != nil {
				return errors.New("Error associating Network Security Group " + securityGroup + " to subnet " + subnetName + ": " + err.Error())
			}

			// Wait until the security group is associated
			if err := mc.WaitForOperation(req, nil); err != nil {
				return errors.New("Error waiting for Network Security Group " + securityGroup + " to be associated with subnet : " + subnetName)
			}
		}

	}

	return nil
}
