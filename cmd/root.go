package cmd

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"strings"

	simpleContract "github.com/HashRebel/gethalyzer/contracts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hpcloud/tail"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	logFile   string
	gethURI   string
	ethClient *ethclient.Client
)

var quit chan bool

type test struct {
	Name        string
	test        func()
	Description string
	FromAccount string
	ToAccount   string
	Contract    string
	Amount      float64
}

const (
	mainAddress         = "0x81edc9fc800e1b9c76be2f83e5c1dcc73f62980d"
	mainPrivateKey      = "6dfafde8135f35253f6482d80df34bd3ec52ad5733ab1edda4ba46110663d7d4"
	hashRebelAddress    = "0xf61bb995b5fb19aa0a38c7ecc9b52ff6199a69ca"
	hashRebelPrivateKey = "6dfafde8135f35253f6482d80df34bd3ec52ad5733ab1edda4ba46110663d7d4"
	secondaryAddress    = "0xc567982f00db259c2af4a6c7ed7b7e8ba393d695"
	logStamp            = "HASH_REBEL_LOG_STAMP"
)

const banner = `
         _          _            _       _    _                   _           _        _       _                 _            _      
        /\ \       /\ \         / /\    / /\ / /\                /\ \     _  /\ \     /\_\   /\ \               /\ \         /\ \    
       /  \ \      \_\ \       / / /   / / // /  \              /  \ \   /\_\\ \ \   / / /  /  \ \             /  \ \       /  \ \   
      / /\ \ \     /\__ \     / /_/   / / // / /\ \            / /\ \ \_/ / / \ \ \_/ / /__/ /\ \ \           / /\ \ \     / /\ \ \  
     / / /\ \_\   / /_ \ \   / /\ \__/ / // / /\ \ \          / / /\ \___/ /   \ \___/ //___/ /\ \ \         / / /\ \_\   / / /\ \_\ 
    / /_/_ \/_/  / / /\ \ \ / /\ \___\/ // / /  \ \ \        / / /  \/____/     \ \ \_/ \___\/ / / /        / /_/_ \/_/  / / /_/ / / 
   / /____/\    / / /  \/_// / /\/___/ // / /___/ /\ \      / / /    / / /       \ \ \        / / /        / /____/\    / / /__\/ /  
  / /\____\/   / / /      / / /   / / // / /_____/ /\ \    / / /    / / /         \ \ \      / / /    _   / /\____\/   / / /_____/   
 / / /______  / / /      / / /   / / // /_________/\ \ \  / / /    / / /           \ \ \     \ \ \__/\_\ / / /______  / / /\ \ \     
/ / /_______\/_/ /      / / /   / / // / /_       __\ \_\/ / /    / / /             \ \_\     \ \___\/ // / /_______\/ / /  \ \ \    
\/__________/\_\/       \/_/    \/_/ \_\___\     /____/_/\/_/     \/_/               \/_/      \/___/_/ \/__________/\/_/    \_\/   
`

const basicContract = `pragma solidity ^0.4.17;

contract simplestorage {
   uint public storedData;

   function simplestorage(uint initVal) public {
      storedData = initVal;
   }

   function set(uint x) public {
      storedData = x;
   }

   function get() public constant returns (uint retVal) {
      return storedData;
   }
}
`

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gethalyzer",
	Short: "Custom Geth log analyzer",
	Long: banner + `                                                                                                                            
This is a log analyzer that is intended to be used along with a hacked version 
of Geth which logs output to show the miner transaction logic. This tool is 
intended for demonstration purposes only and will be used in an interview
with Kaliedo. Please refer to the ../README.md for more information.
`,
	Run: runEthalyzer,
}

func runEthalyzer(cmd *cobra.Command, args []string) {
	if len(logFile) == 0 {
		cmd.Help()
		return
	}

	// TODO: Add new tests
	tests := []test{
		{Name: "Test: Send Eth - single tx", test: sendEthTest, Description: "Single simple TX: Transfer 3 eth.", FromAccount: fmt.Sprintf("Main (%s)", mainAddress), ToAccount: fmt.Sprintf("Hash Rebel (%s)", hashRebelAddress), Contract: "", Amount: 3},
		//{Name: "Test: Send Eth - multipule tx", test: test1, Description: "Multipule simple TX: Transfer 1 eth 20 accounts.", FromAccount: "Main (0x TODO)", ToAccount: "multipule", Contract: "", Amount: 0.1},
		{Name: "Test: Contract - creation", test: sendContractTest, Description: "Contract Creation TX", FromAccount: fmt.Sprintf("Main (%s)", mainAddress), ToAccount: "", Contract: basicContract, Amount: 0},
		//{Name: "Test: Contract - update", test: sendContractWithoutGasTest, Description: "Interact with existing contract.", FromAccount: "Main (0x TODO)", ToAccount: "", Contract: "", Amount: 0},
		{Name: "Test: Contract - not enough gas", test: sendContractWithoutGasTest, Description: "Attempt to submit contract without enough gas.", FromAccount: fmt.Sprintf("Main (%s)", mainAddress), ToAccount: "", Contract: "", Amount: 0},
		{Name: "Test: Contract - nonce too high", test: sendContractTestNonceTooHigh, Description: "Attempt to sumbit a contract with nonce too high.", FromAccount: fmt.Sprintf("Main (%s)", mainAddress), ToAccount: "", Contract: "", Amount: 0},
		{Name: "Test: Contract - nonce too low", test: sendContractTestNonceTooLow, Description: "Attempt to sumbit a contract with nonce too low.", FromAccount: fmt.Sprintf("Main (%s)", mainAddress), ToAccount: "", Contract: "", Amount: 0},
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "\U0001F344  {{ .Name | bgCyan | underline }}",
		Inactive: "   {{ .Name | cyan }}",
		Selected: "\U0001F344  {{ .Name | red | bold  }}",
		Details: banner + `
--------- Test Details ----------
{{ "Name:" | faint }}	{{ .Name }}
{{ "Description:" | faint }}	{{ .Description }}
{{ "To Account:" | faint }}	{{ .ToAccount }}
{{ "From Acount:" | faint }}	{{ .FromAccount }}
{{ "Contract:" | faint }}
{{ .Contract }}`,
	}

	searcher := func(input string, index int) bool {
		test := tests[index]
		name := strings.Replace(strings.ToLower(test.Name), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input)
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		prompt := promptui.Select{
			Label:     "Select the test to run",
			Items:     tests,
			Templates: templates,
			Size:      4,
			Searcher:  searcher,
		}

		i, _, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		tests[i].test()

		reader.ReadString('\n')
	}
}

// Consider passing in new test data type (struct with needed data)
func sendEthTest() {
	fmt.Println("Running Test")

	mainBalance := getBalance(mainAddress)
	hashRebelBalance := getBalance(hashRebelAddress)

	fmt.Println("Initial main account balance: ", mainBalance)
	fmt.Println("Initial Hash Rebel account balance: ", hashRebelBalance)

	// Get private key
	privateKey, fromAddress := getAddresses()

	nonce, err := ethClient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal("Error getting the next nonce: ", err)
	}

	// The ethereum blockchain expects wei so convert for the transaction
	amountWei := new(big.Int)
	amountWei.Mul(big.NewInt(1000000000000000000), big.NewInt(3))
	fmt.Println("Setting up amount of either to send in wei: ", amountWei)

	gasLimit := uint64(21000) // in units
	// Giving myself plenty of gas
	gasPrice := big.NewInt(5000000000000) // in wei (5000 gwei)
	fmt.Println("Setting up gas. gasLimit: ", amountWei, " gagisPrice: ", gasPrice)

	// Get the address to send the eth too
	toAddress := common.HexToAddress(hashRebelAddress)
	fmt.Println("Setting up the to address ", "address", hashRebelAddress)

	// Generate an unsigned transaction
	tx := types.NewTransaction(nonce, toAddress, amountWei, gasLimit, gasPrice, nil)
	fmt.Println("Setting up the unsigned transaction")

	signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, privateKey)
	if err != nil {
		log.Fatal("Error signing the transaction: ", err)
	}
	fmt.Println("Signing Transaction")

	err = ethClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal("Failed sending the transaction to the chain: ", err)
	}

	fmt.Printf("tx sent: %s \n", signedTx.Hash().Hex())

	fmt.Printf("************* Wait for the geth logs and then hit enter *************\n\n")
}

func sendContractTest() {
	fmt.Println("Running Test")

	// Get private key
	privateKey, fromAddress := getAddresses()

	// Get the next nonce
	nonce, err := ethClient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal("Error getting the next nonce: ", err)
	}
	gasPrice, err := ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	input := big.NewInt(1)
	address, tx, instance, err := simpleContract.DeploySimpleContract(auth, ethClient, input)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Contract address: ", address.Hex())
	fmt.Println("Tx Hash: ", tx.Hash().Hex())

	_ = instance

	fmt.Printf("************* Wait for the geth logs and then hit enter *************\n\n")
}

func sendContractTestNonceTooHigh() {
	fmt.Println("Running Test")

	// Get private key
	privateKey, fromAddress := getAddresses()

	// Get the next nonce
	nonce, err := ethClient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal("Error getting the next nonce: ", err)
	}
	gasPrice, err := ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce + 3))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	input := big.NewInt(1)
	address, tx, instance, err := simpleContract.DeploySimpleContract(auth, ethClient, input)

	if err != nil {
		fmt.Println("Yeah! An error occurred. Expected: nonce to high; Got:", err)
	}

	fmt.Println("Address and tx shouldn't exist ", address, tx)

	_ = instance

	fmt.Printf("************* Wait for the geth logs and then hit enter *************\n\n")
}

func sendContractTestNonceTooLow() {
	fmt.Println("Running Test")

	// Get private key
	privateKey, fromAddress := getAddresses()

	// Get the next nonce
	nonce, err := ethClient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal("Error getting the next nonce: ", err)
	}
	gasPrice, err := ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce - 20))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	input := big.NewInt(1)
	address, tx, instance, err := simpleContract.DeploySimpleContract(auth, ethClient, input)

	if err != nil {
		fmt.Println("Yeah! An error occurred. Expected: nonce to low; Got:", err)
	}

	fmt.Println("Address and tx shouldn't exist ", address, tx)

	_ = instance

	fmt.Printf("************* Wait for the geth logs and then hit enter *************\n\n")
}

func sendContractWithoutGasTest() {
	fmt.Println("Running Test")

	// Get private key
	privateKey, fromAddress := getAddresses()

	// Get the next nonce
	nonce, err := ethClient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal("Error getting the next nonce: ", err)
	}

	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // in wei
	auth.GasLimit = uint64(0)  // in units
	auth.GasPrice = big.NewInt(0)

	input := big.NewInt(1)
	address, tx, instance, err := simpleContract.DeploySimpleContract(auth, ethClient, input)
	if err != nil {
		fmt.Printf("************* WORKED *************\n\n", err)
	}

	fmt.Println("Contract address: ", address.Hex())
	fmt.Println("Tx Hash: ", tx.Hash().Hex())

	_ = instance

	fmt.Printf("************* Wait for the geth logs and then hit enter *************\n\n")
}

func getAddresses() (*ecdsa.PrivateKey, common.Address) {
	// Get private key
	privateKey, err := crypto.HexToECDSA(mainPrivateKey)
	if err != nil {
		log.Fatal("error getting private key: ", err)
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA: ", err)
	}
	fmt.Println("Setting up the private key")

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	return privateKey, fromAddress
}

func getBalance(address string) *big.Int {
	account := common.HexToAddress(address)
	balance, err := ethClient.BalanceAt(context.Background(), account, nil)

	if err != nil {
		fmt.Println("not able to get the balance: ", "address", address, "\nError: \n", err)
	}

	return balance
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initEthalyzer)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVarP(&logFile, "logfile", "l", "", "geth miner log output file (required)")
	rootCmd.PersistentFlags().StringVarP(&gethURI, "gethuri", "g", "http://localhost:8501", "URI to the geth node RPC Apis")
}

// initConfig reads in config file and ENV variables if set.
func initEthalyzer() {
	// TODO setup geth connection here. Also start reading from the log file here.
	var err error
	ethClient, err = ethclient.Dial(gethURI)
	if err != nil {
		log.Fatal("Unable to connect to: ", gethURI, "\nError: \n", err)
	}

	quit = make(chan bool)

	// Setup a go routine for watching the logs
	go monitorGethLogs()
}

func monitorGethLogs() {
	if len(logFile) == 0 {
		fmt.Println("No log file flag found.")
		return
	}
	// Open the file at the end to skip printing any left over logs
	location := tail.SeekInfo{Offset: 0, Whence: io.SeekEnd}
	logTail, err := tail.TailFile(logFile, tail.Config{Follow: true, ReOpen: true, Location: &location})
	if err != nil {
		log.Fatal("Unable to load the log file", err)
	}

	// Print new lines as they are added to the log file
	for {
		line := <-logTail.Lines
		if strings.Contains(line.Text, logStamp) {
			fmt.Println(line.Text)
		}
	}
}
