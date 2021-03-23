package smWallet

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/spacemeshos/ed25519"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/common/util"
	"github.com/spf13/viper"

	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/pbkdf2"
)

// Pub Key = 64 bytes
// Prv Key = 128 bytes

// address : 0x7fa75881ca0050028b32f424f860e3a73d4bf168
//           0xf6103aadbba77d2324fc4fad66eaa971bb5b8402

func initViper(t *testing.T) {
	viper.SetConfigFile("./.private.json")
	if err := viper.ReadInConfig(); err != nil {
		t.Fatal("Could not read config file")
	}
}

func TestReadWallet(t *testing.T) {
	initViper(t)
	keystore := viper.GetString("w1")
	smData, err := LoadWallet(keystore)
	if err != nil {
		t.Fatal(err, keystore)
	}
	t.Log(smData.Meta.DisplayName)
	t.Log(smData.Meta.Created)
	t.Log(smData.Meta.NetID)
	t.Log(smData.Meta.Meta.Salt)
	t.Log(smData.Crypto.Cipher)
	t.Fail()
}

//                       0x865330189761187daa2243a1533b0412b8e14613
// a7ce2fc6147c3c8d2610ce83865330189761187daa2243a1533b0412b8e14613
func chkTErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestSmApp(t *testing.T) {
	testUnlock(t, "smApp2020.json", "<<password>>")
}

func testUnlock(t *testing.T, keystore string, password string) {
	t.Log("opening :", keystore)
	//
	smData, err := LoadWallet(keystore)
	chkTErr(t, err)
	t.Log("displayName", smData.Meta.DisplayName)
	t.Log("created", smData.Meta.Created)
	err = smData.Unlock(password) // )
	chkTErr(t, err)
	mnenonic, err := smData.GetMnemonic()
	chkTErr(t, err)
	numberOfAccounts, err := smData.GetNumberOfAccounts()
	chkTErr(t, err)
	t.Log("mnemonic", mnenonic)
	t.Log(numberOfAccounts, "entries")
	for accountNo := 0; accountNo < numberOfAccounts; accountNo++ {
		display, err := smData.GetAccountDisplayName(accountNo)
		chkTErr(t, err)
		t.Log("name", display)
		pubKey, err := smData.GetPublicKey(accountNo)
		chkTErr(t, err)
		t.Log("pubkey", pubKey)
		address, err := smData.GetAddress(accountNo)
		chkTErr(t, err)
		t.Log("address", address.Hex())
		t.Log("pk", smData.Crypto.confidential.Accounts[accountNo].SecretKey)
	}
	t.Log(len(smData.Crypto.confidential.Contacts), " contacts")
	for n, ctc := range smData.Crypto.confidential.Contacts {
		t.Log(n, ctc)
	}

	ciphertext, err := hex.DecodeString(smData.Crypto.CipherText)
	if err != nil {
		return
	}
	plaintextBytes, err := smData.twoWayAES(ciphertext)
	if err != nil {
		return
	}
	t.Log("cipherText :", string(plaintextBytes))
	t.Fail()
}

func TestUnlock1(t *testing.T) {
	initViper(t)
	// created in 0.1.13
	wallet := viper.GetString("w1")
	testUnlock(t, wallet, viper.GetString("p1"))
}

func TestUnlock2(t *testing.T) {
	//
	keystore := "./my_wallet_0_2020-04-25T19-40-50.942Z.json"
	password := "<<password>>"
	testUnlock(t, keystore, password)
}

func TestUnlock3(t *testing.T) {
	// created in 0.1.14
	wallet := "/Users/daveappleton/Library/Application Support/Spacemesh/my_wallet_2020-06-18T20-33-24.592Z.json"
	testUnlock(t, wallet, "<<password>>")
}
func TestReMarshal(t *testing.T) {
	keystore := "./my_wallet_0_2020-04-25T19-40-50.942Z.json"
	smData, err := LoadWallet(keystore)
	if err != nil {
		t.Fatal(err)
	}
	err = smData.Unlock("<<password>>")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(smData.Crypto.confidential.Mnemonic)
	jsoned, err := json.Marshal(smData)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(jsoned))
	t.Fail()
}

func testBip39NewSeed(t *testing.T, phrase string) {
	seed := bip39.NewSeed(phrase, "")
	t.Log(len(seed))
	fmt.Printf("0x%x\n", seed)
	for i := uint64(0); i < 2; i++ {
		pk := ed25519.NewDerivedKeyFromSeed(seed[:32], i, []byte(spaceSalt))
		pub := pk.Public().(ed25519.PublicKey)[:]
		addr := types.BytesToAddress(pub)
		fmt.Println(addr.Hex())
	}
	t.Fail()
}

func TestPhrase1(t *testing.T) {
	phrase := "<<insert your own phrase here>>"
	testBip39NewSeed(t, phrase)
}

func TestPhrase2(t *testing.T) {
	phrase := "<<insert your own phrase here>>"
	testBip39NewSeed(t, phrase)
}

func salt(password []byte) []byte {
	return append([]byte("mnemonic"), password...)
}

func mnemonicToSeedSync(seed []byte, password []byte) []byte {
	return pbkdf2.Key(seed, salt(password), 2048, 64, sha512.New)
}

// insert your own mnemonic phrases here
func TestRollOurOwn(t *testing.T) {
	mnemonic1 := []byte("<phrase1>")
	mnemonic2 := []byte("<<phrase 2>>")
	mnemonic := mnemonic2
	if false {
		mnemonic = mnemonic1
	}
	var password []byte // []byte("<<password>>")
	seed := mnemonicToSeedSync(mnemonic, password)
	t.Log(seed)
	fmt.Printf("0x%x\n", seed)

	t.Log(len(seed))
	for i := uint64(0); i < 5; i++ {
		pk := ed25519.NewDerivedKeyFromSeed(seed[:32], i, []byte("Spacemesh blockmesh"))
		pub := pk.Public().(ed25519.PublicKey)[:]
		addr := types.BytesToAddress(pub)
		fmt.Println("addr ", i, addr.Hex())
		fmt.Printf("pub %d 0x%x\n", i, pub)
		fmt.Printf("private %d 0x%x\n", i, pk)
	}
	t.Fail()
}

// addr 0 0x0ac4c14349786c49ef0c426d4aa9e8463ef0b0a8

func TestNewMnemonic(t *testing.T) {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		t.Fatal(err)
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(mnemonic)
	t.Fail()
}

func TestWallets(t *testing.T) {
	w1 := "/Users/daveappleton/Library/Application Support/Spacemesh/my_wallet_0_2020-06-16T21-16-28.295Z.json"
	w2 := "/Users/daveappleton/Library/Application Support/Spacemesh/my_wallet_2020-06-18T20-33-24.592Z.json"
	testUnlock(t, w1, "<<password>>")
	testUnlock(t, w2, "<<password>>")

	t.Fail()
}

func (w *Wallet) relock() {
	w.unlocked = false
}

func TestReCrypt(t *testing.T) {
	keystore := "/Users/daveappleton/Library/Application Support/Spacemesh/my_wallet_0_2020-06-16T21-16-28.295Z.json"
	password := "<<password>>"
	smData, err := LoadWallet(keystore)
	chkTErr(t, err)
	err = smData.Unlock(password)
	chkTErr(t, err)

	t1 := smData.Crypto.CipherText
	err = smData.reCrypt()
	t2 := smData.Crypto.CipherText
	chkTErr(t, err)
	if len(t1) != len(t2) {
		t.Fatal("Lengths are different", len(t1), len(t2))
	}
	if t1 != t2 {
		b1 := util.Hex2Bytes(t1)
		b2 := util.Hex2Bytes(t2)
		for pos, b := range b1 {
			if b != b2[pos] {
				t.Log(pos, b, b2[pos])
			}
		}
		t.Log(t1)
		t.Log(t2)
	}
	smData.relock()
	err = smData.Unlock(password)
	chkTErr(t, err)

	t.Fail()
}

func TestReCrypt2(t *testing.T) {
	keystore := "/Users/daveappleton/Library/Application Support/Spacemesh/my_wallet_0_2020-06-16T21-16-28.295Z.json"
	password := "<<password>>"
	smData, err := LoadWallet(keystore)
	chkTErr(t, err)
	err = smData.Unlock(password)
	chkTErr(t, err)

	privatebuf1, err := json.Marshal(smData.Crypto.confidential)
	chkTErr(t, err)
	smData.relock()
	err = smData.Unlock(password)
	chkTErr(t, err)
	privatebuf2, err := json.Marshal(smData.Crypto.confidential)
	chkTErr(t, err)
	if len(privatebuf1) != len(privatebuf2) {
		t.Fatal("Lengths different", len(privatebuf1), len(privatebuf2))
	}
	for pos, b1 := range privatebuf1 {
		if b1 != privatebuf2[pos] {
			t.Log(pos, b1, privatebuf2[pos])
			t.Fail()
		}
	}
}

func TestNewWallet(t *testing.T) {
	smData, err := NewWallet("My Silly Wallet", "<<password>>")
	chkTErr(t, err)
	t.Log(smData.Meta.Created)
	smData.SaveWalletAs("tnw")
	testUnlock(t, smData.keystore, "<<password>>")
	err = os.Remove(smData.keystore)
	chkTErr(t, err)
	t.Fail()
}

func TestNewWalletPlusOne(t *testing.T) {
	smData, err := NewWallet("My Silly Wallet", "<<password>>")
	chkTErr(t, err)
	t.Log(smData.Meta.Created)
	n, err := smData.GenerateNewPair("Second Pair")
	chkTErr(t, err)
	err = smData.verifyAccounts()
	chkTErr(t, err)
	t.Log("added #", n)
	for accountNo, acc := range smData.Crypto.confidential.Accounts {
		pubKey, err := smData.GetPublicKey(accountNo)
		chkTErr(t, err)
		t.Log("pubkey", pubKey)
		address, err := smData.GetAddress(accountNo)
		chkTErr(t, err)
		t.Log("address", address.Hex())
		t.Log("pk", acc.SecretKey)
	}
	err = smData.verifyAccounts()
	chkTErr(t, err)
	smData.SaveWalletAs("tnw")
	testUnlock(t, smData.keystore, "<<password>>")
	err = os.Remove(smData.keystore)
	chkTErr(t, err)
	t.Fail()
}

func TestSignTransaction(t *testing.T) {
	keystore := "/Users/daveappleton/Library/Application Support/Spacemesh/my_wallet_0_2020-06-16T21-16-28.295Z.json"
	password := "<<password>>"
	smData, err := LoadWallet(keystore)
	chkTErr(t, err)
	err = smData.Unlock(password)
	chkTErr(t, err)
	var tx types.Transaction
	// random target address
	addr, err := util.Decode("0x865330189761187daa2243a1533b0412b8e14613")
	chkTErr(t, err)
	tx.Recipient = types.BytesToAddress(addr)
	tx.GasLimit = 100
	tx.Fee = 1
	tx.Amount = 1000000000000
	tx.AccountNonce = 3
	acc, err := smData.CurrentAccount()
	pk, _ := acc.PrivateKey()
	t.Logf("PK %x\n", pk)
	t.Log(smData.Crypto.confidential.Accounts[0].SecretKey)
	t.Log("sending from ", acc.Address().Hex())
	txs, err := smData.SignedTransaction(&tx)
	chkTErr(t, err)
	t.Logf("Signed Tx %x\n", txs)
	comma := ""
	fmt.Println("curl --header \"Content-Type: application/json\" \\")
	fmt.Println(" --request POST \\")
	fmt.Print(" --data '{\"tx\":[")
	for _, b := range txs {
		fmt.Print(comma)
		fmt.Print(b)
		comma = ","
	}
	fmt.Println("]}'\\")
	fmt.Println(" http://localhost:9090/v1/submittransaction")
	t.Fail()
}

func byteArrToBody(txs []byte) (body []byte) {
	bodyStr := fmt.Sprint("{\"tx\":[")
	comma := ""
	for _, b := range txs {
		bodyStr += fmt.Sprint(comma)
		bodyStr += fmt.Sprint(b)
		comma = ","
	}
	bodyStr += fmt.Sprintln("]}")
	return []byte(bodyStr)
}

var sendOver bool

// allows you to toggle the destination clients
func send(data []byte) {
	url := "http://localhost:9090/v1/submittransaction"
	// if sendOver {
	// 	url = "http://192.168.1.8:9090/v1/submittransaction"
	// }
	sendOver = !sendOver
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}

func TestBytes(t *testing.T) {
	str := "0xbeaad6d44a6c84f6ccf14ea271e5039a834054ee3efd569d5c1721288dca6082"
	by, _ := util.Decode(str)
	fmt.Print("[")
	comma := ""
	for _, b := range by {
		fmt.Print(comma, b)
		comma = ","
	}
	fmt.Println("]")
	t.Fail()
}

func TestSendTenTransactions(t *testing.T) {
	keystore := "/Users/daveappleton/Library/Application Support/Spacemesh/my_wallet_0_2020-06-16T21-16-28.295Z.json"
	password := "<<password>>"
	smData, err := LoadWallet(keystore)
	chkTErr(t, err)
	err = smData.Unlock(password)
	chkTErr(t, err)
	var tx types.Transaction
	tx.GasLimit = 100
	tx.Fee = 1
	tx.AccountNonce = 90
	//acc, err := smData.CurrentAccount()
	addrx := []string{"0xc57b32284d7d51d710ec52c033909d6ef4dd34bb"} //"0x865330189761187daa2243a1533b0412b8e14613"} //  "0x4a6d9d5cb2ed462c06dd88233686d7d69086bdc6", "0xfbccb6737f56416ddc982448c90e78e8d40b149b", "0x679c9f490e8f8c118f63229ab445cc3bb0b456c8"}
	for _, addry := range addrx {
		tx.Amount = 100000000000
		addr, err := util.Decode(addry)
		chkTErr(t, err)
		tx.Recipient = types.BytesToAddress(addr)
		for j := 0; j < 10; j++ {
			txs1, err := smData.SignedTransaction(&tx)
			chkTErr(t, err)
			tx.AccountNonce++
			tx.Amount += 100000000000
			body1 := byteArrToBody(txs1)
			send(body1)
			time.Sleep(50 * time.Millisecond)
		}
	}
	t.Fail()
}

/*
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"tx":[0,0,0,0,0,0,0,1,134,83,48,24,151,97,24,125,170,34,67,161,83,59,4,18,184,225,70,19,0,0,0,0,0,0,0,100,0,0,0,0,0,0,0,1,0,0,0,0,0,152,150,128,145,36,224,204,15,243,238,162,86,221,188,233,227,255,252,156,90,206,33,62,14,17,13,12,82,103,9,110,194,28,66,219,135,89,72,214,64,92,118,164,182,78,212,176,206,51,203,33,198,12,90,95,167,124,98,198,83,148,120,208,109,203,87,1]}' \
  http://localhost:9090/v1/submittransaction
*/
