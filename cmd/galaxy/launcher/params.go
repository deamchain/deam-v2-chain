package launcher

import (
	"github.com/ethereum/go-ethereum/params"
)

var (
	Bootnodes = []string{
		// testnet
		"enode://67a216ca4355c63e0fba84f2abe2148ac6d8491890efd9c38068555fcd18735d3e718097f16efff9728ba76c288cb84ee4210ac64b6f5993b698257b0fe425ed@65.108.43.238:18887",
		"enode://916cecb1395f4154448247c5f7452c0da6062c7ca7b42b3128a336bbd293fa20867845e463a807cf8da1366eafa8f4b99ebf41fcb0fa57069f4c8954375c97ed@65.108.124.81:18887",
		"enode://33c06dab605f348d6cf14acaa6ab7c8857fca9a8b0f0be50b2183b916007fd0c19f3937f99bbd369886c68ddbe037b6d6ea3f7293c11a1a89d41b67c19227de4@65.108.2.89:18887",
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
