package daos

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"gopkg.in/mgo.v2/bson"
	"sync"
	"ubiq-explorer/models"
	"ubiq-explorer/models/db"
	"ubiq-explorer/node"
)

type TransactionDAO struct{}

func NewTransactionDAO() *TransactionDAO {
	return &TransactionDAO{}
}

func (dao *TransactionDAO) Insert(txn models.Transaction, wg *sync.WaitGroup) (bool, error) {
	if wg != nil {
		wg.Add(1)
		defer wg.Done()
	}
	conn := db.Conn()
	defer conn.Close()
	err := conn.DB("").C("transactions").Insert(txn)

	if err != nil {
		return false, err
	}
	return true, nil
}

func (dao *TransactionDAO) Find(query bson.M, sort string, limit int, cursor string) (models.TransactionPage, error) {
	conn := db.Conn()
	defer conn.Close()

	count, err := conn.DB("").C("transactions").Find(query).Count()

	if cursor != "" {
		query["_id"] = bson.M{"$lt": bson.ObjectIdHex(cursor)}
	}

	var txns []*models.Transaction
	var page = models.TransactionPage{Total: count, Start: "", End: ""}
	err = conn.DB("").C("transactions").Find(query).Sort(sort).Limit(limit).All(&txns)
	if err != nil {
		return page, err
	}
	if len(txns) > 0 {
		page.Start = txns[0].Id.Hex()
		page.End = txns[len(txns)-1].Id.Hex()
	}
	page.Transactions = txns
	return page, nil
}

// I don't like doing this using RPC instead of the Ethereum RPC client, but it's way too tedious to get all of this data using the client because nothing is exported
func (dao *TransactionDAO) GetFromRPC(hash common.Hash) (*models.RpcTransaction, error) {
	node := node.RPC()
	var json *models.RpcTransaction
	err := node.CallContext(context.TODO(), &json, "eth_getTransactionByHash", hash)
	if err != nil {
		return nil, err
	} else if json == nil {
		return nil, fmt.Errorf("Not found")
	}
	return json, nil
}

func (dao *TransactionDAO) Receipt(hash common.Hash) (*models.Receipt, error) {
	node := node.RPC()
	var json *models.Receipt
	err := node.CallContext(context.TODO(), &json, "eth_getTransactionReceipt", hash)
	if err != nil {
		return nil, err
	} else if json == nil {
		return nil, fmt.Errorf("Not found")
	}
	return json, nil
}

func (dao *TransactionDAO) Debug(hash common.Hash) (*models.RpcTraceResult, error) {
	node := node.RPC()
	var json *models.RpcTraceResult
	err := node.CallContext(context.TODO(), &json, "debug_traceTransaction", hash)
	if err != nil {
		return nil, err
	} else if json == nil {
		return nil, fmt.Errorf("Not found")
	}
	return json, nil
}
