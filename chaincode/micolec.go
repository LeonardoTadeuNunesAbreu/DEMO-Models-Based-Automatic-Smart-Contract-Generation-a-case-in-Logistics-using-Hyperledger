package chaincode

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"micolec/chaincode/models"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

type Entity string

const (
	EntityAuction          Entity = "AUCTION"
	EntityParcel           Entity = "PARCEL"
	EntityAuctionHasParcel Entity = "AUCTION_HAS_PARCEL"
	EntityBid              Entity = "BID"
)

// ** -----------------------------------------------------
// ** ENTITY RECORDS METHODS
// ** -> BEGIN
// ** -----------------------------------------------------

// Create a compositeKey
func (s *SmartContract) CreateCompositeKey(ctx contractapi.TransactionContextInterface, entity Entity, ids []string) (string, error) {
	compositeKey, err := ctx.GetStub().CreateCompositeKey(string(entity), ids)
	return compositeKey, err
}

// Create Record -> Receive compositeKey + Dados do registo!
func (s *SmartContract) UpsertEntityRecord(ctx contractapi.TransactionContextInterface, key string, data any) (bool, error) {
	var jsonData, err = json.Marshal(data)
	if err != nil {
		return false, err
	}
	err = ctx.GetStub().PutState(key, jsonData)
	if err != nil {
		return false, err
	}
	return true, nil
}

// Check if record Exists
func (s *SmartContract) EntityRecordExists(ctx contractapi.TransactionContextInterface, key string) (bool, error) {
	recordJSON, err := ctx.GetStub().GetState(key)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return recordJSON != nil, nil
}

// Read record from ledger
func (s *SmartContract) ReadEntity(ctx contractapi.TransactionContextInterface, key string) ([]byte, error) {
	recordJSON, err := ctx.GetStub().GetState(key)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if recordJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", key)
	}
	return recordJSON, nil
}

func (s *SmartContract) createEntityIterator(ctx contractapi.TransactionContextInterface, entity Entity, ids []string) (shim.StateQueryIteratorInterface, error) {
	entityIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(string(entity), ids)
	if err != nil {
		return nil, err
	}
	return entityIterator, nil
}

func (s *SmartContract) StartTransaction(ctx contractapi.TransactionContextInterface) (string, error) {
	transactionId := ctx.GetStub().GetTxID()
	err := ctx.GetStub().SetEvent("tx.start", []byte(transactionId))
	if err != nil {
		return "", err
	}

	return transactionId, nil
}

func (s *SmartContract) CloseTransaction(ctx contractapi.TransactionContextInterface, transactionId string, err error) {
	if err != nil {
		// Roll back the transaction
		_ = ctx.GetStub().SetEvent("tx.rollback", []byte(transactionId))
		_ = ctx.GetStub().DelState(transactionId)
		return
	}
	// Commit the transaction
	_ = ctx.GetStub().SetEvent("tx.commit", []byte(transactionId))
}

// ** -----------------------------------------------------
// ** ENTITY RECORDS METHODS
// ** -> END
// ** -----------------------------------------------------

// ** -----------------------------------------------------
// ** INIT LEDGER
// ** -> BEGIN
// ** -----------------------------------------------------

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	var AUCTIONS = []models.Auction{
		{ID: 1, StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 2)},
		{ID: 2, StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 3)},
		{ID: 3, StartDate: time.Now(), EndDate: time.Now().AddDate(0, 1, 0)},
	}

	var PARCELS = []models.Parcel{
		{
			ID:                   1,
			AddedToPlatform:      time.Now(),
			RequiredDeliveryDate: time.Now(),
			PickupPostalArea:     "Funchal",
			DeliveryPostalArea:   "Santa Cruz",
			BitcircleReward:      10,
			Weight:               100,
			Length:               10,
			Width:                100,
			Height:               13,
			Volume:               10,
		},
		{
			ID:                   2,
			AddedToPlatform:      time.Now(),
			RequiredDeliveryDate: time.Now(),
			PickupPostalArea:     "Ribeira Brava",
			DeliveryPostalArea:   "Caniçal",
			BitcircleReward:      10,
			Weight:               100,
			Length:               10,
			Width:                100,
			Height:               13,
			Volume:               10,
		},
	}

	var AUCTION_HAS_PARCEL = []models.AuctionHasParcel{
		{AuctionID: 1, ParcelID: 1},
		{AuctionID: 2, ParcelID: 2},
		{AuctionID: 3, ParcelID: 1},
	}

	var BIDS = []models.Bid{
		{
			ID:              1,
			AuctionID:       2,
			Date:            time.Now(),
			BitcircleAmount: 0,
			MoneyAmount:     4.95,
			Status:          models.BitStatusOutBidded,
			Winner:          false,
		},
		{
			ID:              1,
			AuctionID:       2,
			Date:            time.Now(),
			BitcircleAmount: 0,
			MoneyAmount:     3.95,
			Status:          models.BitStatusLowerBid,
			Winner:          false,
		},
		{
			ID:              1,
			AuctionID:       3,
			Date:            time.Now(),
			BitcircleAmount: 0,
			MoneyAmount:     6.99,
			Status:          models.BitStatusLowerBid,
			Winner:          false,
		},
	}

	for _, auction := range AUCTIONS {
		auctionJSON, err := json.Marshal(auction)
		if err != nil {
			return err
		}

		compositeKey, err := s.CreateCompositeKey(ctx, EntityAuction, []string{fmt.Sprint(auction.ID)})
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(compositeKey, auctionJSON)
		if err != nil {
			return err
		}
	}

	for _, parcel := range PARCELS {
		parcelJSON, err := json.Marshal(parcel)
		if err != nil {
			return err
		}

		compositeKey, err := s.CreateCompositeKey(ctx, EntityParcel, []string{fmt.Sprint(parcel.ID)})
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(compositeKey, parcelJSON)
		if err != nil {
			return err
		}
	}

	for _, auctionHasParcel := range AUCTION_HAS_PARCEL {
		auctionHasParcelJSON, err := json.Marshal(auctionHasParcel)
		if err != nil {
			return err
		}

		compositeKey, err := s.CreateCompositeKey(ctx, EntityAuctionHasParcel, []string{fmt.Sprint(auctionHasParcel.AuctionID), fmt.Sprint(auctionHasParcel.ParcelID)})
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(compositeKey, auctionHasParcelJSON)
		if err != nil {
			return err
		}
	}

	for _, bid := range BIDS {
		bidJSON, err := json.Marshal(bid)
		if err != nil {
			return err
		}

		compositeKey, err := s.CreateCompositeKey(ctx, EntityBid, []string{fmt.Sprint(bid.ID)})
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(compositeKey, bidJSON)
		if err != nil {
			return err
		}
	}

	return nil
}

// ** -----------------------------------------------------
// ** INIT LEDGER
// ** -> END
// ** -----------------------------------------------------

// ** -----------------------------------------------------
// ** PARCEL
// ** -> BEGIN
// ** -----------------------------------------------------

func (s *SmartContract) ParcelDeliveryParcelAdded(ctx contractapi.TransactionContextInterface, parcel models.Parcel) (bool, error) {
	// Create and store parcel entity
	parcelKey, err := s.CreateCompositeKey(ctx, EntityParcel, []string{fmt.Sprint(parcel.ID)})
	if err != nil {
		return false, err
	}
	// Check if parcel all ready exists
	if parcelRecordExists, err := s.EntityRecordExists(ctx, parcelKey); err != nil {
		return false, err
	} else if parcelRecordExists {
		return false, errors.New("Record already exists")
	}

	// Validate parcel
	if err := validateParcelDeliveryParcelAdded(&parcel); err != nil {
		return false, err
	}

	// Insert Parcel on the blockchain
	_, err = s.UpsertEntityRecord(ctx, parcelKey, parcel)
	if err != nil {
		return false, err
	}

	return true, nil
}

// ** -----------------------------------------------------
// ** PARCEL
// ** -> END
// ** -----------------------------------------------------

// ** -----------------------------------------------------
// ** AUCTION
// ** -> BEGIN
// ** -----------------------------------------------------

func (s *SmartContract) ParcelDeliveryAuctionStart(ctx contractapi.TransactionContextInterface, parcels []models.AuctionHasParcel, auction models.Auction) (bool, error) {
	// Start a new transaction
	transactionId, err := s.StartTransaction(ctx)
	if err != nil {
		return false, err
	}
	// Roolback in case of error Or Commit in case of success
	defer s.CloseTransaction(ctx, transactionId, err)

	// Validate auction
	err = validateAuction(auction)
	if err != nil {
		return false, err
	}

	// Process parcels
	for _, auctionHasParcel := range parcels {
		// Check if ParcelExists
		parcelKey, err := s.CreateCompositeKey(ctx, EntityParcel, []string{fmt.Sprint(auctionHasParcel.ParcelID)})
		if _, err = s.EntityRecordExists(ctx, parcelKey); err != nil {
			return false, err
		}

		// Create and store auctionHasParcel entity
		var auctionHasParcelKey string
		auctionHasParcelKey, err = s.CreateCompositeKey(ctx, EntityAuctionHasParcel, []string{fmt.Sprint(auction.ID), fmt.Sprint(auctionHasParcel.ParcelID)})
		if err != nil {
			return false, err
		}
		_, err = s.UpsertEntityRecord(ctx, auctionHasParcelKey, auctionHasParcel)
		if err != nil {
			return false, err
		}
	}

	// Create and store auction entity
	var auctionKey string
	auctionKey, err = s.CreateCompositeKey(ctx, EntityAuction, []string{fmt.Sprint(auction.ID)})
	if err != nil {
		return false, err
	}
	_, err = s.UpsertEntityRecord(ctx, auctionKey, auction)
	if err != nil {
		return false, err
	}

	return true, nil
}

// ** -----------------------------------------------------
// ** AUCTION
// ** -> END
// ** -----------------------------------------------------

// ** -----------------------------------------------------
// ** BID
// ** -> START
// ** -----------------------------------------------------

func (s *SmartContract) ParcelDeliveryBidingRequest(ctx contractapi.TransactionContextInterface, bidID int, bitcircleAmount float32, moneyAmount float32, reverseEconomy bool, auctionId int) error {
	auctionKey, err := s.CreateCompositeKey(ctx, EntityAuction, []string{fmt.Sprint(auctionId)})
	if err != nil {
		return err
	}

	exists, err := s.EntityRecordExists(ctx, auctionKey)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("auction %d does not exist", auctionId)
	}

	// Check if the new bid is the lowest bid
	bidsIterator, err := s.createEntityIterator(ctx, EntityBid, []string{"", fmt.Sprint(auctionId)})
	if err != nil {
		return err
	}
	defer bidsIterator.Close()

	var lowestBid models.Bid
	var lowestBidKey string
	for bidsIterator.HasNext() {
		bidResponse, err := bidsIterator.Next()
		if err != nil {
			return err
		}

		var bid models.Bid
		err = json.Unmarshal(bidResponse.Value, &bid)
		if err != nil {
			return err
		}

		// Checks if this bid is the current winner and if input money ammount is less then winner bid ammount
		if bid.Winner && bid.MoneyAmount > moneyAmount {
			lowestBid = bid
			lowestBidKey = bidResponse.Key
			// If it finds the winner bet then theirs is no need to keep interating
			break
		}
	}

	// Check if new bid is lower than lowest bid
	// lowestBid.ID != 0 -> Checks if exists a lowerBid
	if lowestBid.ID != 0 && moneyAmount >= lowestBid.MoneyAmount {
		return fmt.Errorf("money amount of new bid should be lower than current lowest bid")
	}

	// Set previous lowest bid to "Outbidded" status and not a winner
	if lowestBidKey != "" {
		var prevBid models.Bid
		prevBidByte, err := ctx.GetStub().GetState(lowestBidKey)
		if err != nil {
			return err
		}
		err = json.Unmarshal(prevBidByte, &prevBid)
		if err != nil {
			return err
		}

		prevBid.Status = models.BitStatusOutBidded
		prevBid.Winner = false

		_, err = s.UpsertEntityRecord(ctx, lowestBidKey, prevBid)
		if err != nil {
			return err
		}
	}

	// Create new bid
	bid := models.Bid{
		ID:              bidID,
		Date:            time.Now(),
		BitcircleAmount: bitcircleAmount,
		MoneyAmount:     moneyAmount,
		Status:          models.BitStatusLowerBid,
		Winner:          true,
		AuctionID:       auctionId,
	}

	bidCompositeKey, err := s.CreateCompositeKey(ctx, EntityBid, []string{fmt.Sprint(bidID), fmt.Sprint(auctionId)})
	if err != nil {
		return err
	}

	_, err = s.UpsertEntityRecord(ctx, bidCompositeKey, bid)
	if err != nil {
		return err
	}

	return nil
}

// ** -----------------------------------------------------
// ** BID
// ** -> END
// ** -----------------------------------------------------

// ** -----------------------------------------------------
// ** VALIDATION METHODS
// ** -> START
// ** -----------------------------------------------------
func validateAuction(auction models.Auction) error {
	var errorMessages []string

	if !(auction.ID >= 1) {
		errorMessages = append(errorMessages, "ID Higher Equal 1")
	}

	// if auction.EndDate.Before(auction.StartDate) {
	// 	return errors.New("End date cannot be before start date")
	// }

	if !(auction.MaximumAmount > 0) {
		errorMessages = append(errorMessages, "MaximumAmount Higher than 0")
	}

	if len(errorMessages) > 0 {
		return errors.New(strings.Join(errorMessages, "\n"))
	}

	return nil
}

func validateParcelDeliveryParcelAdded(parcel *models.Parcel) error {
	var errorMessages []string

	if !(parcel.ID >= 1) {
		errorMessages =
			append(errorMessages, "ID Higher Equal 1")
	}

	if !(len(strings.TrimSpace(parcel.PickupPostalArea)) >= 3) {
		errorMessages =
			append(errorMessages, "PickupPostalArea Min. Length 3")
	}

	if !(len(strings.TrimSpace(parcel.DeliveryPostalArea)) >= 3) {
		errorMessages =
			append(errorMessages, "DeliveryPostalArea Min. Length 3")
	}

	if !(parcel.BitcircleReward <= 0) {
		errorMessages =
			append(errorMessages, "BitcircleReward Higher Equal 0")
	}

	if !(parcel.Weight > 0) {
		errorMessages =
			append(errorMessages, "Weight Higher than 0")
	}

	if !(parcel.Length >= 0) {
		errorMessages =
			append(errorMessages, "Length Higher than 0")
	}

	if !(parcel.Width > 0) {
		errorMessages =
			append(errorMessages, "Width Higher than 0")
	}

	if !(parcel.Height > 0) {
		errorMessages =
			append(errorMessages, "Height Higher than 0")
	}

	if !(parcel.Volume > 0) {
		errorMessages =
			append(errorMessages, "Volume Higher than 0")
	}

	if len(errorMessages) > 0 {
		return errors.New(strings.Join(errorMessages, "\n"))
	}

	return nil
}

// ** -----------------------------------------------------
// ** VALIDATION METHODS
// ** -> END
// ** -----------------------------------------------------
