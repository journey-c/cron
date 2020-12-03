package cron

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	dowAry = map[string]int{
		"sun": 0,
		"mon": 1,
		"tue": 2,
		"wed": 3,
		"thu": 4,
		"fri": 5,
		"sat": 6,
	}
	monAry = map[string]int{
		"jan": 1,
		"feb": 2,
		"mar": 3,
		"apr": 4,
		"may": 5,
		"jun": 6,
		"jul": 7,
		"aug": 8,
		"sep": 9,
		"oct": 10,
		"nov": 11,
		"dec": 12,
	}
)

type job struct {
	expr    string
	cmdName string
	cmd     func()

	cronExpr *cronExpression
	nextTime time.Time
}

func obtainJob(expr, cmdName string, cmd func()) (*job, error) {
	cExpr, err := obtainCronExpression(expr)
	if err != nil {
		return nil, err
	}

	return &job{
		expr:     expr,
		cmdName:  cmdName,
		cmd:      cmd,
		cronExpr: cExpr,
	}, nil
}

func (j *job) updateNextTime() {
	now := time.Now().Add(time.Second - time.Nanosecond) // deviation

	t := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), 0, now.Location())
	yearLimit := now.Year() + 5 // leap year

	var skip bool
	var operated = false

	for t.Year() < yearLimit {
		// month
		skip = false
		for (1<<(t.Month()-1))&j.cronExpr.month == 0 {
			t = t.AddDate(0, 1, 0)
			if operated == false {
				t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
				operated = true
			}
			if t.Month() == time.January {
				skip = true
				break
			}
		}
		if skip == true {
			continue
		}
		// day
		for (1<<(t.Day()-1))&j.cronExpr.day == 0 || (1<<(uint(t.Weekday())-1))&j.cronExpr.week == 0 {
			t = t.AddDate(0, 0, 1)
			if operated == false {
				t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
				operated = true
			}
			if t.Day() == 1 {
				skip = true
				break
			}
		}
		if skip == true {
			continue
		}
		// hour
		for (1<<t.Hour())&j.cronExpr.hour == 0 {
			t = t.Add(time.Second * time.Duration(3600))
			if operated == false {
				t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
				operated = true
			}
			if t.Hour() == 0 {
				skip = true
				break
			}
		}
		if skip == true {
			continue
		}
		// minute
		for (1<<t.Minute())&j.cronExpr.minute == 0 {
			t = t.Add(time.Second * time.Duration(60))
			if operated == false {
				t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location())
				operated = true
			}
			if t.Minute() == 0 {
				skip = true
				break
			}
		}
		if skip == true {
			continue
		}
		// second
		for (1<<t.Second())&j.cronExpr.second == 0 {
			t = t.Add(time.Second)
			if t.Second() == 0 {
				skip = true
				break
			}
		}
		if skip == false {
			break
		}
	}
	j.nextTime = t
	return
}

type cronExpression struct {
	second uint64 // 0-59
	minute uint64 // 0-59
	hour   uint64 // 0-23
	day    uint64 // 0-30
	month  uint64 // 0-11
	week   uint64 // 0-7
}

func obtainCronExpression(expr string) (*cronExpression, error) {
	expr = strings.TrimSpace(expr)

	if expr[0] == '@' {
		ce, err := parseSpecialExpression(expr)
		if err != nil {
			return nil, err
		}
		return ce, nil
	}

	ce, err := parseExpression(expr)
	if err != nil {
		return nil, err
	}

	return ce, nil
}

func parseSpecialExpression(expr string) (*cronExpression, error) {
	switch expr {
	case "@yearly", "@annually":
		return &cronExpression{
			second: getBitsByArray(0),
			minute: getBitsByArray(0),
			hour:   getBitsByArray(0),
			day:    getBitsByArray(0),
			month:  getBitsByArray(0),
			week:   getBitsByBound(0, 6, 1),
		}, nil
	case "@monthly":
		return &cronExpression{
			second: getBitsByArray(0),
			minute: getBitsByArray(0),
			hour:   getBitsByArray(0),
			day:    getBitsByArray(0),
			month:  getBitsByBound(0, 11, 1),
			week:   getBitsByBound(0, 6, 1),
		}, nil
	case "@weekly":
		return &cronExpression{
			second: getBitsByArray(0),
			minute: getBitsByArray(0),
			hour:   getBitsByArray(0),
			day:    getBitsByBound(0, 30, 1),
			month:  getBitsByBound(0, 11, 1),
			week:   getBitsByBound(0, 6, 1),
		}, nil
	case "@daily", "@midnight":
		return &cronExpression{
			second: getBitsByArray(0),
			minute: getBitsByArray(0),
			hour:   getBitsByArray(0),
			day:    getBitsByBound(0, 30, 1),
			month:  getBitsByBound(0, 11, 1),
			week:   getBitsByBound(0, 6, 1),
		}, nil
	case "@hourly":
		return &cronExpression{
			second: getBitsByArray(0),
			minute: getBitsByArray(0),
			hour:   getBitsByBound(0, 23, 1),
			day:    getBitsByBound(0, 30, 1),
			month:  getBitsByBound(0, 11, 1),
			week:   getBitsByBound(0, 6, 1),
		}, nil
	case "@reboot":
		return nil, errors.New("not support @reboot")
	default:
		return nil, errors.New("invalid expression " + expr)
	}
}

func parseExpression(expr string) (*cronExpression, error) {
	fields := strings.Fields(expr)
	switch len(fields) {
	case 5:
		fields = append([]string{"0"}, fields...)
	case 6:
	default:
		return nil, errors.New("invalid expression " + expr)
	}

	ce := new(cronExpression)

	if ce.second = parseUnit(fields[0], 60, 0, nil, false); ce.second == 0 {
		return nil, fmt.Errorf("bad %s: %s", "second", fields[0])
	}
	if ce.minute = parseUnit(fields[1], 60, 0, nil, false); ce.minute == 0 {
		return nil, fmt.Errorf("bad %s: %s", "minute", fields[1])
	}
	if ce.hour = parseUnit(fields[2], 24, 0, nil, false); ce.hour == 0 {
		return nil, fmt.Errorf("bad %s: %s", "hour", fields[2])
	}
	if ce.day = parseUnit(fields[3], 31, -1, nil, false); ce.day == 0 {
		return nil, fmt.Errorf("bad %s: %s", "day", fields[3])
	}
	if ce.month = parseUnit(fields[4], 12, -1, monAry, false); ce.month == 0 {
		return nil, fmt.Errorf("bad %s: %s", "month", fields[4])
	}
	if ce.week = parseUnit(fields[5], 8, 0, dowAry, true); ce.week == 0 {
		return nil, fmt.Errorf("bad %s: %s", "week", fields[5])
	}
	return ce, nil
}

// (a | b-c [/d]), (e | f-g [/h])
func parseUnit(field string, modvalue, offset int, name map[string]int, special bool) uint64 {
	var result uint64

	var pos = 0
	var endp = len(field)
	var tmp int
	var l = -1
	var r = -1

	for pos < endp {
		skip := 0

		if field[pos] == '*' {
			l = 0
			r = modvalue - 1
			pos++
			skip = 1
		} else if isDigit(field[pos]) {
			tmp = 0
			for pos < endp && isDigit(field[pos]) {
				tmp *= 10
				tmp += int(field[pos] - '0')
				pos++
			}
			tmp += offset
			if l < 0 {
				l = tmp
			} else {
				r = tmp
			}
			skip = 1
		} else if name != nil {
			if pos+3 <= endp {
				for k, v := range name {
					if k == field[pos:pos+3] {
						if l < 0 {
							l = v + offset
						} else {
							r = v + offset
						}
						skip = 1
						pos += 3
						break
					}
				}
			}
		}

		if skip == 0 {
			return 0
		}

		if pos != endp && field[pos] == '-' && r < 0 {
			pos++
			continue
		}

		if l < 0 {
			return 0
		}

		if r < 0 {
			r = l
		}

		if pos != endp && field[pos] == '/' {
			pos++
			tmp = 0
			for pos < endp && isDigit(field[pos]) {
				tmp *= 10
				tmp += int(field[pos] - '0')
				pos++
			}
			if tmp == 0 {
				return 0
			}
			skip = tmp
		}

		{
			var s0 = 1
			var failsafe = 1024
			l--
			for {
				l = (l + 1) % modvalue

				s0--
				if s0 == 0 {
					result |= (1 << l)
					s0 = skip
				}

				failsafe--
				if failsafe == 0 {
					return 0
				}

				if l == r {
					break
				}
			}
		}

		if pos >= endp || field[pos] != ',' {
			break
		}
		pos++
		l = -1
		r = -1
	}
	if pos != endp {
		return 0
	}
	return result
}
