package pool

import (
	"testing"
	"time"
)

func BenchmarkPool(b *testing.B) {
	p := NewPool(10)
	for i := 0; i < b.N; i++ {
		go func(a int) {
			if e := p.Schedule(func() {
				time.Sleep(time.Second * 4)
				b.Logf("#%v: Done\n", a)
			}); e != nil {
				b.Logf("#%v: %v\n", a, e)
			}
		}(i)
	}
	if e := p.Stop(time.Second * 5); e != nil {
		b.Logf("%v\n", e)
	} else {
		b.Logf("All tasks completed\n")
	}
}