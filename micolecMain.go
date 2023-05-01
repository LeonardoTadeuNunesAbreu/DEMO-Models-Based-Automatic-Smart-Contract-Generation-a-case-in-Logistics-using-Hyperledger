/*
SPDX-License-Identifier: Apache-2.0
*/
package main

import (
	"log"

	"micolec/chaincode"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	assetChaincode, err := contractapi.NewChaincode(&chaincode.SmartContract{})
	if err != nil {
		log.Panicf("Error creating micolec chaincode: %v", err)
	}

	if err := assetChaincode.Start(); err != nil {
		log.Panicf("Error starting micolec chaincode: %v", err)
	}
}
