package dummy

import (
	//"strings"

	"github.com/tendermint/abci/types"
	"github.com/tendermint/merkleeyes/iavl"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/merkle"
	"github.com/tendermint/go-wire"

)

type DummyApplication struct {
	types.BaseApplication

	state merkle.Tree
}
// Transaction type bytes
const (
	WriteSet byte = 0x01
	WriteRem byte = 0x02
)
func NewDummyApplication() *DummyApplication {
	state := iavl.NewIAVLTree(0, nil)
	return &DummyApplication{state: state}
}

func (app *DummyApplication) Info() (resInfo types.ResponseInfo) {
	return types.ResponseInfo{Data: cmn.Fmt("{\"size\":%v}", app.state.Size())}
}

// tx is either "key=value" or just arbitrary bytes
func (app *DummyApplication) DeliverTx(tx []byte) types.Result {
	//parts := strings.Split(string(tx), "=")
	//if len(parts) == 2 {
	//	app.state.Set([]byte(parts[0]), []byte(parts[1]))
	//} else {
	//	app.state.Set(tx, tx)
	//}
	//return types.OK
	tree := app.state
	return app.doTx(tree, tx)
}
func (app *DummyApplication) doTx(tree merkle.Tree, tx []byte) types.Result {
	if len(tx) == 0 {
		return types.ErrEncodingError.SetLog("Tx length cannot be zero")
	}
	typeByte := tx[0]
	tx = tx[1:]
	switch typeByte {
	case WriteSet: // Set
		key, n, err := wire.GetByteSlice(tx)
		if err != nil {
			return types.ErrEncodingError.SetLog(cmn.Fmt("Error reading key: %v", err.Error()))
		}
		tx = tx[n:]
		value, n, err := wire.GetByteSlice(tx)
		if err != nil {
			return types.ErrEncodingError.SetLog(cmn.Fmt("Error reading value: %v", err.Error()))
		}
		tx = tx[n:]
		if len(tx) != 0 {
			return types.ErrEncodingError.SetLog(cmn.Fmt("Got bytes left over"))
		}

		tree.Set(key, value)
	case WriteRem: // Remove
		key, n, err := wire.GetByteSlice(tx)
		if err != nil {
			return types.ErrEncodingError.SetLog(cmn.Fmt("Error reading key: %v", err.Error()))
		}
		tx = tx[n:]
		if len(tx) != 0 {
			return types.ErrEncodingError.SetLog(cmn.Fmt("Got bytes left over"))
		}
		tree.Remove(key)
	default:
		return types.ErrUnknownRequest.SetLog(cmn.Fmt("Unexpected Tx type byte %X", typeByte))
	}
	return types.OK
}

func (app *DummyApplication) CheckTx(tx []byte) types.Result {
	//return types.OK
	tree := app.state
	return app.doTx(tree, tx)
}

func (app *DummyApplication) Commit() types.Result {
	hash := app.state.Hash()
	return types.NewResultOK(hash, "")
}

func (app *DummyApplication) Query(reqQuery types.RequestQuery) (resQuery types.ResponseQuery) {
	if reqQuery.Prove {
		value, proof, exists := app.state.Proof(reqQuery.Data)
		resQuery.Index = -1 // TODO make Proof return index
		resQuery.Key = reqQuery.Data
		resQuery.Value = value
		resQuery.Proof = proof
		if exists {
			resQuery.Log = "exists"
		} else {
			resQuery.Log = "does not exist"
		}
		return
	} else {
		index, value, exists := app.state.Get(reqQuery.Data)
		resQuery.Index = int64(index)
		resQuery.Value = value
		if exists {
			resQuery.Log = "exists"
		} else {
			resQuery.Log = "does not exist"
		}
		return
	}
}
