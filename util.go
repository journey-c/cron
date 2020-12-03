package cron

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
)

func obtainFuncName(cmd interface{}) (cmdName string, err error) {
	defer func() {
		if err := recover(); err != nil {
			err = fmt.Errorf("%v", err)
		}
	}()
	cmdName = runtime.FuncForPC(reflect.ValueOf(cmd).Pointer()).Name()
	return
}

func getBitsByBound(min, max, step uint64) uint64 {
	if step == 1 {
		return uint64(^(math.MaxUint64 << (max + 1)) & (math.MaxUint64 << min))
	}

	var src uint64
	for i := min; i <= max; i += step {
		src |= (1 << i)
	}
	return src
}

func getBitsByArray(ary ...uint64) uint64 {
	var src uint64
	for _, a := range ary {
		src |= a
	}
	return src
}

func isDigit(b byte) bool {
	if b-'0' >= 0 && b-'0' <= 9 {
		return true
	}
	return false
}
