package utils

import (
	"context"
	"crypto/rand"
	"io"
	"math/big"
	"os"
	"os/signal"
	"time"
)

var letters = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandString(n int) string {
	ret := make([]rune, n)
	max := big.NewInt(int64(len(letters)))
	for i := range n {
		num, _ := rand.Int(rand.Reader, max)
		ret[i] = letters[num.Int64()]
	}

	return string(ret)
}

func ReadAllString(r io.Reader) string {
	d, _ := io.ReadAll(r)
	return string(d)
}

func SigWatch(end context.CancelFunc, wait time.Duration, sigs ...os.Signal) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, sigs...)

	<-signals
	end()

	<-time.After(wait)
}
