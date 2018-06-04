package main

import (
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"encoding/json"
	"bytes"
	"time"
	"strconv"
)

// ===========================
// Init initializes chaincode
// ===========================
func (t *ChainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// ========================================
// Invoke - Entry point for Invocations
// ========================================
func (t *ChainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	// Handle different functions
	if function == "add" {
		return addToCart(stub, args)
	} else if function == "remove" {
		return removeFromCart(stub, args)
	} else if function == "buy" {
		return buy(stub, args)
	} else if function == "getCartInfo" {
		return getCartInfo(stub, args[0])
	} else if function == "getCartHistory" {
		return getCartHistory(stub, args[0])
	} else if function == "registerUser" {
		return registerUser(stub, args[0])
	} else {
		return shim.Error("Received unknown function invocation")
	}
}

// ============================================================================================================================
// registerUser() - Register user
// ============================================================================================================================
func registerUser(stub shim.ChaincodeStubInterface, userID string) pb.Response {

	//Check user exists on chain or not
	userInfoAsBytes, err := stub.GetState(userID)
	if userInfoAsBytes != nil || err != nil{
		return shim.Error("A user with this ID already exists on the system.")
	}

	var user User
	var cart Cart

	cart.Status = EMPTY

	user.UserID = userID
	user.CartDetail = append(user.CartDetail, cart)

	//Marshalling final user detail
	val, err := json.Marshal(user)
	if err != nil {
		shim.Error(MarshalErrorMessage)
	}

	//Storing user detail corresponding to UserID
	err = stub.PutState(userID, []byte(val))
	if err != nil {
		return shim.Error(PutErrorMessage)
	}

	return shim.Success([]byte(userID))
}

// ============================================================================================================================
// addToCart() - Add product to cart
// ============================================================================================================================
func addToCart(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//Check the correct no. of arguments in an array
	if len(args) < 2 {
		str := fmt.Sprintf("Invalid request")
		return shim.Error(str)
	}

	userID := args[0]

	var userDetail User
	userInfoAsBytes, err := stub.GetState(userID)
	if err != nil {
		shim.Error(GetStateErrorMessage)
	}

	//Unmarshalling user json string to native go structure
	err = json.Unmarshal(userInfoAsBytes, &userDetail)
	if err != nil {
		return shim.Error(UnmarshalErrorMessage)
	}

	productData := args[1]
	var productDetail Product

	//Unmarshalling user json string to native go structure
	err = json.Unmarshal([]byte(productData), &productDetail)
	if err != nil {
		fmt.Println(err)
		return shim.Error(UnmarshalErrorMessage)
	}

	index := len(userDetail.CartDetail)

	userDetail.CartDetail[index].Products[productDetail.ProductID] = productDetail;

	if userDetail.CartDetail[index].Status == EMPTY {
		userDetail.CartDetail[index].Status = ADDED
	}

	userDetail.CartDetail[index].TotalPrice = userDetail.CartDetail[index].TotalPrice + productDetail.Price

	//Marshalling final user detail
	val, err := json.Marshal(userDetail)
	if err != nil {
		shim.Error(MarshalErrorMessage)
	}

	//Storing user detail corresponding to UserID
	err = stub.PutState(userID, []byte(val))
	if err != nil {
		return shim.Error(PutErrorMessage)
	}

	return shim.Success([]byte("Success"))
}

// ============================================================================================================================
// removeFromCart() - Remove product from cart
// ============================================================================================================================
func removeFromCart(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//Check the correct no. of arguments in an array
	if len(args) < 2 {
		str := fmt.Sprintf("Invalid request")
		return shim.Error(str)
	}

	userID := args[0]
	productID := args[1]

	var userDetail User
	userInfoAsBytes, err := stub.GetState(userID)
	if err != nil {
		shim.Error(GetStateErrorMessage)
	}

	//Unmarshalling user json string to native go structure
	err = json.Unmarshal(userInfoAsBytes, &userDetail)
	if err != nil {
		return shim.Error(UnmarshalErrorMessage)
	}

	index := len(userDetail.CartDetail)

	userDetail.CartDetail[index].TotalPrice = userDetail.CartDetail[index].TotalPrice - userDetail.CartDetail[index].Products[productID].Price

	delete(userDetail.CartDetail[index].Products, productID)

	if len(userDetail.CartDetail[index].Products) == 0 {
		userDetail.CartDetail[index].Status = EMPTY
	}

	//Marshalling final user detail
	val, err := json.Marshal(userDetail)
	if err != nil {
		shim.Error(MarshalErrorMessage)
	}

	//Storing user detail corresponding to UserID
	err = stub.PutState(userID, []byte(val))
	if err != nil {
		return shim.Error(PutErrorMessage)
	}

	return shim.Success([]byte("Success"))

}

// ============================================================================================================================
// buy() - Buy all the products in the cart
// ============================================================================================================================
func buy(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//Check the correct no. of arguments in an array
	if len(args) < 2 {
		str := fmt.Sprintf("Invalid request")
		return shim.Error(str)
	}

	userID := args[0]

	var userDetail User
	userInfoAsBytes, err := stub.GetState(userID)
	if err != nil {
		shim.Error(GetStateErrorMessage)
	}

	//Unmarshalling user json string to native go structure
	err = json.Unmarshal(userInfoAsBytes, &userDetail)
	if err != nil {
		return shim.Error(UnmarshalErrorMessage)
	}

	index := len(userDetail.CartDetail)

	if userDetail.CartDetail[index].Status == EMPTY {
		return shim.Error("Cart is empty")
	}

	userDetail.CartDetail[index].Status = PURCHASED
	userDetail.CartDetail[index].TransactionID = stub.GetTxID()

	var cart Cart

	//Add new cart
	cart.Status = EMPTY
	userDetail.CartDetail = append(userDetail.CartDetail, cart)

	//Marshalling final user detail
	val, err := json.Marshal(userDetail)
	if err != nil {
		shim.Error(MarshalErrorMessage)
	}

	//Storing user detail corresponding to UserID
	err = stub.PutState(userID, []byte(val))
	if err != nil {
		return shim.Error(PutErrorMessage)
	}

	return shim.Success([]byte("Success"))
}

// ============================================================================================================================
// getHistoryForKey() - Retrieve key history from ledger
// ============================================================================================================================
func getHistoryForKey(stub shim.ChaincodeStubInterface, key string) ([]byte, error) {

	resultsIterator, err := stub.GetHistoryForKey(key)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the key
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	return buffer.Bytes(), nil
}

// ============================================================================================================================
// getCartHistory() - History of cart from blockchain ledger
// ============================================================================================================================
func getCartHistory(stub shim.ChaincodeStubInterface, userID string) pb.Response {
	result, err := getHistoryForKey(stub, userID)

	if err != nil {
		return shim.Error("Unable to retrieve historical records by key")
	}
	return shim.Success(result)
}

// ============================================================================================================================
// getCartInfo() - Buy all the products in the cart
// ============================================================================================================================
func getCartInfo(stub shim.ChaincodeStubInterface, userID string) pb.Response {

	var userDetail User
	userInfoAsBytes, err := stub.GetState(userID)
	if err != nil {
		shim.Error(GetStateErrorMessage)
	}

	//Unmarshalling user json string to native go structure
	err = json.Unmarshal(userInfoAsBytes, &userDetail)
	if err != nil {
		return shim.Error(UnmarshalErrorMessage)
	}

	index := len(userDetail.CartDetail)

	cartInfo := userDetail.CartDetail[index]

	//Marshalling final cart detail
	val, err := json.Marshal(cartInfo)
	if err != nil {
		shim.Error(MarshalErrorMessage)
	}

	return shim.Success([]byte(val))
}

// ===================================================================================
// Bootstrap chaincode
// ===================================================================================
func main() {
	err := shim.Start(new(ChainCode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}