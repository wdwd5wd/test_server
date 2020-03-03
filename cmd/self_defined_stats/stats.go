package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/QuarkChain/goquarkchain/common/hexutil"
	"github.com/ybbus/jsonrpc"
)

func main() {

	ip := flag.String("ip", "localhost", "Cluster IP")
	interval := flag.Uint("i", 10, "Query interval in second")
	// address := flag.String("a", "", "Query account balance if a QKC address is provided")
	// token := flag.String("t", "QKC", "Query account balance for a specific token")
	flag.Parse()
	privateEndPoint := jsonrpc.NewClient(fmt.Sprintf("http://%s:38491", *ip))
	publicEndPoint := jsonrpc.NewClient(fmt.Sprintf("http://%s:38391", *ip))
	QueryDetails(privateEndPoint, publicEndPoint, interval)
}

func QueryDetails(priclient jsonrpc.RPCClient, pubclient jsonrpc.RPCClient, interval *uint) {
	titles := []string{"Timestamp\t", "Syncing", "TPS", "Pend.TX", "Conf.TX", "BPS", "SBPS", "CPU", "ROOT", "CHAIN/SHARD-HEIGHT"}
	fmt.Println(strings.Join(titles, "\t"))
	intv := time.Duration(*interval)
	ticker := time.NewTicker(intv * time.Second)
	fmt.Println(Details(priclient, pubclient))
	for {
		select {
		case <-ticker.C:
			fmt.Println(Details(priclient, pubclient))
		}
	}
}

func Details(priclient jsonrpc.RPCClient, pubclient jsonrpc.RPCClient) interface{} {
	response, err := priclient.Call("getStats")
	if err != nil {
		return err.Error()
	}
	if response.Error != nil {
		return response.Error.Error()
	}
	res := response.Result.(map[string]interface{})
	shardsi := res["shards"].([]interface{})
	shards := make([]string, len(shardsi))
	var txs interface{}
	for i, p := range shardsi {
		shard := p.(map[string]interface{})
		shards[i] = fmt.Sprintf("%x/%x-%x", shard["chainId"], shard["shardId"], shard["height"])

		fullShardId, _ := shard["fullShardId"].(json.Number).Int64()
		height, _ := shard["height"].(json.Number).Int64()
		fmt.Println(height)
		resp, err := pubclient.Call("getMinorBlockByHeight", hexutil.EncodeUint64(uint64(fullShardId)), hexutil.EncodeUint64(uint64(height)), true)
		if err != nil {
			fmt.Println(err.Error())
		}
		// fmt.Println(resp)
		txs = resp.Result.(map[string]interface{})["transactions"]
		fmt.Println(fmt.Sprintf("%T", txs.([]interface{})))
	}
	// msg := strings.Join(shards, " ")

	return txs
}
