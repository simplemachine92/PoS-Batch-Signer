/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"log"
	"math/big"

	/* "net/url" */
	"os"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	signercore "github.com/ethereum/go-ethereum/signer/core"
	"github.com/joho/godotenv"
	signerv4 "github.com/status-im/status-go/services/typeddata"

	"github.com/spf13/cobra"
)

// signtestCmd represents the signtest command
var signtestCmd = &cobra.Command{
	Use:   "signtest",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		err := godotenv.Load(".env")

		if err != nil {
			log.Fatal("Error loading .env file")
		}

		pKey := os.Getenv("PRIVATE_KEY")

		privateKey, err := crypto.HexToECDSA(pKey)
		if err != nil {
			log.Fatal(err)
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
				Name:              "ProofOfStake_Pages",
				Version:           "0",
				ChainId:           math.NewHexOrDecimal256(4),
				VerifyingContract: "0x2d82DDb509E05a58067265d47f8fCd5e2857EFFE",
			},
			Message: signercore.TypedDataMessage{
				"sender":    "0xb010ca9Be09C382A9f31b79493bb232bCC319f01",
				"recipient": "0xb010ca9Be09C382A9f31b79493bb232bCC319f01",
				"pledge":    "0.13370000000000002",
				"timestamp": "1650839022516",
				"msg":       "good",
			},
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

		// Get typed data for api
		/* dataJson, err := json.Unmarshal(signerData)
		if err != nil {
			log.Fatal(err)
		} */

		/* var s Signature

		stringy, err := json.Unmarshal([]byte(signercore.TypedDataMessage), &s)

		fmt.Println("dis", signerData) */

		/* compressed := compressLZW(string(dataJson)) */

		/* formatted, err := signercore.
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("format", formatted) */

		/* lzwer := lzw.NewWriter(os.Stdout, 0, 8)



		this, err := lzwer.Write(signerData.)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("dis", this) */

		/* encoder := base64.NewEncoder(base64.StdEncoding, os.Stdout)

		encoder.Write() */

		/* fmt.Println("comp", compressed) */

		/* urlData := url.Values{"typedData": {string(dataJson[:])}}

		fmt.Println("enc", urlData.Encode()) */

		signed, err := signerv4.SignTypedDataV4(signerData, privateKey, big.NewInt(4))
		if err != nil {
			log.Fatal(err)
		}

		// Yeah that's a valid friggin signature, fricker
		fmt.Println("hope:", signed)

	},
}

func init() {
	rootCmd.AddCommand(signtestCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// signtestCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// signtestCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func compressLZW(testStr string) []int {
	code := 256
	dictionary := make(map[string]int)
	for i := 0; i < 256; i++ {
		dictionary[string(i)] = i
	}

	currChar := ""
	result := make([]int, 0)
	for _, c := range []byte(testStr) {
		phrase := currChar + string(c)
		if _, isTrue := dictionary[phrase]; isTrue {
			currChar = phrase
		} else {
			result = append(result, dictionary[currChar])
			dictionary[phrase] = code
			code++
			currChar = string(c)
		}
	}
	if currChar != "" {
		result = append(result, dictionary[currChar])
	}
	return result
}
