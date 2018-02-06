package main

import (
	"github.com/KyberNetwork/reserve-data/cmd/verification"
	"github.com/KyberNetwork/reserve-data/http"
	"github.com/KyberNetwork/reserve-data/signer"
)

func main() {
	fileSigner := signer.NewFileSigner("/go/src/github.com/KyberNetwork/reserve-data/cmd/config.json")
	hmac512auth := http.KNAuthentication{
		fileSigner.KNSecret,
		fileSigner.KNReadOnly,
		fileSigner.KNConfiguration,
	}
	verify := verification.NewVerification(hmac512auth)
	verify.RunVerification()
}
