package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"math/big"
	"net"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/QuarkChain/goquarkchain/account"
	"github.com/QuarkChain/goquarkchain/common/hexutil"
	"github.com/QuarkChain/goquarkchain/internal/qkcapi"
	"github.com/shirou/gopsutil/mem"
	"github.com/ybbus/jsonrpc"
)

func basic(clt jsonrpc.RPCClient, ip string) string {

	response, err := clt.Call("getStats")
	if err != nil {
		return err.Error()
	}
	if response.Error != nil {
		return response.Error.Error()
	}
	res := response.Result.(map[string]interface{})
	//fmt.Println("response", res)

	msg := "============================\n"
	msg += "QuarkChain Cluster Stats\n"
	msg += "============================\n"
	msg += fmt.Sprintf("CPU:                %d\n", runtime.NumCPU())
	m := runtime.MemStats{}
	runtime.ReadMemStats(&m)
	v, _ := mem.VirtualMemory()
	msg += fmt.Sprintf("Memory:             %d GB\n", v.Total/1024/1024/1024)
	msg += fmt.Sprintf("IP:                 %s\n", ip)
	chains, _ := res["chainSize"].(json.Number).Int64()
	msg += fmt.Sprintf("Chains:             %d\n", chains)
	networkId, _ := res["networkId"].(json.Number).Int64()
	msg += fmt.Sprintf("Network Id:         %d\n", networkId)
	peersI := res["peers"].([]interface{})
	peers := make([]string, len(peersI))
	for i, p := range peersI {
		peers[i] = p.(string)
	}
	msg += fmt.Sprintf("Peers:              %s\n", strings.Join(peers, ","))
	msg += "============================"
	return msg
}

func queryStats(client jsonrpc.RPCClient, interval *uint) {
	titles := []string{"Timestamp\t", "Syncing", "TPS", "Pend.TX", "Conf.TX", "BPS", "SBPS", "CPU", "ROOT", "CHAIN/SHARD-HEIGHT"}
	fmt.Println(strings.Join(titles, "\t"))
	intv := time.Duration(*interval)
	ticker := time.NewTicker(intv * time.Second)
	fmt.Println(stats(client))
	for {
		select {
		case <-ticker.C:
			fmt.Println(stats(client))
		}
	}
}

func stats(client jsonrpc.RPCClient) string {
	response, err := client.Call("getStats")
	if err != nil {
		return err.Error()
	}
	if response.Error != nil {
		return response.Error.Error()
	}
	res := response.Result.(map[string]interface{})
	t := time.Now()
	msg := t.Format("2006-01-02 15:04:05")
	msg += "\t"
	msg += fmt.Sprintf("%t", res["syncing"])
	msg += "\t"
	txCount, _ := res["txCount60s"].(json.Number).Int64()
	msg += fmt.Sprintf("%2.2f", float64(txCount/60))
	msg += "\t"
	pendingTxCount, _ := res["pendingTxCount"].(json.Number).Int64()
	msg += fmt.Sprintf("%d", pendingTxCount)
	msg += "\t"
	totalTxCount, _ := res["totalTxCount"].(json.Number).Int64()
	msg += fmt.Sprintf("%d", totalTxCount)
	msg += "\t"
	blockCount60s, _ := res["blockCount60s"].(json.Number).Float64()
	msg += fmt.Sprintf("%2.2f", blockCount60s/60)
	msg += "\t"
	staleBlockCount60s, _ := res["staleBlockCount60s"].(json.Number).Float64()
	msg += fmt.Sprintf("%2.2f", staleBlockCount60s/60)
	msg += "\t"
	cpuf := res["cpus"].([]interface{})
	var total float64
	for _, p := range cpuf {
		n, _ := p.(json.Number).Float64()
		total += n
	}
	mean := total / float64(len(cpuf))
	msg += fmt.Sprintf("%2.2f", mean)

	msg += "\t"
	rh, _ := res["rootHeight"].(json.Number).Int64()
	msg += fmt.Sprintf("%d", rh)

	msg += "\t"
	shardsi := res["shards"].([]interface{})
	shards := make([]string, len(shardsi))
	for i, p := range shardsi {
		shard := p.(map[string]interface{})
		shards[i] = fmt.Sprintf("%s/%s-%s", shard["chainId"], shard["shardId"], shard["height"])
	}
	msg += strings.Join(shards, " ")
	return msg
}

func queryAddress(client jsonrpc.RPCClient, interval *uint, address, token *string) {
	addr := *address
	if strings.HasPrefix(addr, "0x") {
		addr = addr[2:]
	}
	if len(addr) != 48 {
		fmt.Printf("Err: invalid address %x\n", address)
		return
	}
	fmt.Printf("Querying balances for 0x%s\n", addr)
	titles := []string{"Timestamp\t", "Total", fmt.Sprintf("Shards (%s)", *token)}
	fmt.Println(strings.Join(titles, "\t"))

	queryBalance(client, addr, *token)
	intv := time.Duration(*interval)
	ticker := time.NewTicker(intv * time.Second)
	for {
		select {
		case <-ticker.C:
			queryBalance(client, addr, *token)
		}
	}
}

func queryBalance(client jsonrpc.RPCClient, addr, token string) {
	accBytes, err := hexutil.Decode("0x" + addr)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	acc, err := account.CreatAddressFromBytes(accBytes)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	acc.FullShardKey = 0
	includeShards := true
	response, err := client.Call("getAccountData", qkcapi.GetAccountDataArgs{Address: acc, IncludeShards: &includeShards})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if response.Error != nil {
		fmt.Println(response.Error.Error())
		return
	}
	res := response.Result.(map[string]interface{})
	shardsi := res["shards"].([]interface{})
	shardsQKCStr := make([]string, 0)
	total := big.NewInt(0)
	for _, p := range shardsi {
		shardMap := p.(map[string]interface{})
		balanceMaps := shardMap["balances"].([]interface{})
		//fmt.Println("chainId", shardMap["chainId"])
		if len(balanceMaps) > 0 {
			for _, s := range balanceMaps {
				balanceMap := s.(map[string]interface{})
				tokenStr := balanceMap["tokenStr"].(string)
				if strings.Compare(strings.ToUpper(tokenStr), strings.ToUpper(token)) == 0 {
					balanceWei, _ := new(big.Int).SetString(balanceMap["balance"].(string)[2:], 16)
					total = total.Add(total, balanceWei)
					balance := balanceWei.Div(balanceWei, big.NewInt(1000000000000000000))
					shardsQKCStr = append(shardsQKCStr, balance.String())
				}
			}
		} else {
			shardsQKCStr = append(shardsQKCStr, "0")
		}
	}
	total = total.Div(total, big.NewInt(1000000000000000000))
	shardsQKCs := strings.Join(shardsQKCStr, ", ")
	t := time.Now()
	msg := t.Format("2006-01-02 15:04:05")
	msg += "\t"
	msg += total.String()
	msg += "\t"
	msg += shardsQKCs
	fmt.Println(msg)
}

func cConnHandler(c net.Conn, client jsonrpc.RPCClient, interval *uint) {

	// var genesisAllocTxCount = int64(102365)
	var genesisAllocTxCount = int64(66371700)
	totalTxCountMod := int64(0)
	NewTotalTxCountMod := int64(0)
	EpochInterval := int64(240000)
	EPOCH := int64(1)
	EpochMAX := int64(20)
	Checkpoint := 0
	fileName := "TPS_nomig_8_6_May6.csv"

	var txCountStringToCSV [][]string

	intv := time.Duration(*interval)
	ticker := time.NewTicker(intv * time.Second)

	//缓存 conn 中的数据
	buf := make([]byte, 4096)

	fmt.Println("请输入客户端请求数据...")

	for {

		select {
		case <-ticker.C:
			response, err := client.Call("getStats")
			if err != nil {
				fmt.Println(err.Error())
			}
			if response.Error != nil {
				fmt.Println(response.Error.Error())
			}
			res := response.Result.(map[string]interface{})
			pendingTxCount, _ := res["pendingTxCount"].(json.Number).Int64()
			confirmedTxCount, _ := res["totalTxCount"].(json.Number).Int64()

			txCount, _ := res["txCount60s"].(json.Number).Int64()
			blockCount60s, _ := res["blockCount60s"].(json.Number).Float64()

			var txCountString []string
			if txCount != 0 {
				txCountString = append(txCountString, fmt.Sprintf("%2.2f", float64(txCount/60)))
				txCountString = append(txCountString, fmt.Sprintf("%2.2f", blockCount60s/60))

				txCountStringToCSV = append(txCountStringToCSV, txCountString)
			}

			NewTotalTxCountMod = (pendingTxCount + confirmedTxCount - genesisAllocTxCount) % EpochInterval

			if pendingTxCount+confirmedTxCount > genesisAllocTxCount && NewTotalTxCountMod < totalTxCountMod && EPOCH < EpochMAX && Checkpoint > 36 {
				EPOCH++
				EPOCHstring := fmt.Sprintf("%d", EPOCH)
				//去除输入两端空格
				EPOCHstring = strings.TrimSpace(EPOCHstring)
				//客户端请求数据写入 conn，并传输
				c.Write([]byte(EPOCHstring))
				//服务器端返回的数据写入空buf
				cnt, err := c.Read(buf)

				if err != nil {
					fmt.Printf("客户端读取数据失败 %s\n", err)
					continue
				}

				//回显服务器端回传的信息
				fmt.Print("服务器端回复" + string(buf[0:cnt]))

				Checkpoint = 0

			}

			totalTxCountMod = NewTotalTxCountMod

			Checkpoint++

			WriteTPSToCSV(txCountStringToCSV, fileName)

		}

	}
}

func ClientSocket(client jsonrpc.RPCClient, interval *uint) {
	conn, err := net.Dial("tcp", "54.193.209.253:8087")
	if err != nil {
		fmt.Println("客户端建立连接失败")
		return
	}

	cConnHandler(conn, client, interval)
}

// WriteTPSToCSV 讲TPS信息写入CSV
func WriteTPSToCSV(TPS_tocsv [][]string, fileName string) {
	// 写入csv文件
	f, err := os.Create(fileName) //创建文件
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := csv.NewWriter(f) //创建一个新的写入文件流
	// WriteAll方法使用Write方法向w写入多条记录，并在最后调用Flush方法清空缓存。
	w.WriteAll(TPS_tocsv)
	w.Flush()
}

func main() {

	ip := flag.String("ip", "localhost", "Cluster IP")
	interval := flag.Uint("i", 5, "Query interval in second")
	address := flag.String("a", "", "Query account balance if a QKC address is provided")
	token := flag.String("t", "QKC", "Query account balance for a specific token")
	flag.Parse()
	privateEndPoint := jsonrpc.NewClient(fmt.Sprintf("http://%s:38491", *ip))
	publicEndPoint := jsonrpc.NewClient(fmt.Sprintf("http://%s:38391", *ip))

	go ClientSocket(privateEndPoint, interval)

	fmt.Println(basic(privateEndPoint, *ip))
	if len(*address) > 0 {
		queryAddress(publicEndPoint, interval, address, token)
	} else {
		queryStats(privateEndPoint, interval)
	}
}

// func main() {

// 	ip := flag.String("ip", "localhost", "Cluster IP")
// 	interval := flag.Uint("i", 10, "Query interval in second")
// 	// address := flag.String("a", "", "Query account balance if a QKC address is provided")
// 	// token := flag.String("t", "QKC", "Query account balance for a specific token")
// 	flag.Parse()
// 	privateEndPoint := jsonrpc.NewClient(fmt.Sprintf("http://%s:38491", *ip))
// 	publicEndPoint := jsonrpc.NewClient(fmt.Sprintf("http://%s:38391", *ip))
// 	QueryDetails(privateEndPoint, publicEndPoint, interval)
// }

// func QueryDetails(priclient jsonrpc.RPCClient, pubclient jsonrpc.RPCClient, interval *uint) {
// 	titles := []string{"Timestamp\t", "Syncing", "TPS", "Pend.TX", "Conf.TX", "BPS", "SBPS", "CPU", "ROOT", "CHAIN/SHARD-HEIGHT"}
// 	fmt.Println(strings.Join(titles, "\t"))
// 	intv := time.Duration(*interval)
// 	ticker := time.NewTicker(intv * time.Second)
// 	fmt.Println(Details(priclient, pubclient))
// 	for {
// 		select {
// 		case <-ticker.C:
// 			fmt.Println(Details(priclient, pubclient))
// 		}
// 	}
// }

// func Details(priclient jsonrpc.RPCClient, pubclient jsonrpc.RPCClient) interface{} {
// 	response, err := priclient.Call("getStats")
// 	if err != nil {
// 		return err.Error()
// 	}
// 	if response.Error != nil {
// 		return response.Error.Error()
// 	}
// 	res := response.Result.(map[string]interface{})
// 	shardsi := res["shards"].([]interface{})
// 	shards := make([]string, len(shardsi))
// 	var txs interface{}
// 	for i, p := range shardsi {
// 		shard := p.(map[string]interface{})
// 		shards[i] = fmt.Sprintf("%x/%x-%x", shard["chainId"], shard["shardId"], shard["height"])

// 		fullShardId, _ := shard["fullShardId"].(json.Number).Int64()
// 		height, _ := shard["height"].(json.Number).Int64()
// 		fmt.Println(height)
// 		resp, err := pubclient.Call("getMinorBlockByHeight", hexutil.EncodeUint64(uint64(fullShardId)), hexutil.EncodeUint64(uint64(height)), true)
// 		if err != nil {
// 			fmt.Println(err.Error())
// 		}
// 		// fmt.Println(resp)
// 		txs = resp.Result.(map[string]interface{})["transactions"]
// 		fmt.Println(fmt.Sprintf("%T", txs.([]interface{})))
// 	}
// 	// msg := strings.Join(shards, " ")

// 	return txs
// }
