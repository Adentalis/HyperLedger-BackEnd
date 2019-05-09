// DISCLAIMER:
// THIS SAMPLE CODE MAY BE USED SOLELY AS PART OF THE TEST AND EVALUATION OF THE SAP CLOUD PLATFORM
// BLOCKCHAIN SERVICE (THE “SERVICE”) AND IN ACCORDANCE WITH THE TERMS OF THE AGREEMENT FOR THE SERVICE.
// THIS SAMPLE CODE PROVIDED “AS IS”, WITHOUT ANY WARRANTY, ESCROW, TRAINING, MAINTENANCE, OR SERVICE
// OBLIGATIONS WHATSOEVER ON THE PART OF SAP.

//=================================================================================================
//========================================================================================== IMPORT
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

//=================================================================================================
//================================================================================= RETURN HANDLING

// Success HTTP 2xx with a payload
func Success(rc int32, doc string, payload []byte) peer.Response {
	return peer.Response{
		Status:  rc,
		Message: doc,
		Payload: payload,
	}
}

// Error HTTP 4xx or 5xx with an error message
func Error(rc int32, doc string) peer.Response {
	logger.Errorf("Error %d = %s", rc, doc)
	return peer.Response{
		Status:  rc,
		Message: doc,
	}
}

//=================================================================================================
//======================================================================================= MAIN/INIT
var logger = shim.NewLogger("chaincode")

type CRUD struct {
}

// this is needed to create cars in the init func - this has nothing to do with the model definition in the yaml file
type Car struct {
	Id       int `json:"id"`
	Km       int `json:"km"`
	BorrowId int `json:"borrowId"`
}

type User struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	BorrowId int    `json:"borrowId"`
}

//this one will be written to the Ledger
type CarBorrow struct {
	Id        int    `json:"id"`
	CarId     int    `json:"carId"`
	UserId    int    `json:"userId"`
	StartTime string `json:"startTime"`
}

//this one is just for internal Operations in func borrowACar
type CheckBorrowCarParameter struct {
	CarId int `json:"carId"`
}

//this one is just for internal Operations in func returnACar
type CheckReturnCarParameter struct {
	NewKm int    `json:"newKm"`
	Usage string `json:"usage"`
}

type TravelLog struct {
	Id        int    `json:"id"`
	UserId    int    `json:"userId"`
	CarId     int    `json:"carId"`
	Usage     string `json:"usage"`
	StartKm   int    `json:"startKm"`
	EndKm     int    `json:"endKm"`
	DrivenKm  int    `json:"drivenKm"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

func main() {
	if err := shim.Start(new(CRUD)); err != nil {
		fmt.Printf("Main: Error starting chaincode: %s", err)
	}
	logger.SetLevel(shim.LogInfo)
}

//this func is called when smart contract get instantiated
//init two cars automatically
func (cc *CRUD) Init(stub shim.ChaincodeStubInterface) peer.Response {

	cars := []Car{
		Car{Id: 1, Km: 1000, BorrowId: 0},
		Car{Id: 2, Km: 1500, BorrowId: 0},
		Car{Id: 3, Km: 1800, BorrowId: 0},
		Car{Id: 4, Km: 6500, BorrowId: 0},
		Car{Id: 5, Km: 1690, BorrowId: 0},
	}

	users := []User{
		User{Id: 1, Name: "Alice", BorrowId: 0},
		User{Id: 2, Name: "Bob", BorrowId: 0},
		User{Id: 3, Name: "Charlie", BorrowId: 0},
		User{Id: 4, Name: "Delta", BorrowId: 0},
		User{Id: 5, Name: "Eve", BorrowId: 0},
	}

	i := 0
	for i < len(cars) {
		carAsBytes, _ := json.Marshal(cars[i])
		userAsBytes, _ := json.Marshal(users[i])
		stub.PutState("car"+strconv.Itoa(i+1), carAsBytes)
		stub.PutState("user"+strconv.Itoa(i+1), userAsBytes)
		i = i + 1
	}
	//init borrow counter
	stub.PutState("counterB", []byte(strconv.Itoa(0)))

	return Success(http.StatusNoContent, "OK", nil)
}

//=================================================================================================
//========================================================================================== INVOKE
// Invoke is called to update or query the ledger in a proposal transaction.
// Updated state variables are not committed to the ledger until the
// transaction is committed.
//
func (cc *CRUD) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

	function, args := stub.GetFunctionAndParameters()

	switch strings.ToLower(function) {
	//CAR OPERATIONS
	case "createcar":
		return cc.createCar(stub, args)
	case "getcarbyid":
		return cc.getCar(stub, args)
	case "updatecar":
		return cc.updateCar(stub, args)
	case "deletecar":
		return cc.deleteCar(stub, args)
	case "getallcars":
		return cc.getAllCars(stub, args)

	//USER OPERATIONS
	case "createuser":
		return cc.createUser(stub, args)
	case "getuserbyid":
		return cc.getUser(stub, args)
	case "updateuser":
		return cc.updateUser(stub, args)
	case "deleteuser":
		return cc.deleteUser(stub, args)
	case "getalluser":
		return cc.getAllUser(stub, args)

	//USER OPERATION
	case "userborrowacar":
		return cc.userBorrowACar(stub, args)
	case "userreturnacar":
		return cc.userReturnACar(stub, args)
	case "getalltravellogsforuser":
		return cc.getAllTravelLogsForUser(stub, args)

	//ADMINISTRATION

	case "getborrowlogbyid":
		return cc.getBorrowLogById(stub, args)
	case "gettravellogbyid":
		return cc.getTravelLogById(stub, args)
	case "getallborrowlogs":
		return cc.getAllBorrowLogs(stub, args)
	case "getalltravellogs":
		return cc.getAllTravelLogs(stub, args)

	//TESTING
	case "getallkeys":
		return cc.getAllKeys(stub, args)
	case "getallvalues":
		return cc.getAllValues(stub, args)
	case "getalldata":
		return cc.getAllData(stub, args)
	default:
		logger.Warningf("Invoke('%s') invalid!", function)
		return Error(http.StatusNotImplemented, "Invalid method name!!!")
	}
}

//========================================CAR=================================================
//=====================================GET-CAR================================================
func (cc *CRUD) getCar(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if msg, err := stub.GetState("car" + args[0]); err == nil && msg != nil {
		return Success(http.StatusOK, "OK", msg)
	} else {
		return Error(http.StatusNotFound, "Car Not Found")
	}
}

//===================================CREATE-CAR===============================================
func (cc *CRUD) createCar(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//(path) -> args[0]: "id"
	//(body) -> args[1]: {"id":7,"km":7777,"borrowId":0}
	if obj, err := stub.GetState("car" + args[0]); err != nil || obj != nil {
		return Error(http.StatusConflict, "a car with this id already exists")
	}

	//create a car with the overgiven parameters
	noSlashCar := strings.Replace(args[1], "\\", "", -1)
	var car Car
	err := json.Unmarshal([]byte(noSlashCar), &car)
	if err != nil {
		return Error(http.StatusBadRequest, "Unmarshalling the overgiven Data failed - args[1]= "+args[1]+" :: noSlashCar= "+noSlashCar)
	}

	//check if the car has all three values
	if car.Id == 0 || car.Km == 0 || car.BorrowId != 0 {
		return Error(http.StatusBadRequest, "One parameter is wrong! car .id= "+strconv.Itoa(car.Id)+" - .km= "+strconv.Itoa(car.Km)+" - .borrowId= "+strconv.Itoa(car.BorrowId))
	}

	//check if car.id is the same id as in path
	erg, err := strconv.Atoi(args[0])
	if err != nil {
		return Error(http.StatusInternalServerError, "path id cant be changed to a string - should never happen!")
	}
	if car.Id != erg {
		return Error(http.StatusBadRequest, "id of path and id of car are different!")
	}

	if err := stub.PutState("car "+args[0], []byte(noSlashCar)); err == nil {
		stub.SetEvent("Car created"+args[1]+" __ "+noSlashCar, []byte("Success"))
		return Success(http.StatusCreated, "Ok", nil)
	} else {
		return Error(http.StatusInternalServerError, err.Error())
	}

}

//====================================PUT-CAR=================================================
func (cc *CRUD) updateCar(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if obj, err := stub.GetState("car" + args[0]); obj == nil || err != nil {
		return Error(http.StatusNotFound, "this car does not exist")
	}

	var car Car
	json.Unmarshal([]byte(args[1]), &car)

	//check if the car has all three values
	if car.Id == 0 || car.Km == 0 || car.BorrowId != 0 {
		return Error(http.StatusBadRequest, "one parameter is wrong!")
	}

	//check if car.id is the same id as in path
	erg, err := strconv.Atoi(args[0])
	if err != nil {
		return Error(http.StatusBadRequest, "path id cant be changed to a string - should never happen!")
	}
	if car.Id != erg {
		return Error(http.StatusBadRequest, "id of path and id of car are different!")
	}

	if err := stub.PutState("car"+args[0], []byte(args[1])); err == nil {
		stub.SetEvent("Car updated", []byte("Success"))
		return Success(http.StatusCreated, "Updated", nil)
	} else {
		return Error(http.StatusInternalServerError, err.Error())
	}

}

//====================================DELETE-CAR==================================================
func (cc *CRUD) deleteCar(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if msg, err := stub.GetState("car" + args[0]); err != nil || msg == nil {
		return Error(http.StatusNotFound, "Car Not Found")
	}

	err := stub.DelState("car" + args[0])
	if err != nil {
		return Error(http.StatusInternalServerError, "Something bad happend")
	} else {
		stub.SetEvent("Car deleted", []byte("Success"))
		return Success(http.StatusOK, "OK", []byte("Car deleted"))
	}
}

//========================================================================================
//========================================USERS===========================================
//=======================================GET-USER=========================================
func (cc *CRUD) getUser(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if msg, err := stub.GetState("user" + args[0]); err == nil && msg != nil {
		return Success(http.StatusOK, "OK", msg)
	} else {
		return Error(http.StatusNotFound, "User Not Found")
	}
}

//===============================POST-USER=============================================
func (cc *CRUD) createUser(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if obj, err := stub.GetState("user" + args[0]); err != nil || obj != nil {
		return Error(http.StatusConflict, "this user already exists")
	}

	//validate args[1] -> JSON object. This must have a specific look
	//args[]: "idNumber"
	//args[1]: {"id":"string","km":"string","owner":"string"} standard

	//create a car with the overgiven parameters
	var user User
	json.Unmarshal([]byte(args[1]), &user)

	//check if the car has all three values
	if user.Id == 0 || user.Name == "" || user.BorrowId != 0 {
		return Error(http.StatusBadRequest, "one parameter is wrong!")
	}

	//check if car.id is the same id as in path
	erg, err := strconv.Atoi(args[0])
	if err != nil {
		return Error(http.StatusBadRequest, "path id cant be changed to a string - should never happen!")
	}
	if user.Id != erg {
		return Error(http.StatusBadRequest, "id of path and id of car are different!")
	}

	if err := stub.PutState("user"+args[0], []byte(args[1])); err == nil {
		stub.SetEvent("User created", []byte("Success"))
		return Success(http.StatusCreated, "Created", nil)
	} else {
		return Error(http.StatusInternalServerError, err.Error())
	}

}

//=============================PUT-USER====================================================
func (cc *CRUD) updateUser(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if obj, err := stub.GetState("user" + args[0]); obj == nil || err != nil {
		return Error(http.StatusLocked, "this user does not exist")
	}

	var user User
	json.Unmarshal([]byte(args[1]), &user)

	//check if the car has all three values
	if user.Id == 0 || user.Name == "" || user.BorrowId != 0 {
		return Error(http.StatusBadRequest, "one parameter is wrong!")
	}

	//check if car.id is the same id as in path
	erg, err := strconv.Atoi(args[0])
	if err != nil {
		return Error(http.StatusBadRequest, "path id cant be changed to a string - should never happen!")
	}
	if user.Id != erg {
		return Error(http.StatusBadRequest, "id of path and id of car are different!")
	}

	if err := stub.PutState("user"+args[0], []byte(args[1])); err == nil {
		stub.SetEvent("User updated", []byte("Success"))
		return Success(http.StatusCreated, "Created", nil)
	} else {
		return Error(http.StatusInternalServerError, err.Error())
	}

}

//===============================DELETE-USER===================================================
func (cc *CRUD) deleteUser(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if msg, err := stub.GetState("user" + args[0]); err != nil || msg == nil {
		return Error(http.StatusNotFound, "User Not Found")
	}

	err := stub.DelState("user" + args[0])
	if err != nil {
		return Error(http.StatusInternalServerError, "Something bad happend")
	} else {
		stub.SetEvent("User deleted", []byte("Success"))
		return Success(http.StatusOK, "OK", []byte("User deleted"))
	}
}

//======================================================================================
//============================USER OPERATIONS=======================================
//===========================USER BORROW A CAR===========================================

func (cc *CRUD) userBorrowACar(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	//create a borrow obj. and init it with the overgiven parameters
	var overgivenParam CheckBorrowCarParameter
	json.Unmarshal([]byte(args[1]), &overgivenParam)

	var overgivenUserId, err = strconv.Atoi(args[0])
	if err != nil {
		return Error(http.StatusBadRequest, "Cant Atoi args[0] ")
	}

	//check if all parameters have a value
	if overgivenParam.CarId == 0 || overgivenUserId == 0 {
		return Error(http.StatusBadRequest, "one parameter is wrong!")
	}

	//create counter and init it with the data in ledger
	obj, _ := stub.GetState("counterB")
	counter, _ := strconv.Atoi(string(obj))
	counter += 1

	//create Starttime
	time := time.Now()
	timeString := time.Format("2006-01-02 15:04:05")

	//create CarBorrow struct and put it in the ledger
	carBorrow := CarBorrow{Id: counter, CarId: overgivenParam.CarId, UserId: overgivenUserId, StartTime: timeString}
	carBorrowAsBytes, _ := json.Marshal(carBorrow)
	stub.PutState("borrow"+strconv.Itoa(counter), carBorrowAsBytes)

	//update cborrow
	stub.PutState("counterB", []byte(strconv.Itoa(counter)))

	//update user
	ledgerUser, _ := stub.GetState("user" + strconv.Itoa(overgivenUserId))
	var user User
	json.Unmarshal([]byte(ledgerUser), &user)

	if user.Id == 0 {
		return Error(http.StatusNotFound, "User Not Found - wrong overgiven userId!")
	}
	if user.BorrowId != 0 {
		return Error(http.StatusConflict, "User is already borrowing a car!")
	}

	user.BorrowId = counter
	userAsBytes, _ := json.Marshal(user)
	stub.PutState("user"+strconv.Itoa(user.Id), userAsBytes)

	//update car
	ledgerCar, _ := stub.GetState("car" + strconv.Itoa(overgivenParam.CarId))
	var car Car
	json.Unmarshal([]byte(ledgerCar), &car)

	if car.Id == 0 {
		return Error(http.StatusNotFound, "Car Not Found - wrong overgiven carId!")
	}
	if car.BorrowId != 0 {
		return Error(http.StatusConflict, "Car is already borrowed by a Car!")
	}

	car.BorrowId = counter
	carAsBytes, _ := json.Marshal(car)
	stub.PutState("car"+strconv.Itoa(car.Id), carAsBytes)

	//Create Event
	eventString := "User with id: " + strconv.Itoa(user.Id) + " borrowed successfully car with id: " + strconv.Itoa(car.Id)
	stub.SetEvent("Borrow a car", []byte(eventString))
	return Success(http.StatusOK, "OK", []byte("Borrow a car accepted"))

}

//========================GET A TRAVELLOG BY ID========================================
func (cc *CRUD) getAllTravelLogsForUser(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return Error(http.StatusBadRequest, "Parameter Mismatch")
	}

	intargs, err := strconv.Atoi(args[0])
	if err != nil {
		return Error(http.StatusBadRequest, "overgiven header cant be converted to an int")
	}

	//return all keys between t and u = all TravelLogs
	resultsIterator, err := stub.GetStateByRange("t", "u")
	if err != nil {
		return Error(http.StatusInternalServerError, err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[\n")

	var travelLog TravelLog
	for resultsIterator.HasNext() {
		it, _ := resultsIterator.Next()

		ledgerTravelLog, _ := stub.GetState(it.Key)
		json.Unmarshal([]byte(ledgerTravelLog), &travelLog)

		if travelLog.UserId == intargs {

			buffer.WriteString(string(it.Value))
			buffer.WriteString(",\n")
		}
	}

	buffer.WriteString("]")

	return Success(http.StatusOK, "OK", buffer.Bytes())
}

//===========================USER CAN RETURN HIS CAR==============================
func (cc *CRUD) userReturnACar(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	//get User out of ledger and init it here in Code
	ledgerUser, err := stub.GetState("user" + args[0])
	if err != nil || ledgerUser == nil {
		return Error(http.StatusBadRequest, "This user doenst exist!")
	}

	var user User
	json.Unmarshal([]byte(ledgerUser), &user)

	//check if a car is borrowed
	if user.BorrowId == 0 {
		return Error(http.StatusBadRequest, "This user dont have a borrowed car!")
	}

	//init values of body
	var overgivenParam CheckReturnCarParameter
	json.Unmarshal([]byte(args[1]), &overgivenParam)

	//check if all parameters have a value
	if overgivenParam.NewKm == 0 || overgivenParam.Usage == "" {
		return Error(http.StatusBadRequest, "Overgiven paramters are wrong!")
	}

	//get the borrowInformation to get the borrowed car
	var carBorrow CarBorrow
	ledgerBorrow, _ := stub.GetState("borrow" + strconv.Itoa(user.BorrowId))
	json.Unmarshal([]byte(ledgerBorrow), &carBorrow)

	if carBorrow.CarId == 0 {
		return Error(http.StatusBadRequest, "Couldnt find car!")
	}

	//get the car
	var car Car
	ledgerCar, _ := stub.GetState("car" + strconv.Itoa(carBorrow.CarId))
	json.Unmarshal([]byte(ledgerCar), &car)

	if car.BorrowId != user.BorrowId {
		return Error(http.StatusBadRequest, "This Should not happen - check for user.Borrowid != car.Borrowid")
	}

	//check the overgiven Km for corectness
	if car.Km > overgivenParam.NewKm {
		return Error(http.StatusBadRequest, "the overgiven newKm are lower than the km of the car when borrowed")
	}

	//create new travelLog and put it in the ledger
	time := time.Now()
	timeString := time.Format("2006-01-02 15:04:05")
	drivenKm := overgivenParam.NewKm - car.Km

	travelLog := TravelLog{
		Id:        carBorrow.Id,
		UserId:    user.Id,
		CarId:     car.Id,
		Usage:     overgivenParam.Usage,
		StartKm:   car.Km,
		EndKm:     overgivenParam.NewKm,
		DrivenKm:  drivenKm,
		StartTime: carBorrow.StartTime,
		EndTime:   timeString,
	}

	travelLogAsBytes, _ := json.Marshal(travelLog)
	if err := stub.PutState("travelLog"+strconv.Itoa(travelLog.Id), travelLogAsBytes); err != nil {
		return Error(http.StatusInternalServerError, "create travelLog failed")
	}

	//update user
	user.BorrowId = 0
	userAsBytes, _ := json.Marshal(user)
	if err := stub.PutState("user"+strconv.Itoa(user.Id), userAsBytes); err != nil {
		return Error(http.StatusInternalServerError, "Update user failed")
	}

	//update car
	car.BorrowId = 0
	car.Km = overgivenParam.NewKm
	carAsBytes, _ := json.Marshal(car)
	if err := stub.PutState("car"+strconv.Itoa(car.Id), carAsBytes); err != nil {
		return Error(http.StatusInternalServerError, "Update car failed")
	}

	//create Event when everything went right
	str := "User: " + strconv.Itoa(user.Id) + " returened his Car: " + strconv.Itoa(car.Id) + " --> TravelLog: " + strconv.Itoa(travelLog.Id) + " created!"
	stub.SetEvent("User returned Car", []byte(str))
	return Success(http.StatusOK, "OK", []byte("Car returned"))

}

//=====================================================================================
//===========================ADMINISTRATION============================================
//========================GET A TRAVELLOG BY ID========================================
func (cc *CRUD) getTravelLogById(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if msg, err := stub.GetState("travelLog" + args[0]); err == nil && msg != nil {
		return Success(http.StatusOK, "OK", msg)
	} else {
		return Error(http.StatusNotFound, "TravelLog Not Found")
	}

}

//===========================USER GETS ALL HIS BORROWLOGS===========================================
func (cc *CRUD) getBorrowLogById(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if msg, err := stub.GetState("borrow" + args[0]); err == nil && msg != nil {
		return Success(http.StatusOK, "OK", msg)
	} else {
		return Error(http.StatusNotFound, "Borrow Not Found")
	}
}

//==============GET ALL CARS================================================================== READ
func (cc *CRUD) getAllCars(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	//check for param length - No params needed here
	if len(args) != 0 {
		return Error(http.StatusBadRequest, "Parameter Mismatch")
	}

	//safe in resultsIterator all avaible keys for cars
	resultsIterator, err := stub.GetStateByRange("car", "caw")
	if err != nil {
		return Error(http.StatusInternalServerError, err.Error())
	}
	defer resultsIterator.Close()

	//create bytearray buffer to write in
	var buffer bytes.Buffer
	buffer.WriteString("[\n")

	//as long there is a next key, write the value of the next key to the buffer
	for resultsIterator.HasNext() {
		it, _ := resultsIterator.Next()

		buffer.WriteString(string(it.Value))
		buffer.WriteString(",\n")
	}

	buffer.WriteString("]")

	return Success(http.StatusOK, "OK", buffer.Bytes())
}

//============GET ALL USERS===SAME AS GETALLCARS======================================== READ
func (cc *CRUD) getAllUser(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 0 {
		return Error(http.StatusBadRequest, "Parameter Mismatch")
	}

	resultsIterator, err := stub.GetStateByRange("u", "v")
	if err != nil {
		return Error(http.StatusInternalServerError, err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[\n")

	for resultsIterator.HasNext() {
		it, _ := resultsIterator.Next()

		buffer.WriteString(string(it.Value))
		buffer.WriteString(",\n")
	}

	buffer.WriteString("]")

	return Success(http.StatusOK, "OK", buffer.Bytes())
}

//==============GET ALL BORROWLOGS======SAME AS GETALLCARS============================
func (cc *CRUD) getAllBorrowLogs(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 0 {
		return Error(http.StatusBadRequest, "Parameter Mismatch")
	}

	resultsIterator, err := stub.GetStateByRange("bor", "bot")
	if err != nil {
		return Error(http.StatusInternalServerError, err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[\n")

	for resultsIterator.HasNext() {
		it, _ := resultsIterator.Next()

		buffer.WriteString(string(it.Value))
		buffer.WriteString(",\n")
	}

	buffer.WriteString("]")

	return Success(http.StatusOK, "OK", buffer.Bytes())
}

//==================GET ALL TRAVELLOGS=======SAME AS GETALLCARS=====================
func (cc *CRUD) getAllTravelLogs(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 0 {
		return Error(http.StatusBadRequest, "Parameter Mismatch")
	}

	resultsIterator, err := stub.GetStateByRange("t", "u")
	if err != nil {
		return Error(http.StatusInternalServerError, err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[\n")

	for resultsIterator.HasNext() {
		it, _ := resultsIterator.Next()

		buffer.WriteString(string(it.Value))
		buffer.WriteString(",\n")
	}

	buffer.WriteString("]")

	return Success(http.StatusOK, "OK", buffer.Bytes())
}

//=====================TESTING=====================================
//==================GET ALL KEYS==================================
func (cc *CRUD) getAllKeys(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 0 {
		return Error(http.StatusBadRequest, "Parameter Mismatch")
	}

	resultsIterator, err := stub.GetStateByRange("", "")
	if err != nil {
		return Error(http.StatusInternalServerError, err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("{ \"ids\": [")

	for resultsIterator.HasNext() {
		it, _ := resultsIterator.Next()
		buffer.WriteString("\n")
		buffer.WriteString("\"")
		buffer.WriteString(it.Key)
		buffer.WriteString("\"")
	}
	buffer.WriteString("\n")
	buffer.WriteString("]}")

	return Success(http.StatusOK, "OK", buffer.Bytes())
}

//==================GET ALL Values==================================
func (cc *CRUD) getAllValues(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 0 {
		return Error(http.StatusBadRequest, "Parameter Mismatch")
	}

	resultsIterator, err := stub.GetStateByRange("", "")
	if err != nil {
		return Error(http.StatusInternalServerError, err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")

	for resultsIterator.HasNext() {
		it, _ := resultsIterator.Next()
		buffer.WriteString("\n")
		buffer.WriteString("\"")
		buffer.WriteString(string(it.Value))
		buffer.WriteString("\"")
	}

	buffer.WriteString("\n")
	buffer.WriteString("]}")

	return Success(http.StatusOK, "OK", buffer.Bytes())
}

func (cc *CRUD) getAllData(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 0 {
		return Error(http.StatusBadRequest, "Parameter Mismatch")
	}

	resultsIterator, err := stub.GetStateByRange("", "")
	if err != nil {
		return Error(http.StatusInternalServerError, err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")

	for resultsIterator.HasNext() {
		it, _ := resultsIterator.Next()
		buffer.WriteString("\n")
		buffer.WriteString("\"")
		buffer.WriteString(it.Key)
		buffer.WriteString("\" --> ")
		buffer.WriteString(string(it.Value))

	}
	buffer.WriteString("\n")
	buffer.WriteString("]")

	return Success(http.StatusOK, "OK", buffer.Bytes())
}
