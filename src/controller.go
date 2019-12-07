package main

// access database
import (
	"fmt"
)

const (
	// ResourceNotFound -
	ResourceNotFound = "resource_not_found"
)

// init
func init() {
	fmt.Printf("initialize account test data\n")
	// TODO: build a list of existing accounts for validation
	//       fetch all accounts from DB and build an in-memory database

}
