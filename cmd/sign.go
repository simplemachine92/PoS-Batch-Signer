/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"

	firebase "firebase.google.com/go"

	"google.golang.org/api/option"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
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
		if len(args) < 1 {
			log.Fatal("Username not specified")
		}
		address := args[0]

		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}

		fmt.Println("Signing to address:", address)
		fmt.Println("Message Input:", message)

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

		/* type Copy struct {
			key       string
			signature Signature
		} */

		// Call our FB Realtime Database and return what matches the request query
		q := client.NewRef("PoS").OrderByKey()

		result, err := q.GetOrdered(ctx)
		if err != nil {
			log.Fatal(err)
		}

		s := make([]string, len(result))
		// Results will be logged in the increasing order of balance.
		for _, r := range result {
			var acc Signature

			if err := r.Unmarshal(&acc); err != nil {
				log.Fatal(err)
			}
			/* log.Printf("%s => %v\n", r.Key(), acc) */

			// Put our address results in a slice, these are not comma separated like arrays
			s = append(s, r.Key())

		}

		// Print (later compare) after range function is completed and slice is populated
		/* log.Println("Slice", s) */

		rinkebyWS := os.Getenv("RINKEBY_WS")
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

		contractAddress := common.HexToAddress("0x3437030B6992Cd309e362269187a1b104DE0130E")
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

		// Slice to hold data from events
		s2 := make([]string, len(logs))

		for _, vLog := range logs {

			// Let's see if we have addresses, keeping as we may use this for operations later..
			/* fmt.Println("Pledgee", common.HexToAddress(vLog.Topics[1].Hex())) */

			toAppend := common.HexToAddress(vLog.Topics[1].Hex())

			s2 = append(s2, toAppend.String())

			// Grab pledge amount (in wei), log as string here, keeping as we may use this for operations later..
			/* fmt.Println("PValue", string(vLog.Topics[2].Big().String())) */

			/* domain, err := ens.ReverseResolve(mainnetClient, common.HexToAddress(vLog.Topics[1].Hex()))
			if err != nil {
				log.Print(err)
			} else {

				fmt.Println("User ENS", domain)
			} */
			/* pledgeeAddress := common.HexToAddress("0x3437030B6992Cd309e362269187a1b104DE0130E") */

			/* fmt.Println(([]common.Address(event.pledgee))) // foo
			fmt.Println([]*big.Int(event.pledgeValue))     // bar */

			/* var topics [3]string */

			/* fmt.Println("address (Pledgee):", common.HexToAddress(topics[0])) // 0xe79e73da417710ae99aa2088575580a60415d359acfad9cdd3382d59c80281d4 */
		}

		/* log.Println("before mod s1", s)
		log.Println("Before mod", s2) */

		// Check slice a (s) against slice b (s2)
		for i := 0; i < len(s2); i++ {
			idx := slices.Contains(s, s2[i])
			if idx {
				RemoveIndex(s2, i)
			}
		}

		// Omit duplicates from slice, and this is our "to sign" list
		log.Println("Slice after mod", removeDuplicateStr(s2))

	},
}
var message string

/* var creditAmount int64 */

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
