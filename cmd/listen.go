package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	firebase "firebase.google.com/go"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	signercore "github.com/ethereum/go-ethereum/signer/core"
	signerv4 "github.com/status-im/status-go/services/typeddata"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"google.golang.org/api/option"
)

// listenCmd represents the listen command
var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Listens and Signs",
	Long:  `Listens and Signs -l='your string' designates the string to be signed`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len([]rune(lmessage)) <= 1 {
			return errors.New("lmessage string is required")
		}

		if len([]rune(lmessage)) > 60 {
			return errors.New("lmessage string must be less than 60 characters")
		}
		return nil
	},
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
		signerPublic := os.Getenv("PUBLIC_KEY")
		cName := os.Getenv("CONTRACT_NAME")
		cAddress := os.Getenv("CONTRACT_ADDRESS")
		cVersion := os.Getenv("CONTRACT_VERSION")
		cChain := os.Getenv("CONTRACT_CHAIN")
		directory := os.Getenv("DB_DIRECTORY")
		rpc := os.Getenv("WEB_SOCKET_RPC")

		chainInt, err := strconv.ParseInt(cChain, 10, 64)

		cChainb := math.NewHexOrDecimal256(chainInt)

		privateKey, err := crypto.HexToECDSA(pKey)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Signing to new donors with Msg:", lmessage, "\n")

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

		ref := client.NewRef(directory)

		rClient, err := ethclient.Dial(rpc)
		if err != nil {
			log.Fatal(err)
		}

		contractAddress := common.HexToAddress(cAddress)
		query := ethereum.FilterQuery{
			Addresses: []common.Address{
				contractAddress,
			},
			Topics: [][]common.Hash{{common.HexToHash("0x5e91ea8ea1c46300eb761859be01d7b16d44389ef91e03a163a87413cbf55b95")}},
		}

		logs := make(chan types.Log)
		sub, err := rClient.SubscribeFilterLogs(context.Background(), query, logs)
		if err != nil {
			log.Fatal(err)
		}

		Counter := 0

		for {
			select {
			case err := <-sub.Err():
				log.Fatal(err)
			case vLog := <-logs:
				toAppend := common.HexToAddress(vLog.Topics[1].Hex())
				s := fmt.Sprintf("%.18f", weiToEther(vLog.Topics[2].Big()))

				currentDono := donation{
					from:   toAppend.String(),
					amount: s,
					toSign: true,
				}

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
						ChainId:           cChainb,
						VerifyingContract: cAddress,
					},
					Message: signercore.TypedDataMessage{
						"sender":    signerPublic,
						"recipient": currentDono.from,
						"pledge":    currentDono.amount,
						"timestamp": fmt.Sprint(time.Now().Unix()),
						"msg":       lmessage,
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
						ChainId:           cChainb,
						VerifyingContract: cAddress,
					},
					Message: signercore.TypedDataMessage{
						"sender":    signerPublic,
						"recipient": currentDono.from,
						"pledge":    currentDono.amount,
						"timestamp": fmt.Sprint(time.Now().Unix()),
						"msg":       lmessage,
					},
				}

				signed, err := signerv4.SignTypedDataV4(signerData, privateKey, big.NewInt(chainInt))
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

				usersRef := ref.Child(currentDono.from)

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
				Counter++
				fmt.Println("Signed and stored a message for address", currentDono.from, "\n")
				fmt.Println("Signed so far while listening:", Counter, "\n")
			}
		}
	},
}

var lmessage string

func init() {
	rootCmd.AddCommand(listenCmd)
	listenCmd.Flags().StringVarP(&lmessage, "lmessage", "l", "", "Listen Message to be signed")
	listenCmd.MarkFlagRequired("lmessage")
}
