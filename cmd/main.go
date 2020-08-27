package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

func main() {
	fmt.Println("Az VM Size Lister!")

	// create a VirtualMachineSizes client
	vmSizesClient := compute.NewVirtualMachineSizesClient(os.Getenv("SUBSCRIPTION_ID"))

	// create authorizer for client
	authorizer, err := auth.NewAuthorizerFromFile("https://management.azure.com/")
	if err == nil {
		vmSizesClient.Authorizer = authorizer
	}

	vmSizesList, err := vmSizesClient.List(context.Background(), "westus2")

	fmt.Println(err)
	fmt.Println(vmSizesList)
}
