/*
Copyright © 2022 Justin Pulley <justinpulley@gitcoin.co>

*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/exec"
	"strings"
	"time"

	firebase "firebase.google.com/go"

	"github.com/ethereum/go-ethereum/crypto"
	signercore "github.com/ethereum/go-ethereum/signer/core"
	signerv4 "github.com/status-im/status-go/services/typeddata"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
	"google.golang.org/api/option"
)

// signCmd represents the sign command
var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}

		if err != nil {
			log.Fatal("Error loading .env file")
		}

		// .env .fun .composable .kinda!
		pKey := os.Getenv("PRIVATE_KEY")
		cName := os.Getenv("CONTRACT_NAME")
		cAddress := os.Getenv("CONTRACT_ADDRESS")
		cVersion := os.Getenv("CONTRACT_VERSION")
		signerPublic := os.Getenv("SIGNER_PUBLIC")
		directory := os.Getenv("DB_DIRECTORY")
		rpc := os.Getenv("WEB_SOCKET_RPC")

		privateKey, err := crypto.HexToECDSA(pKey)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Signing to all pending users with Msg:", message, "\n")

		ctx := context.Background()
		conf := &firebase.Config{
			DatabaseURL: "https://proofofstake-91004-default-rtdb.firebaseio.com/",
		}
		// Fetch the service account key JSON file contents
		opt := option.WithCredentialsFile("service-account-key.json")

		// Initialize the app with a service account, granting admin privileges
		app, err := firebase.NewApp(ctx, conf, opt)
		if err != nil {
			log.Fatalln("Error initializing app:", err)
		}

		client, err := app.Database(ctx)
		if err != nil {
			log.Fatalln("Error initializing database client:", err)
		}

		// Call our FB Realtime Database and return what matches the request query
		q := client.NewRef(directory).OrderByKey()

		ref := client.NewRef(directory)

		result, err := q.GetOrdered(ctx)
		if err != nil {
			log.Fatal(err)
		}

		var strSlice []string

		// Results will be logged in the increasing order of balance.
		for _, r := range result {
			var acc Signature

			if err := r.Unmarshal(&acc); err != nil {
				log.Fatal(err)
			}

			strSlice = append(strSlice, r.Key())

		}

		rClient, err := ethclient.Dial(rpc)
		if err != nil {
			log.Fatal(err)
		}

		contractAddress := common.HexToAddress(cAddress)
		query := ethereum.FilterQuery{
			// You could specify blocks here...
			/* FromBlock: big.NewInt(10485867),
			ToBlock:   big.NewInt(239420100), */

			Addresses: []common.Address{
				contractAddress,
			},
			Topics: [][]common.Hash{{common.HexToHash("0x5e91ea8ea1c46300eb761859be01d7b16d44389ef91e03a163a87413cbf55b95")}},
		}

		logs, err := rClient.FilterLogs(context.Background(), query)
		if err != nil {
			log.Fatal(err)
		}

		var s2 []string

		items := []donation{}

		box := donos{items}

		Counter := 0

		for _, vLog := range logs {

			test := common.HexToAddress(vLog.Topics[1].Hex()).String()

			if slices.Contains(s2, test) == true {
				/* fmt.Println("Multiple Donator:", test) */
			} else {

				toAppend := common.HexToAddress(vLog.Topics[1].Hex())

				s := fmt.Sprintf("%.18f", weiToEther(vLog.Topics[2].Big()))

				s2 = append(s2, toAppend.String())

				currentDono := donation{
					from:   toAppend.String(),
					amount: s,
					toSign: true,
				}

				box.AddItem(currentDono)

				Counter++
			}
		}

		Counter2 := 0

		// Check slice a (s2) against slice b (strSlice)
		for i := 0; i < len(strSlice); i++ {
			idx := slices.Contains(s2, strSlice[i])
			if idx {
				/* fmt.Println("Already Signed:", strSlice[i]) */
				Counter2++
				RemoveIndex(s2, slices.Index(s2, strSlice[i]))
			}
		}

		for i := 0; i < len(strSlice); i++ {
			result, key := isExists(strSlice[i], box.Items)

			if result {
				box.Items[key].toSign = false
			}
		}

		Counter3 := 0

		for i := 0; i < len(box.Items); i++ {
			if box.Items[i].toSign {

				signerData := signercore.TypedData{
					Types: signercore.Types{
						"signature": []signercore.Type{
							{Name: "sender", Type: "address"},
							{Name: "recipient", Type: "address"},
							{Name: "pledge", Type: "string"},
							{Name: "timestamp", Type: "string"},
							{Name: "msg", Type: "string"},
						},
						"EIP712Domain": []signercore.Type{
							{Name: "name", Type: "string"},
							{Name: "version", Type: "string"},
							{Name: "chainId", Type: "uint256"},
							{Name: "verifyingContract", Type: "address"},
						},
					},
					PrimaryType: "signature",
					Domain: signercore.TypedDataDomain{
						Name:              cName,
						Version:           cVersion,
						ChainId:           math.NewHexOrDecimal256(42),
						VerifyingContract: cAddress,
					},
					Message: signercore.TypedDataMessage{
						"sender":    signerPublic,
						"recipient": box.Items[i].from,
						"pledge":    box.Items[i].amount,
						"timestamp": fmt.Sprint(time.Now().Unix()),
						"msg":       message,
					},
				}

				dbData3 := signercore.TypedData{
					Types: signercore.Types{
						"signature": []signercore.Type{
							{Name: "sender", Type: "address"},
							{Name: "recipient", Type: "address"},
							{Name: "pledge", Type: "string"},
							{Name: "timestamp", Type: "string"},
							{Name: "msg", Type: "string"},
						},
					},
					PrimaryType: "signature",
					Domain: signercore.TypedDataDomain{
						Name:              cName,
						Version:           cVersion,
						ChainId:           math.NewHexOrDecimal256(42),
						VerifyingContract: cAddress,
					},
					Message: signercore.TypedDataMessage{
						"sender":    signerPublic,
						"recipient": box.Items[i].from,
						"pledge":    box.Items[i].amount,
						"timestamp": fmt.Sprint(time.Now().Unix()),
						"msg":       message,
					},
				}

				signed, err := signerv4.SignTypedDataV4(signerData, privateKey, big.NewInt(42))
				if err != nil {
					log.Fatal(err)
				}

				m := dbData3.Map()

				b, err := json.Marshal(m)
				if err != nil {
					panic(err)
				}

				os.WriteFile("/data.txt", b, 0644)

				// For more granular writes, open a file for writing.
				f, err := os.Create("./dat.txt")
				if err != nil {
					panic(err)
				}

				n2, err := f.Write(b)

				fmt.Printf("LZW Encoded %d bytes\n", n2)

				defer f.Close()

				command := "node parser.js"
				parts := strings.Fields(command)
				data, err := exec.Command(parts[0], parts[1:]...).Output()
				if err != nil {
					panic(err)
				}

				output := string(data)

				usersRef := ref.Child(box.Items[i].from)

				err2 := usersRef.Set(ctx, DbSignature{
					signercore.TypedData{
						Message: signerData.Message,
						Domain:  signerData.Domain,
					},
					signed.String(),
					// after properly encoding, we will put typeddata here where "message" lies rn.
					output,
				})
				if err2 != nil {
					log.Fatalln("Error setting value:", err)
				}
				Counter3++
			}
		}

		fmt.Println("Donation Events Total:", len(logs), "\n")
		fmt.Println("Unique Donation Events:", Counter, "\n")
		fmt.Println("Unique Signatures (DB):", Counter2+Counter3, "\n")
		fmt.Println("Sigs Generated This Run:", Counter3, "\n")

	},
}
var message string

type Signature struct {
	Message struct {
		Msg       string `json:"msg"`
		Pledge    string `json:"pledge"`
		Recipient string `json:"recipient"`
		Sender    string `json:"sender"`
		Timestamp string `json:"timestamp"`
	} `json:"message"`
	Signature string `json:"signature"`
	TypedData string `json:"typedData"`
}

type DbSignature struct {
	signercore.TypedData
	Signature2 string `json:"signature"`
	TypedData2 string `json:"typedData"`
}

type donation struct {
	from   string
	amount string
	toSign bool
}

type donos struct {
	Items []donation
}

func init() {
	rootCmd.AddCommand(signCmd)
	signCmd.Flags().StringVarP(&message, "message", "m", "", "Message to be signed")
	signCmd.MarkFlagRequired("message")
}

func RemoveIndex(slice []string, index int) []string {
	return append(slice[:index], slice[index+1:]...)
}

func removeDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

const (
	Wei   = 1
	GWei  = 1e9
	Ether = 1e18
)

func weiToEther(wei *big.Int) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(Ether))
}

func (donoBox *donos) AddItem(item donation) []donation {
	donoBox.Items = append(donoBox.Items, item)
	return donoBox.Items
}

func isExists(id string, box2 []donation) (result bool, rKey int) {
	result = false
	rKey = 0
	for key, donoor := range box2 {
		if donoor.from == id {
			result = true
			rKey = key
			break
		}
	}
	return result, rKey
}
