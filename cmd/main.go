package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/hattan/az-vm-region-lister/pkg/models"
)

var subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
var all models.VmSizes

func getVMSizes(region string, c chan models.VmSizes, onExit func()) {
	go func() {
		defer onExit()
		fmt.Printf("Processing: %s", region)
		vmSizesClient := compute.NewVirtualMachineSizesClient(subscriptionID)

		authorizer, err := auth.NewAuthorizerFromEnvironment()
		if err == nil {
			vmSizesClient.Authorizer = authorizer
		}

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

func getLocations() []string {
	jsonFile, err := os.Open("../regions.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	fmt.Println("Successfully Opened regions.json")
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var regions []string
	json.Unmarshal(byteValue, &regions)
	return regions

	// // Create an Azure management client
	// azureClient := management.NewAnonymousClient()

	// // Create a location client from the management client
	// locationClient := location.newClient(azureClient)

	// // Get list of locations
	// locationResponse, err := locationClient.ListLications()
	// fmt.Println(err)
	// fmt.Println(locationResponse)
}

func main() {
	getLocations()
	var wg sync.WaitGroup
	locations := getLocations()
	c := make(chan models.VmSizes)

	for _, location := range locations {
		wg.Add(1)
		getVMSizes(location, c, func() { wg.Done() })
	}

	go func() {
		defer close(c)
		wg.Wait()
	}()

	for sizes := range c {
		if sizes != nil && len(sizes) > 0 {
			all = append(all, sizes...)
		}
	}

	file, _ := json.MarshalIndent(all, "", " ")
	_ = ioutil.WriteFile("test.json", file, 0644)

	fmt.Println("Complete")
}
