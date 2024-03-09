// filePath: satmine/mrc20.go

package satmine

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"golang.org/x/net/html"

	jsoniter "github.com/json-iterator/go"
)

// MRC721Protocol defines the structure for the MRC-721 genesis miner inscription protocol.
type MRC721Protocol struct {
	P     string   `json:"p"`
	Miner Miner    `json:"miner"`
	Token Token    `json:"token"`
	Ltry  *Lottery `json:"ltry,omitempty"` // Pointer to allow the field to be empty
	Burn  *Burn    `json:"burn,omitempty"` // Pointer to allow the field to be empty
}

// Miner defines the structure for miner information in MRC-721.
type Miner struct {
	Name string `json:"name"`
	Max  string `json:"max"`
	Lim  string `json:"lim"`
}

// GetUpperName returns the miner's name in uppercase.
func (m *Miner) GetUpperName() string {
	return strings.ToUpper(m.Name)
}

// Token defines the token parameters in MRC-721.
type Token struct {
	Tick  string `json:"tick"`
	Total string `json:"total"`
	Beg   string `json:"beg"`
	Halv  string `json:"halv"`
	Dcr   string `json:"dcr"`
}

func (t *Token) GetLowerTick() string {
	return strings.ToLower(t.Tick)
}

// Lottery defines the lottery parameters in MRC-721.
type Lottery struct {
	Pool  string `json:"pool"`
	Intvl string `json:"intvl"`
	Winp  string `json:"winp"`
	Dist  string `json:"dist"`
}

// Stake defines the staking parameters in MRC-721.
type Burn struct {
	Unit  string `json:"unit"`
	Boost string `json:"boost"`
}

// MRC20Protocol defines the structure for the MRC-20 token transfer protocol.
type MRC20Protocol struct {
	P    string  `json:"p"`
	Op   string  `json:"op"`
	Tick string  `json:"tick"`
	Amt  string  `json:"amt"`
	Dec  *string `json:"dec,omitempty"` // Pointer to allow the field to be empty
	Insc *string `json:"insc,omitempty"`
}

// ParseMRC721Protocol parses the MRC721Protocol data from a byte slice.
func ParseMRC721Protocol(data []byte) (*MRC721Protocol, error) {
	var protocol MRC721Protocol
	err := jsoniter.Unmarshal(data, &protocol)
	if err != nil {
		return nil, err
	}

	// Trim leading and trailing whitespace from all string fields
	protocol.P = strings.TrimSpace(protocol.P)
	// if protocol.P == "test-888" {
	// 	protocol.P = "mrc-721"
	// }

	protocol.Miner.Name = strings.TrimSpace(protocol.Miner.Name)
	protocol.Miner.Max = strings.TrimSpace(protocol.Miner.Max)
	protocol.Miner.Lim = strings.TrimSpace(protocol.Miner.Lim)
	protocol.Token.Tick = strings.TrimSpace(protocol.Token.Tick)
	protocol.Token.Total = strings.TrimSpace(protocol.Token.Total)
	protocol.Token.Beg = strings.TrimSpace(protocol.Token.Beg)
	protocol.Token.Halv = strings.TrimSpace(protocol.Token.Halv)
	protocol.Token.Dcr = strings.TrimSpace(protocol.Token.Dcr)
	if protocol.Ltry != nil {
		protocol.Ltry.Pool = strings.TrimSpace(protocol.Ltry.Pool)
		protocol.Ltry.Intvl = strings.TrimSpace(protocol.Ltry.Intvl)
		protocol.Ltry.Winp = strings.TrimSpace(protocol.Ltry.Winp)
		protocol.Ltry.Dist = strings.TrimSpace(protocol.Ltry.Dist)
	}
	if protocol.Burn != nil {
		protocol.Burn.Unit = strings.TrimSpace(protocol.Burn.Unit)
		protocol.Burn.Boost = strings.TrimSpace(protocol.Burn.Boost)
	}

	// Modification of SATMINE data for community reasons
	// https://twitter.com/SatMineOfficial/status/1744347847208702031
	// Due to an unfortunate error in SATMINE's initial inscription process, an extra 32% of the supply was inscribed. Our indexer currently does not recognize these NFTs.
	// We fully value our community's input, so the handling of these additional NFTs will be put to a poll. We can either 1) continue to disregard these extra NFTs, or 2) include them and increase the total supply of SATMINE to 13263. Voting will last 24 hours.
	// Disregard the extra NFTs 34.1%    Include the extra NFTs  65.9%
	if protocol.Miner.GetUpperName() == "SATMINE" {
		protocol.Miner.Max = "13263"
		//protocol.Miner.Max = "20000"
		protocol.Miner.Lim = "100"
	}

	return &protocol, nil
}

func HtmlToNameID(data []byte) (mrc721name string, mrc721ID string, err error) {
	doc, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return "", "", fmt.Errorf("error parsing HTML: %v", err)
	}

	//var mrc721name, mrc721ID string
	var parseNode func(*html.Node)
	parseNode = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "body" {
			for _, attr := range n.Attr {
				if attr.Key == "name" {
					mrc721name = attr.Val
				} else if attr.Key == "mrc-721" {
					mrc721ID = attr.Val
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			parseNode(c)
		}
	}

	parseNode(doc)

	if mrc721name == "" || mrc721ID == "" {
		return "", "", errors.New("required attributes not found")
	}

	return mrc721name, mrc721ID, nil

}

// HtmlToImgSrc parses the given HTML byte data to find the first <img> tag and return the value of its src attribute.
func HtmlToImgSrc(data []byte) (imgSrc string, err error) {
	doc, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("error parsing HTML: %v", err) // Error parsing the HTML data
	}

	// Define a recursive function to traverse HTML nodes
	var parseNode func(*html.Node)
	parseNode = func(n *html.Node) {
		// If the node is an Element node and its tag name is img
		if n.Type == html.ElementNode && n.Data == "img" {
			for _, attr := range n.Attr {
				if attr.Key == "src" {
					imgSrc = attr.Val // Assign the src attribute value to imgSrc
					return            // Return immediately after finding the src attribute
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			parseNode(c) // Recursively search for child nodes
		}
	}

	parseNode(doc) // Start parsing from the root node

	if imgSrc == "" {
		return "", fmt.Errorf("img src attribute not found") // Return an error if the src attribute is not found
	}

	return imgSrc, nil
}

// ParseMRC721HtmlProtocol parses the MRC721Protocol data from an HTML document.
func ParseMRC721HtmlProtocol(txn *badger.Txn, data []byte) (*MRC721Protocol, error) {
	// doc, err := html.Parse(bytes.NewReader(data))
	// if err != nil {
	// 	return nil, fmt.Errorf("error parsing HTML: %v", err)
	// }

	// var mrc721name, mrc721ID string
	// var parseNode func(*html.Node)
	// parseNode = func(n *html.Node) {
	// 	if n.Type == html.ElementNode && n.Data == "body" {
	// 		for _, attr := range n.Attr {
	// 			if attr.Key == "name" {
	// 				mrc721name = attr.Val
	// 			} else if attr.Key == "mrc-721" {
	// 				mrc721ID = attr.Val
	// 			}
	// 		}
	// 	}
	// 	for c := n.FirstChild; c != nil; c = c.NextSibling {
	// 		parseNode(c)
	// 	}
	// }

	// parseNode(doc)

	mrc721name, mrc721ID, err := HtmlToNameID(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %v", err)
	}

	if mrc721name == "" || mrc721ID == "" {
		return nil, errors.New("required attributes not found")
	}

	// if mrc721name == "test-888" {
	// 	mrc721name = "mrc-721"
	// }

	//fmt.Printf("Name: %s, MRC-721: %s\n", mrc721name, mrc721ID)

	// Retrieve Mrc721GenesisData using mrc721name
	geninscKey := fmt.Sprintf("mrc721::geninsc::%s", mrc721name)
	item, err := txn.Get([]byte(geninscKey))
	if err != nil {
		return nil, fmt.Errorf("error retrieving Mrc721GenesisData: %v", err)
	}

	var genesisData Mrc721GenesisData
	err = item.Value(func(val []byte) error {
		return jsoniter.Unmarshal(val, &genesisData)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal Mrc721GenesisData: %v", err)
	}

	// Retrieve HookInscription using genesisData.ID
	hookInscrKey := fmt.Sprintf("inscr::%s", genesisData.ID)
	item, err = txn.Get([]byte(hookInscrKey))
	if err != nil {
		return nil, fmt.Errorf("error retrieving HookInscription: %v", err)
	}

	var hookInscription HookInscription
	err = item.Value(func(val []byte) error {
		return jsoniter.Unmarshal(val, &hookInscription)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal HookInscription: %v", err)
	}

	// Parse MRC721Protocol from HookInscription
	firstMrc721, err := ParseMRC721Protocol(*hookInscription.ContentByte)
	if err != nil {
		return nil, fmt.Errorf("failed to parse MRC721 protocol: %v", err)
	}

	return firstMrc721, nil
}

// ParseMRC20Protocol parses the MRC20Protocol data from a byte slice.
func ParseMRC20Protocol(data []byte) (*MRC20Protocol, error) {
	var protocol MRC20Protocol
	err := jsoniter.Unmarshal(data, &protocol)
	if err != nil {
		return nil, err
	}
	// Trim leading and trailing whitespace from all string fields
	protocol.P = strings.TrimSpace(protocol.P)
	protocol.Op = strings.TrimSpace(protocol.Op)
	protocol.Tick = strings.TrimSpace(protocol.Tick)
	protocol.Amt = strings.TrimSpace(protocol.Amt)
	if protocol.Dec != nil {
		trimmedDec := strings.TrimSpace(*protocol.Dec)
		protocol.Dec = &trimmedDec
	}

	return &protocol, nil
}

// ValidateProtocolData validates the protocol data for either MRC-20, MRC-721, or MRC-721 HTML.
// It checks if all required fields in the protocol have a non-zero string length.
// Returns a boolean indicating validity, the protocol name, and an error if any.
func ValidateProtocolData(data []byte) (bool, string, error) {
	// Trim leading and trailing whitespace from the data
	trimmedData := strings.TrimSpace(string(data))
	//fmt.Println("---trimmedData=", trimmedData)

	// Check if the trimmed data starts with "<!DOCTYPE html>", which indicates HTML data.
	if strings.HasPrefix(trimmedData, "<!DOCTYPE html>") {
		return validateMRC721HtmlData([]byte(trimmedData))
	}
	if strings.HasPrefix(trimmedData, "<html>") {
		return validateMRC721HtmlData([]byte(trimmedData))
	}

	var p struct {
		P string `json:"p"`
	}

	// Decode the JSON data to get the protocol type.
	err := jsoniter.Unmarshal(data, &p)
	if err != nil {
		return false, "", err
	}

	// Choose the validation function based on the protocol type.
	switch p.P {
	case "mrc-721":
		return validateMRC721Data(data)
	case "mrc-20":
		return validateMRC20Data(data)
	default:
		return false, "", errors.New("unknown protocol")
	}
}

// validateMRC721HtmlData validates the MRC-721 HTML data.
// It checks for the presence of the 'body' tag with 'name' and 'mrc-721' attributes,
// and an 'img' tag with a non-zero length 'src' attribute.
func validateMRC721HtmlData(data []byte) (bool, string, error) {
	doc, err := html.Parse(strings.NewReader(string(data)))
	if err != nil {
		return false, "", err
	}
	var bodyFound, nameFound, mrc721Found, imgFound, srcFound bool
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "body" {
			bodyFound = true
			for _, a := range n.Attr {
				if a.Key == "name" && len(a.Val) > 0 {
					nameFound = true
				}
				if a.Key == "mrc-721" && len(a.Val) > 0 {
					mrc721Found = true
				}
			}
		}
		if n.Type == html.ElementNode && n.Data == "img" {
			imgFound = true
			for _, a := range n.Attr {
				if a.Key == "src" && len(a.Val) > 0 {
					srcFound = true
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
	if bodyFound && nameFound && mrc721Found && imgFound && srcFound {
		return true, "mrc-721html", nil
	}
	return false, "", errors.New("validation failed")
}

// It uses ParseMRC721Protocol to parse the byte slices. If parsing fails or any field is not equal, it returns false.
// If all fields are equal, it returns true.
func IsEqual721DataByte(a, b *[]byte) bool {
	// Parse the first byte slice using ParseMRC721Protocol
	protocolA, err := ParseMRC721Protocol(*a)
	if err != nil {
		return false
	}

	// Parse the second byte slice using ParseMRC721Protocol
	protocolB, err := ParseMRC721Protocol(*b)
	if err != nil {
		return false
	}

	// Compare each field of the protocol structs
	if protocolA.P != protocolB.P ||
		protocolA.Miner.Name != protocolB.Miner.Name ||
		protocolA.Miner.Max != protocolB.Miner.Max ||
		protocolA.Miner.Lim != protocolB.Miner.Lim ||
		protocolA.Token.Tick != protocolB.Token.Tick ||
		protocolA.Token.Total != protocolB.Token.Total ||
		protocolA.Token.Beg != protocolB.Token.Beg ||
		protocolA.Token.Halv != protocolB.Token.Halv ||
		protocolA.Token.Dcr != protocolB.Token.Dcr {
		return false
	}

	// Check Lottery fields if present
	if (protocolA.Ltry != nil && protocolB.Ltry != nil) &&
		(protocolA.Ltry.Pool != protocolB.Ltry.Pool ||
			protocolA.Ltry.Intvl != protocolB.Ltry.Intvl ||
			protocolA.Ltry.Winp != protocolB.Ltry.Winp ||
			protocolA.Ltry.Dist != protocolB.Ltry.Dist) {
		return false
	}

	// Check Burn fields if present
	if (protocolA.Burn != nil && protocolB.Burn != nil) &&
		(protocolA.Burn.Unit != protocolB.Burn.Unit ||
			protocolA.Burn.Boost != protocolB.Burn.Boost) {
		return false
	}

	// All fields are equal
	return true
}

// It uses ParseMRC721Protocol to parse the byte slices. If parsing fails or any field is not equal, it returns false.
// If all fields are equal, it returns true.
func IsEqual721Data(a, b *HookInscription) bool {
	// Parse the first byte slice using ParseMRC721Protocol
	if IsEqual721DataByte(a.ContentByte, b.ContentByte) {
		return true
	}

	protocolA, err := ParseMRC721Protocol(*a.ContentByte)
	if err != nil {
		return false
	}

	//mrc721html,err :=  ParseMRC721HtmlProtocol(txn *badger.Txn, data []byte)
	mrc721name, mrc721ID, err := HtmlToNameID(*b.ContentByte)
	if err != nil {
		return false
	}
	// mrc721name ToUpper
	mrc721name = strings.ToUpper(mrc721name)
	if protocolA.Miner.GetUpperName() == mrc721name && a.ID == mrc721ID {
		return true
	}

	return false
}

// validateMRC20Data validates the MRC-20 protocol data.
func validateMRC20Data(data []byte) (bool, string, error) {
	var protocol MRC20Protocol
	err := jsoniter.Unmarshal(data, &protocol)
	if err != nil {
		return false, "mrc-20", err
	}
	//fmt.Println("validateMRC20Data protocol=", protocol)

	// Validate the 'P' field to be exactly "mrc-20".
	if protocol.P != "mrc-20" {
		return false, "mrc-20", errors.New("invalid protocol type")
	}

	// Validate the 'Op' field to be either "transfer" or "burn".
	if protocol.Op != "transfer" && protocol.Op != "burn" {
		return false, "mrc-20", errors.New("invalid operation type")
	}

	// Validate the 'Tick' field to be lowercase and have a length <= 4.
	if len(protocol.Tick) > 4 || protocol.Tick != strings.ToLower(protocol.Tick) {
		return false, "mrc-20", errors.New("invalid ticker format")
	}

	// Validate the 'Amt' field to be a valid big.Int and have a length <= 666.
	if len(protocol.Amt) > 666 {
		return false, "mrc-20", errors.New("amount exceeds maximum length")
	}
	_, ok := new(big.Int).SetString(protocol.Amt, 10)
	if !ok {
		return false, "mrc-20", errors.New("invalid amount format")
	}

	// Validate the 'Dec' field to be either nil or "8".
	if protocol.Dec != nil && *protocol.Dec != "8" {
		return false, "mrc-20", errors.New("invalid decimal value")
	}

	return true, "mrc-20", nil
}

// // validateMRC721Data validates the MRC-721 protocol data.
// func validateMRC721Data(data []byte) (bool, string, error) {
// 	var protocol MRC721Protocol
// 	err := jsoniter.Unmarshal(data, &protocol)
// 	if err != nil {
// 		logger.Error("MRC721Protocol err: ", zap.Error(err))
// 		return false, "mrc-721", err
// 	}

// 	return true, "mrc-721", nil
// }

// validateMRC721Data validates the MRC-721 protocol data against specific rules.
func validateMRC721Data(data []byte) (bool, string, error) {
	var protocol MRC721Protocol
	err := jsoniter.Unmarshal(data, &protocol)
	if err != nil {
		return false, "mrc-721", err
	}

	// Validate the 'P' field
	if protocol.P != "mrc-721" {
		return false, "mrc-721", errors.New("protocol P must be 'mrc-721'")
	}

	// Validate Token
	if err := validateToken(protocol.Token); err != nil {
		fmt.Println("validateMRC721Data err=", err)
		return false, "mrc-721", err
	}

	// Validate Miner
	if err := validateMiner(protocol.Miner); err != nil {
		return false, "mrc-721", err
	}

	// Validate Lottery if it's not nil
	if protocol.Ltry != nil {
		if err := validateLottery(*protocol.Ltry); err != nil {
			return false, "mrc-721", err
		}
	}

	// Validate Burn if it's not nil
	if protocol.Burn != nil {
		if err := validateBurn(*protocol.Burn, protocol.Token.Total); err != nil {
			return false, "mrc-721", err
		}
	}

	return true, "mrc-721", nil
}

// validateMiner checks if the Miner structure meets the defined requirements.
func validateMiner(miner Miner) error {
	maxBigInt, ok := big.NewInt(0).SetString(miner.Max, 10)
	if !ok || maxBigInt.Cmp(big.NewInt(1)) == -1 || maxBigInt.Cmp(big.NewInt(100000000)) == 1 {
		return errors.New("miner Max must be a big.Int between 1 and 100000000")
	}

	limBigInt, ok := big.NewInt(0).SetString(miner.Lim, 10)
	if !ok || limBigInt.Cmp(big.NewInt(1)) == -1 || limBigInt.Cmp(maxBigInt) == 1 {
		return errors.New("miner Lim must be a big.Int between 1 and Max value")
	}

	return nil
}

// validateLottery checks if the Lottery structure meets the defined requirements.
func validateLottery(ltry Lottery) error {
	// Validate Pool, Winp, Dist
	if err := validatePercentageField(ltry.Pool); err != nil {
		return err
	}
	if err := validatePercentageField(ltry.Winp); err != nil {
		return err
	}
	if err := validatePercentageField(ltry.Dist); err != nil {
		return err
	}

	// Validate Intvl
	intvlBigInt, ok := big.NewInt(0).SetString(ltry.Intvl, 10)
	if !ok || intvlBigInt.Cmp(big.NewInt(1)) == -1 || intvlBigInt.Cmp(big.NewInt(100000000)) == 1 {
		return errors.New("lottery Intvl must be a big.Int between 1 and 100000000")
	}

	return nil
}

// validateBurn checks if the Burn structure meets the defined requirements.
func validateBurn(burn Burn, max string) error {
	// Validate Unit
	maxBigInt, _ := big.NewInt(0).SetString(max, 10) // Already validated in validateMiner
	unitBigInt, ok := big.NewInt(0).SetString(burn.Unit, 10)

	if !ok || unitBigInt.Cmp(big.NewInt(1)) == -1 || unitBigInt.Cmp(maxBigInt) == 1 {
		return errors.New("burn Unit must be a big.Int between 1 and Max value")
	}

	// Validate Boost
	if err := validatePercentageField(burn.Boost); err != nil {
		return err
	}

	return nil
}

// validatePercentageField checks if a string field can be converted to a percentage big.Int between 0 and 1000.
func validatePercentageField(field string) error {
	if len(field) > 5 {
		return errors.New("field value length must not exceed 5 characters")
	}
	resultBigInt := stringToPercentageBigInt(field)
	if resultBigInt.Cmp(big.NewInt(0)) == -1 || resultBigInt.Cmp(big.NewInt(1000)) == 1 {
		return errors.New("field value must be between 0 and 1000 after conversion")
	}

	return nil
}

// validateToken checks if the Token structure meets the defined requirements.
func validateToken(token Token) error {
	// Check if Tick is not more than 4 characters
	if len(token.Tick) > 4 {
		return errors.New("token Tick must be a string with a maximum of 4 characters")
	}

	// Determine if token.Tick is all lowercase
	if token.Tick != strings.ToLower(token.Tick) {
		return errors.New("token Tick must be lowercase")
	}

	// Check if Total is a valid big.Int and does not exceed 100 characters
	totalBigInt, ok := big.NewInt(0).SetString(token.Total, 10)
	if !ok || len(token.Total) > 100 {
		return errors.New("token Total must be a string convertible to big.Int and not exceed 100 characters")
	}

	// Check if Beg is a valid big.Int, does not exceed 100 characters, and is less than or equal to Total
	begBigInt, ok := big.NewInt(0).SetString(token.Beg, 10)
	if !ok || len(token.Beg) > 100 || begBigInt.Cmp(totalBigInt) == 1 {
		return errors.New("token Beg must be a string convertible to big.Int, not exceed 100 characters, and be less than or equal to Total")
	}

	// Validate Halv as a big.Int between 1 and 100000000
	halvBigInt, ok := big.NewInt(0).SetString(token.Halv, 10)
	if !ok || halvBigInt.Cmp(big.NewInt(1)) == -1 || halvBigInt.Cmp(big.NewInt(100000000)) == 1 {
		return errors.New("token Halv must be a string convertible to big.Int and between 1 and 100000000")
	}

	// Validate Dcr using validatePercentageField
	if err := validatePercentageField(token.Dcr); err != nil {
		return fmt.Errorf("token Dcr validation error: %s", err)
	}

	return nil
}
