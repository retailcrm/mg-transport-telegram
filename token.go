package main

import (
	"crypto/sha256"
	"fmt"
	"sync/atomic"
	"time"
)

var tokenCounter uint32

//Generate token method
func GenerateToken() string {
	c := atomic.AddUint32(&tokenCounter, 1)

	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%d%d", time.Now().UnixNano(), c))))
}
