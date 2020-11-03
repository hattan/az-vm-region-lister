package services

import (
	"context"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/resources/mgmt/subscriptions"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/hattan/az-vm-region-lister/pkg/models"
)

var subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
var authorizer autorest.Authorizer = nil

func getAuthorizer() autorest.Authorizer {
	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
		fmt.Printf("Authorization Error: %s", err)
		os.Exit(-1)
	}
	return authorizer
}

func GetVMSizes(region string, c chan models.VmSizes, onExit func()) {
	go func() {
		if authorizer == nil {
			authorizer = getAuthorizer()
		}
		defer onExit()
		fmt.Printf("Processing: %s", region)

		// Set up vmSizesClient
		vmSizesClient := compute.NewVirtualMachineSizesClient(subscriptionID)
		vmSizesClient.Authorizer = authorizer

		var vmSizes models.VmSizes
		vmSizesList, err := vmSizesClient.List(context.Background(), region)
		if err != nil {
			fmt.Printf("region:%s error:%s\n", region, err)
		} else {
			for _, vm := range *vmSizesList.Value {
				size := models.VmSize{
					Name:     *vm.Name,
					Location: region,
				}
				vmSizes = append(vmSizes, size)
			}
			fmt.Printf("region:%s complete\n", region)
			c <- vmSizes
		}
	}()
}

func GetLocations() []string {
	if authorizer == nil {
		authorizer = getAuthorizer()
	}

	subscriptionClient := subscriptions.NewClient()
	subscriptionClient.Authorizer = authorizer
	locations, err := subscriptionClient.ListLocations(context.Background(), subscriptionID)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(locations)

	var locationNames []string

	for _, location := range *locations.Value {
		name := location.Name
		locationNames = append(locationNames, *name)
	}

	return locationNames
}
