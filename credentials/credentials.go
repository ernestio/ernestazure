package credentials

import (
	"errors"
	"log"

	"github.com/Azure/azure-sdk-for-go/arm/network"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"

	aes "github.com/ernestio/crypto/aes"
)

// AzureClient : ---
type AzureClient struct {
	VirtualNetworkClient network.VirtualNetworksClient
}

// Client : Get an Azure client based on encrypted data
func Client(clientID, clientSecret, tenantID, subscriptionID, cryptoKey string) (client *AzureClient, err error) {
	if cryptoKey != "" {
		crypto := aes.New()
		if clientID, err = crypto.Decrypt(clientID, cryptoKey); err != nil {
			log.Println(err.Error())
			return nil, err
		}
		if clientSecret, err = crypto.Decrypt(clientSecret, cryptoKey); err != nil {
			log.Println(err.Error())
			return nil, err
		}
		if tenantID, err = crypto.Decrypt(tenantID, cryptoKey); err != nil {
			log.Println(err.Error())
			return nil, err
		}
		if subscriptionID, err = crypto.Decrypt(subscriptionID, cryptoKey); err != nil {
			log.Println(err.Error())
			return nil, err
		}
	}

	oauthConfig, err := azure.PublicCloud.OAuthConfigForTenant(tenantID)
	if err != nil {
		return nil, err
	}

	// OAuthConfigForTenant returns a pointer, which can be nil.
	if oauthConfig == nil {
		return nil, errors.New("Unable to configure OAuthConfig for tenant " + tenantID)
	}

	spt, err := azure.NewServicePrincipalToken(*oauthConfig, clientID, clientSecret,
		azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		return nil, err
	}

	vnc := network.NewVirtualNetworksClient(subscriptionID)
	vnc.Authorizer = spt
	vnc.Sender = autorest.CreateSender()
	client.VirtualNetworkClient = vnc

	return client, nil
}
