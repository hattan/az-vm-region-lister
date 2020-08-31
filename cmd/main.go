package main

import (
	"fmt"
	"sync"

	"github.com/hattan/az-vm-region-lister/pkg/models"
	"github.com/hattan/az-vm-region-lister/pkg/services"
)

var all models.VmSizes

func main() {
	var wg sync.WaitGroup
	locations := services.GetLocations()

	c := make(chan models.VmSizes)

	for _, location := range locations {
		wg.Add(1)
		services.GetVMSizes(location, c, func() { wg.Done() })
	}

	go func() {
		defer close(c)
		wg.Wait()
	}()

	var virtualMachineSizes []string
	for sizes := range c {
		if sizes != nil && len(sizes) > 0 {
			for _, size := range sizes {
				if !services.StringInSlice(size.Name, virtualMachineSizes) {
					virtualMachineSizes = append(virtualMachineSizes, size.Name)
				}
			}
			all = append(all, sizes...)
		}
	}

	services.Save(all, virtualMachineSizes)
	fmt.Println("Complete")
}
