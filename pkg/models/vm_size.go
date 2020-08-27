package models

type VmSize struct {
	Name     string
	Location Region
}

type VmSizes []VmSize
