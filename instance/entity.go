/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package vritualnetwork

import (
	"errors"
	"sync"

	"github.com/ernestio/ernestazure"
	"github.com/ernestio/ernestazure/credentials"

	"github.com/Azure/azure-sdk-for-go/management"
	"github.com/Azure/azure-sdk-for-go/management/networksecuritygroup"
	"github.com/Azure/azure-sdk-for-go/management/virtualnetwork"
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
	SubscriptionID string   `json:"azure_subscription_id"`
	Settings       string   `json:"azure_settings"`

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
