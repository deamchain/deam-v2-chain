package launcher

import (
	"github.com/ethereum/go-ethereum/params"
)

var (
	Bootnodes = []string{
		// testnet
		"enode://e38d9a1d4295f2aa5c03d13b25e3292f7bdc3bce5898e6e94fd521a9103ddbf6a601d5c1ce53c626854c3092a1938c711eb1302e8787ad0f6ca08076f7150498@65.108.43.238:18887",
		"enode://0b65894a615bd2b2eda6ccbeac9cdd8f5ff56303080359298493c05a2b40fb898231fa62738f1e0d95f95950f579682435fb1ac547c43a865bba60b3283858c0@65.108.124.81:18887",
		"enode://4ade775182f1fff8ccaf18620b36824725ea94cbc9b655ae261873e2a6ea0048b21c5e7bc43a57b6d01214e575a4059802bf3fc23a9d5e15b7ec7c1965e9d049@65.108.2.89:18887",
		// mainnet
		"enode://f424c35794f0d73ea8a6cfb740865d910568ef5df18efe5258510711efe1c8dd107100e95bbe9d4788bad3434fd52c863740bdce7858a2b1e80919fe1af07acb@65.108.141.55:18887",
		"enode://7d45a6987abf274aeae326fa3a99bf18d90e9a695d9504150aae44290ee6830e2366d4f0b84754c0921be782297dd8c2eb848690cfde8ec28966a2abfa443ba0@65.21.89.25:18887",
		"enode://38a5d52445c70cf99fed60e79149faa0cd539fee9c105e6196ebbdfd3895f8b284899f439c2412ad4dc4b4276044d22d855ef5d9abcccd0797a19f4c23816411@95.217.205.113:18887",
	}
)

func overrideParams() {
	params.MainnetBootnodes = []string{}
	params.RopstenBootnodes = []string{}
	params.RinkebyBootnodes = []string{}
	params.GoerliBootnodes = []string{}
}
