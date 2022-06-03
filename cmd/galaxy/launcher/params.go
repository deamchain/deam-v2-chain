package launcher

import (
	"github.com/ethereum/go-ethereum/params"
)

var (
	Bootnodes = []string{
		// testnet
		"enode://66fe467ba140632bb8c47d505ea3e4b23418818c33804f15402227e47e734b141abc024982cb412d17da40e7d5fb8127b34f1aa6300b3c04cc0ce7af3fdad555@65.108.43.238:18887",
		"enode://c6e1bf0b0c2654017140930e38f21bcfc6685aa49d37ad4fb5d84551b1a78dc001578cf422c5a94c413621eec825137f3e875cccf63833b96114d125f1eaa684@65.108.124.81:18887",
		"enode://3af9605199cc0fe8ce7f9ae5ff5147092f30ec2cb83c39b7a8b276cd5c6c313895930653c6e5aef01526953db5d1fedb3896a54f72b1d287b707573207f657c6@65.108.2.89:18887",
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
