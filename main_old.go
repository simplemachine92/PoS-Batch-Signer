package none

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
	"github.com/wealdtech/go-ens/v3"
	// for demo
)

func old() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	rinkebyWS := os.Getenv("RINKEBY_WS")
	mainWS := os.Getenv("MAINNET_WS")

	rClient, err := ethclient.Dial(rinkebyWS)
	if err != nil {
		log.Fatal(err)
	}

	mainnetClient, err := ethclient.Dial(mainWS)
	if err != nil {
		log.Fatal(err)
	}

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

	/* pledge, err := stake.NewStakeFilterer(contractAddress, stake.fil) */
	/* []common.Hash{stake.NewStakeFilterer(stake.StakeABI)} */

	/* contractAbi, err := abi.JSON(strings.NewReader(string(stake.StakeABI)))
	if err != nil {
		log.Fatal(err)
	} */

	for _, vLog := range logs {
		/* fmt.Println(vLog.BlockHash.Hex()) // 0x3404b8c050aa0aacd0223e91b5c32fee6400f357764771d0684fa7b3f448f1a8
		fmt.Println(vLog.BlockNumber)     // 2394201
		fmt.Println(vLog.TxHash.Hex()) */ // 0x280201eda63c9ff6f305fcee51d5eb86167fab40ca3108ec784e8652a0e2b1a6

		/* event := struct {
			pledgee     []common.Address
			pledgeValue []*big.Int
		}{}
		err := contractAbi.UnpackIntoInterface(&event, "pledge", vLog.Topics[1][:])
		// Pledge was case sensitive :weary:
		if err != nil {
			log.Fatal(err)
		} */

		// Let's see if we have addresses, keeping as we may use this for operations later..
		fmt.Println("Pledgee", common.HexToAddress(vLog.Topics[1].Hex()))

		// Grab pledge amount (in wei), log as string here, keeping as we may use this for operations later..
		fmt.Println("PValue", string(vLog.Topics[2].Big().String()))

		domain, err := ens.ReverseResolve(mainnetClient, common.HexToAddress(vLog.Topics[1].Hex()))
		if err != nil {
			log.Print(err)
		} else {

			fmt.Println("User ENS", domain)
		}
		/* pledgeeAddress := common.HexToAddress("0x3437030B6992Cd309e362269187a1b104DE0130E") */

		/* fmt.Println(([]common.Address(event.pledgee))) // foo
		fmt.Println([]*big.Int(event.pledgeValue))     // bar */

		/* var topics [3]string */

		/* fmt.Println("address (Pledgee):", common.HexToAddress(topics[0])) // 0xe79e73da417710ae99aa2088575580a60415d359acfad9cdd3382d59c80281d4 */
	}

	/* eventSignature := []byte("Pledge(bytes20, uint)")
	hash := crypto.Keccak256Hash(eventSignature)
	fmt.Println(hash.Hex()) // 0xe79e73da417710ae99aa2088575580a60415d359acfad9cdd3382d59c80281d4 */
}
