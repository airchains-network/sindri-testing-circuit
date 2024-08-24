package ogcircuit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/frontend"
)

type Circuit struct {
	To              [25]frontend.Variable `gnark:",public"`
	From            [25]frontend.Variable `gnark:",public"`
	Amount          [25]frontend.Variable
	TransactionHash [25]frontend.Variable `gnark:",public"`
	FromBalances    [25]frontend.Variable
	ToBalances      [25]frontend.Variable
}

// Define checks and updates balances for each circuit element.
func (circuit *Circuit) Define(api frontend.API) error {
	for i := 0; i < 25; i++ {
		api.AssertIsLessOrEqual(circuit.Amount[i], circuit.FromBalances[i])

		updatedFromBalance := api.Sub(circuit.FromBalances[i], circuit.Amount[i])
		updatedToBalance := api.Add(circuit.ToBalances[i], circuit.Amount[i])

		api.AssertIsEqual(updatedFromBalance, api.Sub(circuit.FromBalances[i], circuit.Amount[i]))
		api.AssertIsEqual(updatedToBalance, api.Add(circuit.ToBalances[i], circuit.Amount[i]))
	}
	return nil
}

// Common utility for reading JSON in from a file.
func ReadFromInputPath(pathInput string) (map[string]interface{}, error) {
	absPath, err := filepath.Abs(pathInput)
	if err != nil {
		fmt.Println("Error constructing absolute path:", err)
		return nil, err
	}

	file, err := os.Open(absPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var data map[string]interface{}
	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		panic(err)
	}

	return data, nil
}

// Construct a witness from input data in a JSON file.
func FromJson(pathInput string) witness.Witness {
	data, err := ReadFromInputPath(pathInput)
	if err != nil {
		panic(err)
	}

	// Extract arrays from JSON data
	toArray := data["to"].([]interface{})
	fromArray := data["from"].([]interface{})
	amountArray := data["amount"].([]interface{})
	transactionHashArray := data["transactionHash"].([]interface{})
	fromBalancesArray := data["fromBalances"].([]interface{})
	toBalancesArray := data["toBalances"].([]interface{})

	// Initialize the Circuit struct with array values
	var circuit Circuit
	for i := 0; i < 25; i++ {
		circuit.To[i] = frontend.Variable(toArray[i])
		circuit.From[i] = frontend.Variable(fromArray[i])
		circuit.Amount[i] = frontend.Variable(amountArray[i])
		circuit.TransactionHash[i] = frontend.Variable(transactionHashArray[i])
		circuit.FromBalances[i] = frontend.Variable(fromBalancesArray[i])
		circuit.ToBalances[i] = frontend.Variable(toBalancesArray[i])
	}

	// Create the witness
	w, err := frontend.NewWitness(&circuit, ecc.BLS12_381.ScalarField())
	if err != nil {
		panic(err)
	}
	return w
}
