/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"encoding/hex"
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
	"github.com/ethereum/go-ethereum/common/hexutil"
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

		/* if len(args) < 1 {
			log.Fatal("Message not specified")
		} */
		/* address := args[0] */

		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}

		if err != nil {
			log.Fatal("Error loading .env file")
		}

		pKey := os.Getenv("PRIVATE_KEY")

		privateKey, err := crypto.HexToECDSA(pKey)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Signing to all pending users with Msg:", message)

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

		// Call our FB Realtime Database and return what matches the request query
		q := client.NewRef("PoS").OrderByKey()

		ref := client.NewRef("PoS")

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
			/* log.Printf("%s", r.Key())
			fmt.Println("sig", acc) */

			// Put our address results in a slice, these are not comma separated like arrays
			strSlice = append(strSlice, r.Key())

		}

		// Print (later compare) after range function is completed and slice is populated
		/* log.Println("Slice", strSlice) */

		rinkebyWS := os.Getenv("KOVAN_WS")
		/* uKey := os.Getenv("PRIVATE_KEY") */
		/* mainWS := os.Getenv("MAINNET_WS") */

		rClient, err := ethclient.Dial(rinkebyWS)
		if err != nil {
			log.Fatal(err)
		}

		/* mainnetClient, err := ethclient.Dial(mainWS)
		if err != nil {
			log.Fatal(err)
		} */

		contractAddress := common.HexToAddress("0x7Ea9411959fF856c1956f90b7569eDC3F0421c22")
		query := ethereum.FilterQuery{
			// FromBlock should make this a lot more efficient, don't forget to change..
			FromBlock: big.NewInt(10485867),
			ToBlock:   big.NewInt(239420100),
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

		for _, vLog := range logs {

			// Let's see if we have addresses, keeping, as we may use this for operations later..
			/* fmt.Println("Pledgee", common.HexToAddress(vLog.Topics[1].Hex())) */
			toAppend := common.HexToAddress(vLog.Topics[1].Hex())

			trimmed := common.TrimLeftZeroes(vLog.Topics[2].Bytes())

			encoded := hex.EncodeToString(trimmed)[1:]
			/* fmt.Println("encoded", encoded) */

			amount2, err := hexutil.DecodeBig("0x" + encoded)
			if err != nil {
				log.Fatal(err)
			}

			s := fmt.Sprintf("%.18f", weiToEther(amount2))

			/* fmt.Println(s)

			fmt.Println("ETH Amount", weiToEther(amount2)) */

			s2 = append(s2, toAppend.String())

			currentDono := donation{
				from:   toAppend.String(),
				amount: s,
				toSign: true,
			}

			box.AddItem(currentDono)

			/* fmt.Println("box", box.Items) */
		}

		// Check slice a (s) against slice b (s2)
		for i := 0; i < len(strSlice); i++ {
			idx := slices.Contains(s2, strSlice[i])
			/* log.Println("bool1", strSlice[i])
			log.Println("bool", idx) */
			if idx {
				log.Println("index", slices.Index(s2, strSlice[i]))
				RemoveIndex(s2, slices.Index(s2, strSlice[i]))
			}
		}

		// Omit duplicates from slice, and this is our "to sign" list
		/* log.Println("Slice after mod", removeDuplicateStr(s2)) */
		// This is database addresses vvv
		/* log.Println("strSlice", strSlice) */

		for i := 0; i < len(strSlice); i++ {
			result, key := isExists(strSlice[i], box.Items)

			if result {
				box.Items[key].toSign = false
			}
		}
		/* fmt.Println("b0x", box) */

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
						Name:              "ProofOfStake_Pages",
						Version:           "0",
						ChainId:           math.NewHexOrDecimal256(42),
						VerifyingContract: "0x522149fd0A0c8E2A2ffdb4dBeDB333e533Fbe2Ae",
					},
					Message: signercore.TypedDataMessage{
						"sender":    "0x55A178b6AfB3879F4a16c239A9F528663e7d76b3",
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
						Name:              "ProofOfStake_Pages",
						Version:           "0",
						ChainId:           math.NewHexOrDecimal256(42),
						VerifyingContract: "0x522149fd0A0c8E2A2ffdb4dBeDB333e533Fbe2Ae",
					},
					Message: signercore.TypedDataMessage{
						"sender":    "0x55A178b6AfB3879F4a16c239A9F528663e7d76b3",
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

				/* litw := 8 */

				m := dbData3.Map()

				b, err := json.Marshal(m)
				if err != nil {
					panic(err)
				}

				fmt.Println("json?", string(b))

				os.WriteFile("/data.txt", b, 0644)

				// For more granular writes, open a file for writing.
				f, err := os.Create("./dat.txt")
				if err != nil {
					panic(err)
				}

				n2, err := f.Write(b)

				fmt.Printf("wrote %d bytes\n", n2)

				// It's idiomatic to defer a `Close` immediately
				// after opening a file.
				defer f.Close()

				command := "node parser.js"
				parts := strings.Fields(command)
				fmt.Println("parts", parts[0], parts[1:])
				data, err := exec.Command(parts[0], parts[1:]...).Output()
				if err != nil {
					panic(err)
				}

				output := string(data)

				/* msgp, err := msgpack.Marshal(m)
				if err != nil {
					panic(err)
				}

				var data = []byte(msgp)

				fmt.Printf("input: %#v\n", string(data))

				var buf bytes.Buffer

				com := lzw.NewWriter(&buf, lzw.LSB, litw)

				w, err := com.Write(data)

				if err != nil {
					fmt.Println("write error:", err)
				}

				fmt.Println("wrote", w, "bytes")

				fmt.Println("buf")

				com.Close() */

				usersRef := ref.Child(box.Items[i].from)

				err2 := usersRef.Set(ctx, DbSignature{
					signercore.TypedData{
						Message: signerData.Message,
						/* Domain:  signerData.Domain, */
					},
					signed.String(),
					// after properly encoding, we will put typeddata here where "message" lies rn.
					/* b64.RawURLEncoding.EncodeToString(buf.Bytes()), */
					/* b64.URLEncoding.EncodeToString(buf.Bytes()), */
					output,
					/* b64.RawURLEncoding.WithPadding(buf.Bytes()) */
				})
				if err2 != nil {
					log.Fatalln("Error setting value:", err)
				}

			}
		}
	},
}
var message string

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
	/* signCmd.Flags().Int64VarP(&creditAmount, "amount", "a", 0, "Amount to be credited") */
	signCmd.MarkFlagRequired("message")
	/* signCmd.MarkFlagRequired("amount") */
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
