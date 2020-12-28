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

	if err := cron.JobAdd("*/10 * * * * *", test2); err != nil {
		t.Log(err.Error())
		return
	}

	cron.Start()

	fmt.Println("=== Start ===")

	go func() {
		time.Sleep(time.Second * 5)
		if err := cron.JobAdd("* * * * * *", test1); err != nil {
			t.Log(err.Error())
			return
		}
	}()

	time.Sleep(60 * time.Second)

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
	// ce, err := obtainCronExpression("0-9,20-29/2,40-49 0,10,20,30,40,50 * 1 aug *")
	ce, err := obtainCronExpression("* * * * * *")
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
	j, err := obtainJob("0 30 */2 * * *", funcName, HelloWorld)
	if err != nil {
		t.Error(err.Error())
		return
	}
	for h := 0; h < 24; h++ {
		for m := 0; m < 60; m++ {
			for s := 0; s < 60; s += 10 {
				now := time.Date(2020, time.Month(12), 28, h, m, s, 0, time.Now().Location())
				j.updateNextTime(now)
				fmt.Println("现在:", now.String(), "下次:", j.nextTime.String())
			}
		}
	}
}
