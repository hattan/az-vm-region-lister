package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"sync"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/resources/mgmt/subscriptions"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/hattan/az-vm-region-lister/pkg/models"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

var subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
var all models.VmSizes

func getAuthorizer() autorest.Authorizer {
	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
		fmt.Printf("Authorization Error: %s", err)
		os.Exit(-1)
	}
	return authorizer
}

func getVMSizes(authorizer autorest.Authorizer, region string, c chan models.VmSizes, onExit func()) {
	go func() {
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

func getLocations(authorizer autorest.Authorizer) []string {
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

func saveFileToBlobStore(fileName string) {
	accountName, accountKey := os.Getenv("AZURE_STORAGE_ACCOUNT"), os.Getenv("AZURE_STORAGE_ACCESS_KEY")
	if len(accountName) == 0 || len(accountKey) == 0 {
		log.Fatal("Either the AZURE_STORAGE_ACCOUNT or AZURE_STORAGE_ACCESS_KEY environment variable is not set")
	}
	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		log.Fatal("Invalid credentials with error: " + err.Error())
	}
	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	URL, _ := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net/%s", accountName, "$web"))
	containerURL := azblob.NewContainerURL(*URL, p)
	blobURL := containerURL.NewBlockBlobURL(fileName)
	file, err := os.Open(fileName)
	handleErrors(err)
	fmt.Printf("Uploading the file with blob name: %s\n", fileName)
	_, err = azblob.UploadFileToBlockBlob(context.Background(), file, blobURL, azblob.UploadToBlockBlobOptions{
		BlockSize:   4 * 1024 * 1024,
		Parallelism: 16})
	handleErrors(err)
}

func handleErrors(err error) {
	if err != nil {
		if serr, ok := err.(azblob.StorageError); ok { // This error is a Service-specific
			switch serr.ServiceCode() { // Compare serviceCode to ServiceCodeXxx constants
			case azblob.ServiceCodeContainerAlreadyExists:
				fmt.Println("Received 409. Container already exists")
				return
			}
		}
		log.Fatal(err)
	}
}

func saveVMSizesAsJSON(fileName string, data models.VmSizes) {
	file, _ := json.MarshalIndent(data, "", " ")
	_ = ioutil.WriteFile(fileName, file, 0644)
}

func saveStringArrayAsJSON(fileName string, data []string) {
	file, _ := json.MarshalIndent(data, "", " ")
	_ = ioutil.WriteFile(fileName, file, 0644)
}

func main() {
	// Create authorizer
	authorizer := getAuthorizer()

	var wg sync.WaitGroup
	locations := getLocations(authorizer)

	c := make(chan models.VmSizes)

	for _, location := range locations {
		wg.Add(1)
		getVMSizes(authorizer, location, c, func() { wg.Done() })
	}

	go func() {
		defer close(c)
		wg.Wait()
	}()

	var virtualMachineSizes []string
	for sizes := range c {
		if sizes != nil && len(sizes) > 0 {
			for _, size := range sizes {
				if !stringInSlice(size.Name, virtualMachineSizes) {
					virtualMachineSizes = append(virtualMachineSizes, size.Name)
				}
			}
			all = append(all, sizes...)
		}
	}

	saveVMSizesAsJSON("vm-size-location.json", all)
	saveStringArrayAsJSON("vm-size.json", virtualMachineSizes)

	saveFileToBlobStore("vm-size.json")

	fmt.Println("Complete")
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
