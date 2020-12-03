package cron

import (
	"fmt"
	"testing"
	"time"
)

func test1() {
	fmt.Println(time.Now().UnixNano(), "test1")
}

func test2() {
	fmt.Println(time.Now().UnixNano(), "test2")
}

func TestCron(t *testing.T) {
	cron := NewCron()

	for i := 0; i < 1000; i++ {
		if err := cron.JobAdd("* * * * * *", test1); err != nil {
			t.Log(err.Error())
			return
		}
	}

	if err := cron.JobAdd("* * * * * *", test2); err != nil {
		t.Log(err.Error())
		return
	}

	for i := 0; i < 1000; i++ {
		if err := cron.JobRemove("* * * * * *", test1); err != nil {
			t.Log(err.Error())
			return
		}
	}

	time.Sleep(time.Second * 5)

	cron.Start()

	go func() {
		for {
			time.Sleep(time.Second)
			if err := cron.JobAdd("* * * * * *", test1); err != nil {
				t.Log(err.Error())
				return
			}
		}
	}()

	time.Sleep(300 * time.Second)

	cron.Stop()

	time.Sleep(5 * time.Second)
}

func printUint64(a uint64) {
	str := ""
	for i := 0; i < 64; i++ {
		if a%2 == 0 {
			str += "0"
		} else {
			str += "1"
		}
		a /= 2
	}
	reStr := ""
	for i := 63; i >= 0; i-- {
		if i != 63 && (i+1)%8 == 0 {
			reStr += " "
		}
		reStr += str[i : i+1]
	}
	fmt.Println(reStr)

}

func TestObtainCronExpression(t *testing.T) {
	ce, err := obtainCronExpression("0-9,20-29/2,40-49 0,10,20,30,40,50 * 1 aug *")
	if err != nil {
		t.Error(err.Error())
		return
	}
	printUint64(ce.second)
	printUint64(ce.minute)
	printUint64(ce.hour)
	printUint64(ce.day)
	printUint64(ce.month)
	printUint64(ce.week)
}

func HelloWorld() {
	fmt.Println("HelloWorld!")
}

func TestNextTime(t *testing.T) {
	funcName, err := obtainFuncName(HelloWorld)
	if err != nil {
		t.Error(err.Error())
		return
	}
	j, err := obtainJob("* * * * * *", funcName, HelloWorld)
	if err != nil {
		t.Error(err.Error())
		return
	}
	j.updateNextTime()
	fmt.Println("时间:", j.nextTime.String())
}
