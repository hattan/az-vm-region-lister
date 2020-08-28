package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/hattan/az-vm-region-lister/pkg/models"
)

var subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")

func getVMSizes(region string) (models.VmSizes, error) {
	vmSizesClient := compute.NewVirtualMachineSizesClient(subscriptionID)

	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err == nil {
		vmSizesClient.Authorizer = authorizer
	}

	var vmSizes models.VmSizes
	vmSizesList, err := vmSizesClient.List(context.Background(), region)
	if err != nil {
		return nil, err
	}
	for _, vm := range *vmSizesList.Value {
		size := models.VmSize{
			Name:     *vm.Name,
			Location: region,
		}
		vmSizes = append(vmSizes, size)
	}
	return vmSizes, nil
}

func getLocations() []string {
	return []string{
		"eastus",
		"eastus2",
		"southcentralus",
		"westus2",
	}
}
func main() {
	locations := getLocations()
	var all models.VmSizes
	for _, location := range locations {
		sizes, err := getVMSizes(location)
		if err != nil {
			fmt.Println(err)
		} else {
			if sizes != nil && len(sizes) > 0 {
				all = append(all, sizes...)
			}
		}
	}

	fmt.Println(all)

}
