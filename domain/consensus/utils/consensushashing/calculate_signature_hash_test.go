package consensushashing_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/D-Stacks/go-secp256k1"

	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/util"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// shortened versions of SigHash types to fit in single line of test case
const (
	all                = consensushashing.SigHashAll
	none               = consensushashing.SigHashNone
	single             = consensushashing.SigHashSingle
	allAnyoneCanPay    = consensushashing.SigHashAll | consensushashing.SigHashAnyOneCanPay
	noneAnyoneCanPay   = consensushashing.SigHashNone | consensushashing.SigHashAnyOneCanPay
	singleAnyoneCanPay = consensushashing.SigHashSingle | consensushashing.SigHashAnyOneCanPay
)

func modifyOutput(outputIndex int) func(tx *externalapi.DomainTransaction) *externalapi.DomainTransaction {
	return func(tx *externalapi.DomainTransaction) *externalapi.DomainTransaction {
		clone := tx.Clone()
		clone.Outputs[outputIndex].Value = 100
		return clone
	}
}

func modifyInput(inputIndex int) func(tx *externalapi.DomainTransaction) *externalapi.DomainTransaction {
	return func(tx *externalapi.DomainTransaction) *externalapi.DomainTransaction {
		clone := tx.Clone()
		clone.Inputs[inputIndex].PreviousOutpoint.Index = 2
		return clone
	}
}

func modifyAmountSpent(inputIndex int) func(tx *externalapi.DomainTransaction) *externalapi.DomainTransaction {
	return func(tx *externalapi.DomainTransaction) *externalapi.DomainTransaction {
		clone := tx.Clone()
		utxoEntry := clone.Inputs[inputIndex].UTXOEntry
		clone.Inputs[inputIndex].UTXOEntry = utxo.NewUTXOEntry(666, utxoEntry.ScriptPublicKey(), false, 100)
		return clone
	}
}

func modifyScriptPublicKey(inputIndex int) func(tx *externalapi.DomainTransaction) *externalapi.DomainTransaction {
	return func(tx *externalapi.DomainTransaction) *externalapi.DomainTransaction {
		clone := tx.Clone()
		utxoEntry := clone.Inputs[inputIndex].UTXOEntry
		scriptPublicKey := utxoEntry.ScriptPublicKey()
		scriptPublicKey.Script = append(scriptPublicKey.Script, 1, 2, 3)
		clone.Inputs[inputIndex].UTXOEntry = utxo.NewUTXOEntry(utxoEntry.Amount(), scriptPublicKey, false, 100)
		return clone
	}
}

func modifySequence(inputIndex int) func(tx *externalapi.DomainTransaction) *externalapi.DomainTransaction {
	return func(tx *externalapi.DomainTransaction) *externalapi.DomainTransaction {
		clone := tx.Clone()
		clone.Inputs[inputIndex].Sequence = 12345
		return clone
	}
}

func modifyPayload(tx *externalapi.DomainTransaction) *externalapi.DomainTransaction {
	clone := tx.Clone()
	clone.Payload = []byte{6, 6, 6, 4, 2, 0, 1, 3, 3, 7}
	return clone
}

func modifyGas(tx *externalapi.DomainTransaction) *externalapi.DomainTransaction {
	clone := tx.Clone()
	clone.Gas = 1234
	return clone
}

func modifySubnetworkID(tx *externalapi.DomainTransaction) *externalapi.DomainTransaction {
	clone := tx.Clone()
	clone.SubnetworkID = externalapi.DomainSubnetworkID{6, 6, 6, 4, 2, 0, 1, 3, 3, 7}
	return clone
}

func TestCalculateSignatureHashSchnorr(t *testing.T) {
	nativeTx, subnetworkTx, err := generateTxs()
	if err != nil {
		t.Fatalf("Error from generateTxs: %+v", err)
	}

	// Note: Expected values were generated by the same code that they test,
	// As long as those were not verified using 3rd-party code they only check for regression, not correctness
	tests := []struct {
		name                  string
		tx                    *externalapi.DomainTransaction
		hashType              consensushashing.SigHashType
		inputIndex            int
		modificationFunction  func(*externalapi.DomainTransaction) *externalapi.DomainTransaction
		expectedSignatureHash string
	}{
		// native transactions

		// sigHashAll
		{name: "native-all-0", tx: nativeTx, hashType: all, inputIndex: 0,
			expectedSignatureHash: "b363613fe99c8bb1d3712656ec8dfaea621ee6a9a95d851aec5bb59363b03f5e"},
		{name: "native-all-0-modify-input-1", tx: nativeTx, hashType: all, inputIndex: 0,
			modificationFunction:  modifyInput(1), // should change the hash
			expectedSignatureHash: "34ae2989115068fc73a1b2cae023ad79c3cdb5cbe532a46fa91d9181a36990fd"},
		{name: "native-all-0-modify-output-1", tx: nativeTx, hashType: all, inputIndex: 0,
			modificationFunction:  modifyOutput(1), // should change the hash
			expectedSignatureHash: "043441346c66e461f9f1dc618ebbfe7fd87f74e363f267bf8b3243a7bfe0c870"},
		{name: "native-all-0-modify-sequence-1", tx: nativeTx, hashType: all, inputIndex: 0,
			modificationFunction:  modifySequence(1), // should change the hash
			expectedSignatureHash: "de8d3d46bc8c51f51a1b85470f8bf01ee38214901d6d514fd13bafe4efc8aa0f"},
		{name: "native-all-anyonecanpay-0", tx: nativeTx, hashType: allAnyoneCanPay, inputIndex: 0,
			expectedSignatureHash: "19897764789644c2ac5cd6d83f7a78a1208f3ce6d15e8788f9b9fa6d7c91d8f1"},
		{name: "native-all-anyonecanpay-0-modify-input-0", tx: nativeTx, hashType: allAnyoneCanPay, inputIndex: 0,
			modificationFunction:  modifyInput(0), // should change the hash
			expectedSignatureHash: "f1ff39b1b9ce86d2fdfac61a75f3b13e98fe5e0f1057b4ec69245031ecf7be37"},
		{name: "native-all-anyonecanpay-0-modify-input-1", tx: nativeTx, hashType: allAnyoneCanPay, inputIndex: 0,
			modificationFunction:  modifyInput(1), // shouldn't change the hash
			expectedSignatureHash: "19897764789644c2ac5cd6d83f7a78a1208f3ce6d15e8788f9b9fa6d7c91d8f1"},
		{name: "native-all-anyonecanpay-0-modify-sequence", tx: nativeTx, hashType: allAnyoneCanPay, inputIndex: 0,
			modificationFunction:  modifySequence(1), // shouldn't change the hash
			expectedSignatureHash: "19897764789644c2ac5cd6d83f7a78a1208f3ce6d15e8788f9b9fa6d7c91d8f1"},

		// sigHashNone
		{name: "native-none-0", tx: nativeTx, hashType: none, inputIndex: 0,
			expectedSignatureHash: "7a5b0fef8219bb72ef1912db5335c71c4fdfac873a6096c24b2f0b5c3774349c"},
		{name: "native-none-0-modify-output-1", tx: nativeTx, hashType: none, inputIndex: 0,
			modificationFunction:  modifyOutput(1), // shouldn't change the hash
			expectedSignatureHash: "7a5b0fef8219bb72ef1912db5335c71c4fdfac873a6096c24b2f0b5c3774349c"},
		{name: "native-none-0-modify-sequence-0", tx: nativeTx, hashType: none, inputIndex: 0,
			modificationFunction:  modifySequence(0), // should change the hash
			expectedSignatureHash: "852011233473ee1e61a9d0e51fb5ecd65857ceca65ebea4c54b6d557f2006f2a"},
		{name: "native-none-0-modify-sequence-1", tx: nativeTx, hashType: none, inputIndex: 0,
			modificationFunction:  modifySequence(1), // shouldn't change the hash
			expectedSignatureHash: "7a5b0fef8219bb72ef1912db5335c71c4fdfac873a6096c24b2f0b5c3774349c"},
		{name: "native-none-anyonecanpay-0", tx: nativeTx, hashType: noneAnyoneCanPay, inputIndex: 0,
			expectedSignatureHash: "1624d46e77d09cb09e4a7dcbf419b8c37671bd0274b9dc6aba0668922da83935"},
		{name: "native-none-anyonecanpay-0-modify-amount-spent", tx: nativeTx, hashType: noneAnyoneCanPay, inputIndex: 0,
			modificationFunction:  modifyAmountSpent(0), // should change the hash
			expectedSignatureHash: "235f0766528865a4c478a46b0b3eef6b4760c6a05c792a452c06fab9ad0bd57c"},
		{name: "native-none-anyonecanpay-0-modify-script-public-key", tx: nativeTx, hashType: noneAnyoneCanPay, inputIndex: 0,
			modificationFunction:  modifyScriptPublicKey(0), // should change the hash
			expectedSignatureHash: "42b408acc6df78f1b1aef605339233af129b6e656788e8c93712e4954d28583d"},

		// sigHashSingle
		{name: "native-single-0", tx: nativeTx, hashType: single, inputIndex: 0,
			expectedSignatureHash: "c9f7adaa7a22af87195183cf1f10e368429139f16069597d5631a0f522e320a5"},
		{name: "native-single-0-modify-output-0", tx: nativeTx, hashType: single, inputIndex: 0,
			modificationFunction:  modifyOutput(0), // should change the hash
			expectedSignatureHash: "af40fbd0ac061c586484c4f266d44007c0715eb0b80d20eb89be65325db05716"},
		{name: "native-single-0-modify-output-1", tx: nativeTx, hashType: single, inputIndex: 0,
			modificationFunction:  modifyOutput(1), // shouldn't change the hash
			expectedSignatureHash: "c9f7adaa7a22af87195183cf1f10e368429139f16069597d5631a0f522e320a5"},
		{name: "native-single-0-modify-sequence-0", tx: nativeTx, hashType: single, inputIndex: 0,
			modificationFunction:  modifySequence(0), // should change the hash
			expectedSignatureHash: "c40f48b35fc933d5930c612c420e80bad336388126aaba6073588e31d95aca2c"},
		{name: "native-single-0-modify-sequence-1", tx: nativeTx, hashType: single, inputIndex: 0,
			modificationFunction:  modifySequence(1), // shouldn't change the hash
			expectedSignatureHash: "c9f7adaa7a22af87195183cf1f10e368429139f16069597d5631a0f522e320a5"},
		{name: "native-single-2-no-corresponding-output", tx: nativeTx, hashType: single, inputIndex: 2,
			expectedSignatureHash: "145487f676cd1d5f8042b9d042cc63bc0ecdf20563d324fa0b847714eeb94816"},
		{name: "native-single-2-no-corresponding-output-modify-output-1", tx: nativeTx, hashType: single, inputIndex: 2,
			modificationFunction:  modifyOutput(1), // shouldn't change the hash
			expectedSignatureHash: "145487f676cd1d5f8042b9d042cc63bc0ecdf20563d324fa0b847714eeb94816"},
		{name: "native-single-anyonecanpay-0", tx: nativeTx, hashType: singleAnyoneCanPay, inputIndex: 0,
			expectedSignatureHash: "4f3f758e1ed9c438dcc241efd31dd07e6bf2e11e900e105eebd4d337391e48fe"},
		{name: "native-single-anyonecanpay-2-no-corresponding-output", tx: nativeTx, hashType: singleAnyoneCanPay, inputIndex: 2,
			expectedSignatureHash: "200207998528ab3b58cbdfe578cd079572eb3093e68fb5c728e505b847e91c64"},

		// subnetwork transaction
		{name: "subnetwork-all-0", tx: subnetworkTx, hashType: all, inputIndex: 0,
			expectedSignatureHash: "b2f421c933eb7e1a91f1d9e1efa3f120fe419326c0dbac487752189522550e0c"},
		{name: "subnetwork-all-modify-payload", tx: subnetworkTx, hashType: all, inputIndex: 0,
			modificationFunction:  modifyPayload, // should change the hash
			expectedSignatureHash: "12ab63b9aea3d58db339245a9b6e9cb6075b2253615ce0fb18104d28de4435a1"},
		{name: "subnetwork-all-modify-gas", tx: subnetworkTx, hashType: all, inputIndex: 0,
			modificationFunction:  modifyGas, // should change the hash
			expectedSignatureHash: "2501edfc0068d591160c4bd98646c6e6892cdc051182a8be3ccd6d67f104fd17"},
		{name: "subnetwork-all-subnetwork-id", tx: subnetworkTx, hashType: all, inputIndex: 0,
			modificationFunction:  modifySubnetworkID, // should change the hash
			expectedSignatureHash: "a5d1230ede0dfcfd522e04123a7bcd721462fed1d3a87352031a4f6e3c4389b6"},
	}

	for _, test := range tests {
		tx := test.tx
		if test.modificationFunction != nil {
			tx = test.modificationFunction(tx)
		}

		actualSignatureHash, err := consensushashing.CalculateSignatureHashSchnorr(
			tx, test.inputIndex, test.hashType, &consensushashing.SighashReusedValues{})
		if err != nil {
			t.Errorf("%s: Error from CalculateSignatureHashSchnorr: %+v", test.name, err)
			continue
		}

		if actualSignatureHash.String() != test.expectedSignatureHash {
			t.Errorf("%s: expected signature hash: '%s'; but got: '%s'",
				test.name, test.expectedSignatureHash, actualSignatureHash)
		}
	}
}

func TestCalculateSignatureHashECDSA(t *testing.T) {
	nativeTx, subnetworkTx, err := generateTxs()
	if err != nil {
		t.Fatalf("Error from generateTxs: %+v", err)
	}

	// Note: Expected values were generated by the same code that they test,
	// As long as those were not verified using 3rd-party code they only check for regression, not correctness
	tests := []struct {
		name                  string
		tx                    *externalapi.DomainTransaction
		hashType              consensushashing.SigHashType
		inputIndex            int
		modificationFunction  func(*externalapi.DomainTransaction) *externalapi.DomainTransaction
		expectedSignatureHash string
	}{
		// native transactions

		// sigHashAll
		{name: "native-all-0", tx: nativeTx, hashType: all, inputIndex: 0,
			expectedSignatureHash: "6ec7f4949d0c095d78bf41475310fd38eb054f3e7c4240daf91ea888e4eb9a30"},
		{name: "native-all-0-modify-input-1", tx: nativeTx, hashType: all, inputIndex: 0,
			modificationFunction:  modifyInput(1), // should change the hash
			expectedSignatureHash: "34fcc1cb538736c473c1778eba4df5f88c3d9f27508b0d842ec2348d097cd103"},
		{name: "native-all-0-modify-output-1", tx: nativeTx, hashType: all, inputIndex: 0,
			modificationFunction:  modifyOutput(1), // should change the hash
			expectedSignatureHash: "faf02d20d32f0e4536dfb0a86c67f97b394c11a34069bd74a2f7533ea964b10f"},
		{name: "native-all-0-modify-sequence-1", tx: nativeTx, hashType: all, inputIndex: 0,
			modificationFunction:  modifySequence(1), // should change the hash
			expectedSignatureHash: "25484c5dcc89d21e5b5858847964c8c2938d5090be54b21a590099ce4f792b14"},
		{name: "native-all-anyonecanpay-0", tx: nativeTx, hashType: allAnyoneCanPay, inputIndex: 0,
			expectedSignatureHash: "458a711830a66d592c89845cd6406b525b5f89f4d9ca50abbdbb48dbb5adbb07"},
		{name: "native-all-anyonecanpay-0-modify-input-0", tx: nativeTx, hashType: allAnyoneCanPay, inputIndex: 0,
			modificationFunction:  modifyInput(0), // should change the hash
			expectedSignatureHash: "67157f1984a881c71ea92c9959da1b856383489a8bb0150783cdc4d58bca95ea"},
		{name: "native-all-anyonecanpay-0-modify-input-1", tx: nativeTx, hashType: allAnyoneCanPay, inputIndex: 0,
			modificationFunction:  modifyInput(1), // shouldn't change the hash
			expectedSignatureHash: "458a711830a66d592c89845cd6406b525b5f89f4d9ca50abbdbb48dbb5adbb07"},
		{name: "native-all-anyonecanpay-0-modify-sequence", tx: nativeTx, hashType: allAnyoneCanPay, inputIndex: 0,
			modificationFunction:  modifySequence(1), // shouldn't change the hash
			expectedSignatureHash: "458a711830a66d592c89845cd6406b525b5f89f4d9ca50abbdbb48dbb5adbb07"},

		// sigHashNone
		{name: "native-none-0", tx: nativeTx, hashType: none, inputIndex: 0,
			expectedSignatureHash: "bf92d39b8381e49d4b2f37a7d2e2d9b4f126b6659cb873b84ae3db8910cd9664"},
		{name: "native-none-0-modify-output-1", tx: nativeTx, hashType: none, inputIndex: 0,
			modificationFunction:  modifyOutput(1), // shouldn't change the hash
			expectedSignatureHash: "bf92d39b8381e49d4b2f37a7d2e2d9b4f126b6659cb873b84ae3db8910cd9664"},
		{name: "native-none-0-modify-sequence-0", tx: nativeTx, hashType: none, inputIndex: 0,
			modificationFunction:  modifySequence(0), // should change the hash
			expectedSignatureHash: "20550f85a6ac0d4b20ebb0d8df9b1f4ec0ecb3df5adf539c9d6ad9af03f712d6"},
		{name: "native-none-0-modify-sequence-1", tx: nativeTx, hashType: none, inputIndex: 0,
			modificationFunction:  modifySequence(1), // shouldn't change the hash
			expectedSignatureHash: "bf92d39b8381e49d4b2f37a7d2e2d9b4f126b6659cb873b84ae3db8910cd9664"},
		{name: "native-none-anyonecanpay-0", tx: nativeTx, hashType: noneAnyoneCanPay, inputIndex: 0,
			expectedSignatureHash: "a048ec7a396397e1357b42905f26c51d0ec6c0943298ff4f2b8707ec3e8e1aa0"},
		{name: "native-none-anyonecanpay-0-modify-amount-spent", tx: nativeTx, hashType: noneAnyoneCanPay, inputIndex: 0,
			modificationFunction:  modifyAmountSpent(0), // should change the hash
			expectedSignatureHash: "66125d23d3dc9711683a6dbc96d4d4411af41e71f92596e9983ea8c5e3a04753"},
		{name: "native-none-anyonecanpay-0-modify-script-public-key", tx: nativeTx, hashType: noneAnyoneCanPay, inputIndex: 0,
			modificationFunction:  modifyScriptPublicKey(0), // should change the hash
			expectedSignatureHash: "0ba5f527f8408b252eb77ea54efe63b831c736fea4bed58fc47c4ceaabf3f6cf"},

		// sigHashSingle
		{name: "native-single-0", tx: nativeTx, hashType: single, inputIndex: 0,
			expectedSignatureHash: "b21ec5c5e1830f8b9b3cb13bfbd542318a17d89d9844bd64167696ca36374f7f"},
		{name: "native-single-0-modify-output-0", tx: nativeTx, hashType: single, inputIndex: 0,
			modificationFunction:  modifyOutput(0), // should change the hash
			expectedSignatureHash: "e15914f6b22979f70162f5c57b3ad7ceff91b8a2356960f66a23dc8e602303fe"},
		{name: "native-single-0-modify-output-1", tx: nativeTx, hashType: single, inputIndex: 0,
			modificationFunction:  modifyOutput(1), // shouldn't change the hash
			expectedSignatureHash: "b21ec5c5e1830f8b9b3cb13bfbd542318a17d89d9844bd64167696ca36374f7f"},
		{name: "native-single-0-modify-sequence-0", tx: nativeTx, hashType: single, inputIndex: 0,
			modificationFunction:  modifySequence(0), // should change the hash
			expectedSignatureHash: "a09f20428456475bc5fcff07242416d439faa0dec37152e31a8546874f323473"},
		{name: "native-single-0-modify-sequence-1", tx: nativeTx, hashType: single, inputIndex: 0,
			modificationFunction:  modifySequence(1), // shouldn't change the hash
			expectedSignatureHash: "b21ec5c5e1830f8b9b3cb13bfbd542318a17d89d9844bd64167696ca36374f7f"},
		{name: "native-single-2-no-corresponding-output", tx: nativeTx, hashType: single, inputIndex: 2,
			expectedSignatureHash: "7cc3c80a6250599e47e4ceca66e3670b4fc74a009aba2b7df737bc37e8cb5b79"},
		{name: "native-single-2-no-corresponding-output-modify-output-1", tx: nativeTx, hashType: single, inputIndex: 2,
			modificationFunction:  modifyOutput(1), // shouldn't change the hash
			expectedSignatureHash: "7cc3c80a6250599e47e4ceca66e3670b4fc74a009aba2b7df737bc37e8cb5b79"},
		{name: "native-single-anyonecanpay-0", tx: nativeTx, hashType: singleAnyoneCanPay, inputIndex: 0,
			expectedSignatureHash: "8040f5ebfc6c5a8285272d5e1956dd3036eaa9a7abec9b18cb1b614a015f2fc7"},
		{name: "native-single-anyonecanpay-2-no-corresponding-output", tx: nativeTx, hashType: singleAnyoneCanPay, inputIndex: 2,
			expectedSignatureHash: "5e1ac311544301aa6afa578f18e1d1871ffbc15915e01f25f2375715c3a3147d"},

		// subnetwork transaction
		{name: "subnetwork-all-0", tx: subnetworkTx, hashType: all, inputIndex: 0,
			expectedSignatureHash: "807d351414ff592ba097daa5c7937311d6382107f23a6ae415954e248a0527e0"},
		{name: "subnetwork-all-modify-payload", tx: subnetworkTx, hashType: all, inputIndex: 0,
			modificationFunction:  modifyPayload, // should change the hash
			expectedSignatureHash: "0bb2a9a37cc27a60c91c1c9b5ff29bc09f1b39faa3ec55edb15dcbc6c9ce03d7"},
		{name: "subnetwork-all-modify-gas", tx: subnetworkTx, hashType: all, inputIndex: 0,
			modificationFunction:  modifyGas, // should change the hash
			expectedSignatureHash: "78dcfa1ea6a6f01c31805bda3cc71d7356f32b87a8bf3b80b4a4d0d5f95e8741"},
		{name: "subnetwork-all-subnetwork-id", tx: subnetworkTx, hashType: all, inputIndex: 0,
			modificationFunction:  modifySubnetworkID, // should change the hash
			expectedSignatureHash: "6412917f0d5d856c37897d9a98c3817dc1f1668deff73efeefbe2529e00e3511"},
	}

	for _, test := range tests {
		tx := test.tx
		if test.modificationFunction != nil {
			tx = test.modificationFunction(tx)
		}

		actualSignatureHash, err := consensushashing.CalculateSignatureHashECDSA(
			tx, test.inputIndex, test.hashType, &consensushashing.SighashReusedValues{})
		if err != nil {
			t.Errorf("%s: Error from CalculateSignatureHashECDSA: %+v", test.name, err)
			continue
		}

		if actualSignatureHash.String() != test.expectedSignatureHash {
			t.Errorf("%s: expected signature hash: '%s'; but got: '%s'",
				test.name, test.expectedSignatureHash, actualSignatureHash)
		}
	}
}

func generateTxs() (nativeTx, subnetworkTx *externalapi.DomainTransaction, err error) {
	genesisCoinbase := dagconfig.SimnetParams.GenesisBlock.Transactions[0]
	genesisCoinbaseTransactionID := consensushashing.TransactionID(genesisCoinbase)

	address1Str := "kaspasim:qzpj2cfa9m40w9m2cmr8pvfuqpp32mzzwsuw6ukhfduqpp32mzzws59e8fapc"
	address1, err := util.DecodeAddress(address1Str, util.Bech32PrefixKaspaSim)
	if err != nil {
		return nil, nil, fmt.Errorf("error decoding address1: %+v", err)
	}
	address1ToScript, err := txscript.PayToAddrScript(address1)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating script: %+v", err)
	}

	address2Str := "kaspasim:qr7w7nqsdnc3zddm6u8s9fex4ysk95hm3v30q353ymuqpp32mzzws59e8fapc"
	address2, err := util.DecodeAddress(address2Str, util.Bech32PrefixKaspaSim)
	if err != nil {
		return nil, nil, fmt.Errorf("error decoding address2: %+v", err)
	}
	address2ToScript, err := txscript.PayToAddrScript(address2)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating script: %+v", err)
	}

	txIns := []*externalapi.DomainTransactionInput{
		{
			PreviousOutpoint: *externalapi.NewDomainOutpoint(genesisCoinbaseTransactionID, 0),
			Sequence:         0,
			UTXOEntry:        utxo.NewUTXOEntry(100, address1ToScript, false, 0),
		},
		{
			PreviousOutpoint: *externalapi.NewDomainOutpoint(genesisCoinbaseTransactionID, 1),
			Sequence:         1,
			UTXOEntry:        utxo.NewUTXOEntry(200, address2ToScript, false, 0),
		},
		{
			PreviousOutpoint: *externalapi.NewDomainOutpoint(genesisCoinbaseTransactionID, 2),
			Sequence:         2,
			UTXOEntry:        utxo.NewUTXOEntry(300, address2ToScript, false, 0),
		},
	}

	txOuts := []*externalapi.DomainTransactionOutput{
		{
			Value:           300,
			ScriptPublicKey: address2ToScript,
		},
		{
			Value:           300,
			ScriptPublicKey: address1ToScript,
		},
	}

	nativeTx = &externalapi.DomainTransaction{
		Version:      0,
		Inputs:       txIns,
		Outputs:      txOuts,
		LockTime:     1615462089000,
		SubnetworkID: externalapi.DomainSubnetworkID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}
	subnetworkTx = &externalapi.DomainTransaction{
		Version:      0,
		Inputs:       txIns,
		Outputs:      txOuts,
		LockTime:     1615462089000,
		SubnetworkID: externalapi.DomainSubnetworkID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		Gas:          250,
		Payload:      []byte{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
	}

	return nativeTx, subnetworkTx, nil
}

func BenchmarkCalculateSignatureHashSchnorr(b *testing.B) {
	sigHashTypes := []consensushashing.SigHashType{
		consensushashing.SigHashAll,
		consensushashing.SigHashNone,
		consensushashing.SigHashSingle,
		consensushashing.SigHashAll | consensushashing.SigHashAnyOneCanPay,
		consensushashing.SigHashNone | consensushashing.SigHashAnyOneCanPay,
		consensushashing.SigHashSingle | consensushashing.SigHashAnyOneCanPay}

	for _, size := range []int{10, 100, 1000} {
		tx := generateTransaction(b, sigHashTypes, size)

		b.Run(fmt.Sprintf("%d-inputs-and-outputs", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				reusedValues := &consensushashing.SighashReusedValues{}
				for inputIndex := range tx.Inputs {
					sigHashType := sigHashTypes[inputIndex%len(sigHashTypes)]
					_, err := consensushashing.CalculateSignatureHashSchnorr(tx, inputIndex, sigHashType, reusedValues)
					if err != nil {
						b.Fatalf("Error from CalculateSignatureHashSchnorr: %+v", err)
					}
				}
			}
		})
	}
}

func generateTransaction(b *testing.B, sigHashTypes []consensushashing.SigHashType, inputAndOutputSizes int) *externalapi.DomainTransaction {
	sourceScript := getSourceScript(b)
	tx := &externalapi.DomainTransaction{
		Version:      0,
		Inputs:       generateInputs(inputAndOutputSizes, sourceScript),
		Outputs:      generateOutputs(inputAndOutputSizes, sourceScript),
		LockTime:     123456789,
		SubnetworkID: externalapi.DomainSubnetworkID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		Gas:          125,
		Payload:      []byte{9, 8, 7, 6, 5, 4, 3, 2, 1},
		Fee:          0,
		Mass:         0,
		ID:           nil,
	}
	signTx(b, tx, sigHashTypes)
	return tx
}

func signTx(b *testing.B, tx *externalapi.DomainTransaction, sigHashTypes []consensushashing.SigHashType) {
	sourceAddressPKStr := "a4d85b7532123e3dd34e58d7ce20895f7ca32349e29b01700bb5a3e72d2570eb"
	privateKeyBytes, err := hex.DecodeString(sourceAddressPKStr)
	if err != nil {
		b.Fatalf("Error parsing private key hex: %+v", err)
	}
	keyPair, err := secp256k1.DeserializeSchnorrPrivateKeyFromSlice(privateKeyBytes)
	if err != nil {
		b.Fatalf("Error deserializing private key: %+v", err)
	}
	for i, txIn := range tx.Inputs {
		signatureScript, err := txscript.SignatureScript(
			tx, i, sigHashTypes[i%len(sigHashTypes)], keyPair, &consensushashing.SighashReusedValues{})
		if err != nil {
			b.Fatalf("Error from SignatureScript: %+v", err)
		}
		txIn.SignatureScript = signatureScript
	}

}

func generateInputs(size int, sourceScript *externalapi.ScriptPublicKey) []*externalapi.DomainTransactionInput {
	inputs := make([]*externalapi.DomainTransactionInput, size)

	for i := 0; i < size; i++ {
		inputs[i] = &externalapi.DomainTransactionInput{
			PreviousOutpoint: *externalapi.NewDomainOutpoint(
				externalapi.NewDomainTransactionIDFromByteArray(&[32]byte{12, 3, 4, 5}), 1),
			SignatureScript: nil,
			Sequence:        uint64(i),
			UTXOEntry:       utxo.NewUTXOEntry(uint64(i), sourceScript, false, 12),
		}
	}

	return inputs
}

func getSourceScript(b *testing.B) *externalapi.ScriptPublicKey {
	sourceAddressStr := "kaspasim:qz6f9z6l3x4v3lf9mgf0t934th4nx5kgzu663x9yjh"

	sourceAddress, err := util.DecodeAddress(sourceAddressStr, util.Bech32PrefixKaspaSim)
	if err != nil {
		b.Fatalf("Error from DecodeAddress: %+v", err)
	}

	sourceScript, err := txscript.PayToAddrScript(sourceAddress)
	if err != nil {
		b.Fatalf("Error from PayToAddrScript: %+v", err)
	}
	return sourceScript
}

func generateOutputs(size int, script *externalapi.ScriptPublicKey) []*externalapi.DomainTransactionOutput {
	outputs := make([]*externalapi.DomainTransactionOutput, size)

	for i := 0; i < size; i++ {
		outputs[i] = &externalapi.DomainTransactionOutput{
			Value:           uint64(i),
			ScriptPublicKey: script,
		}
	}

	return outputs
}
