package dh

import (
	"fmt"
	"testing"
)

func TestDH(t *testing.T) {
	//p, g公用

	X1, E1 := DHExchange()
	X2, E2 := DHExchange()

	fmt.Println("Secret 1:", X1, E1)
	fmt.Println("Secret 2:", X2, E2)

	//产生对称密钥
	KEY1 := DHKey(X1, E2) //通过X1 私钥 ＋ E2公钥
	KEY2 := DHKey(X2, E1) //通过X2 私钥 ＋ E1公钥

	fmt.Println("KEY1:", KEY1)
	fmt.Println("KEY2:", KEY2)

	if KEY1.Cmp(KEY2) != 0 {
		t.Error("Diffie-Hellman failed")
	}
}

func BenchmarkDH(b *testing.B) {
	for i := 0; i < b.N; i++ {
		X1, E1 := DHExchange()
		X2, E2 := DHExchange()

		KEY1 := DHKey(X1, E2)
		KEY2 := DHKey(X2, E1)
		if KEY1.Cmp(KEY2) != 0 {
			b.Error("Diffie-Hellman failed")
		}
	}
}
