package clients

import (
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"strconv"

	"github.com/NethermindEth/juno/core/felt"
)

type GatewayClient struct {
	baseUrl string
}

func NewGatewayClient(baseUrl string) *GatewayClient {
	return &GatewayClient{
		baseUrl: baseUrl,
	}
}

// `buildQueryString` builds the query url with encoded parameters
func (c *GatewayClient) buildQueryString(endpoint string, args map[string]string) string {
	base, err := url.Parse(c.baseUrl)
	if err != nil {
		panic("Malformed feeder gateway base URL")
	}

	base.Path += endpoint

	params := url.Values{}
	for k, v := range args {
		params.Add(k, v)
	}
	base.RawQuery = params.Encode()

	return base.String()
}

// get performs a "GET" http request with the given URL and returns the response body
func (c *GatewayClient) get(queryUrl string) ([]byte, error) {
	res, err := http.Get(queryUrl)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	return body, err
}

// StateUpdate object returned by the gateway in JSON format for "get_state_update" endpoint
type StateUpdate struct {
	BlockHash *felt.Felt `json:"block_hash"`
	NewRoot   *felt.Felt `json:"new_root"`
	OldRoot   *felt.Felt `json:"old_root"`

	StateDiff struct {
		StorageDiffs map[string][]struct {
			Key   *felt.Felt `json:"key"`
			Value *felt.Felt `json:"value"`
		} `json:"storage_diffs"`

		Nonces            interface{} `json:"nonces"` // todo: define
		DeployedContracts []struct {
			Address   *felt.Felt `json:"address"`
			ClassHash *felt.Felt `json:"class_hash"`
		} `json:"deployed_contracts"`
		DeclaredContracts interface{} `json:"declared_contracts"` // todo: define
	} `json:"state_diff"`
}

func (c *GatewayClient) GetStateUpdate(blockNumber uint64) (*StateUpdate, error) {
	queryUrl := c.buildQueryString("get_state_update", map[string]string{
		"blockNumber": strconv.FormatUint(blockNumber, 10),
	})

	body, err := c.get(queryUrl)
	update := new(StateUpdate)
	if err = json.Unmarshal(body, update); err != nil {
		return nil, err
	}

	return update, nil
}

// Transaction object returned by the gateway in JSON format for multiple endpoints
type Transaction struct {
	Hash                *felt.Felt   `json:"transaction_hash"`
	Version             *felt.Felt   `json:"version"`
	ContractAddress     *felt.Felt   `json:"contract_address"`
	ContractAddressSalt *felt.Felt   `json:"contract_address_salt"`
	ClassHash           *felt.Felt   `json:"class_hash"`
	ConstructorCalldata []*felt.Felt `json:"constructor_calldata"`
	Type                string       `json:"type"`
	// invoke
	MaxFee             *felt.Felt   `json:"max_fee"`
	Signature          []*felt.Felt `json:"signature"`
	Calldata           []*felt.Felt `json:"calldata"`
	EntryPointSelector *felt.Felt   `json:"entry_point_selector"`
	// declare/deploy_account
	Nonce *felt.Felt `json:"nonce"`
	// declare
	SenderAddress *felt.Felt `json:"sender_address"`
}

type TransactionStatus struct {
	Status           string       `json:"status"`
	BlockHash        *felt.Felt   `json:"block_hash"`
	BlockNumber      *big.Int     `json:"block_number"`
	TransactionIndex *big.Int     `json:"transaction_index"`
	Transaction      *Transaction `json:"transaction"`
}

func (c *GatewayClient) GetTransaction(transactionHash *felt.Felt) (*TransactionStatus, error) {
	queryUrl := c.buildQueryString("get_transaction", map[string]string{
		"transactionHash": "0x" + transactionHash.Text(16),
	})

	body, err := c.get(queryUrl)
	txStatus := new(TransactionStatus)
	if err = json.Unmarshal(body, txStatus); err != nil {
		return nil, err
	}

	return txStatus, nil
}
