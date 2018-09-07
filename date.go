package php

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	_tz   string
	_zone *time.Location
)

var _err error

func init() {
	DateDefaultTimezoneSet("UTC")
}

// Time - php time()
func Time() int64 {
	return nowTime().Unix()
}

// Strtotime - php strtotime()
func Strtotime(layout string, timestamps ...int64) int64 {
	var diff int64
	now := Time()
	if len(timestamps) > 0 {
		now = timestamps[0]
	}
	if strings.ToLower(layout) == "now" {
		return now
	}
	reg, _ := regexp.Compile("(?:\\+|\\-)?(?:\\s?(\\d+)\\s?(year|month|week|day|hour|minute|second)s?)")
	matches := reg.FindAllStringSubmatch(layout, -1)
	if len(matches) > 0 {
		m := make(map[string]int64)
		for i := 0; i < len(matches); i++ {
			num, _ := strconv.ParseInt(matches[i][1], 10, 64)
			m[matches[i][2]] = num
		}
		for k, v := range m {
			switch k {
			case "second":
				diff += v
			case "minute":
				diff += v * 60
			case "hour":
				diff += v * 3600
			case "day":
				diff += v * 86400
			case "week":
				diff += v * 86400 * 7
			case "month":
				diff += v * 86400 * 30
			case "year":
				diff += v * 86400 * 365
			}
		}
		if string([]byte(layout)[0]) == "-" {
			return now - diff
		}
		return now + diff
	}
	for _, p := range _defaultPatterns {

		reg = regexp.MustCompile(p.regexp)

		if reg.MatchString(layout) {
			t, err := time.Parse(p.layout, layout)
			if err == nil {
				return t.Unix()
			}
		}
	}
	return -1
}

// Date - php date()
func Date(layout string, timestamps ...int64) string {
	var date bytes.Buffer
	now := Time()
	if len(timestamps) > 0 {
		now = timestamps[0]
	}
	reg, _ := regexp.Compile("(?:[^\\d]|\\d+)")
	matches := reg.FindAllStringSubmatch(layout, -1)
	tm := time.Unix(now, 0).In(_zone)
	if len(matches) > 0 {
		for i := 0; i < len(matches); i++ {
			reg1, _ := regexp.Compile("[aABcdDeFhHgGiIjlLmMnNoOPrsStTuUvwWyYzZ]")
			match := reg1.FindStringSubmatch(matches[i][0])
			if len(match) > 0 {
				date.WriteString(recognize(match[0], tm))
			} else {
				date.WriteString(matches[i][0])
			}
		}
	}
	return date.String()
}

// DateDefaultTimezoneSet - php date_default_timezone_set
func DateDefaultTimezoneSet(tz string) bool {
	_tz = tz
	_zone, _err = time.LoadLocation(tz)
	if _err != nil {
		return false
	}
	return true
}

// DateDefaultTimezoneGet - php date_default_timezone_get
func DateDefaultTimezoneGet() string {
	return _tz
}

// LastDateOfMonth gets the last date of the month which the given time is in
func LastDateOfMonth(timestamps ...int64) int64 {
	tm := nowTime(timestamps...)
	t2 := FirstDateOfNextMonth(tm.Unix())
	return t2 - 86400
}

// FirstDateOfMonth gets the first date of the month which the given time is in
func FirstDateOfMonth(timestamps ...int64) int64 {
	tm := nowTime(timestamps...)
	year, month, _ := tm.Date()
	return time.Date(year, month, 1, tm.Hour(), tm.Minute(), tm.Second(), 0, _zone).Unix()
}

// FirstDateOfNextMonth gets the first date of next month
func FirstDateOfNextMonth(timestamps ...int64) int64 {
	tm := nowTime(timestamps...)
	year, month, _ := tm.Date()
	if month == time.December {
		year++
		month = time.January
	} else {
		month++
	}
	return time.Date(year, month, 1, tm.Hour(), tm.Minute(), tm.Second(), 0, _zone).Unix()
}

// FirstDateOfLastMonth gets the first date of last month
func FirstDateOfLastMonth(timestamps ...int64) int64 {
	tm := nowTime(timestamps...)
	year, month, _ := tm.Date()
	if month == time.January {
		year--
		month = time.December
	} else {
		month--
	}
	return time.Date(year, month, 1, tm.Hour(), tm.Minute(), tm.Second(), 0, _zone).Unix()
}

// Mktime - php mktime()
func Mktime(hour int64, min int64, sec int64, mon time.Month, day int64, year int64) int64 {
	return time.Date(int(year), mon, int(day), int(hour), int(min), int(sec), 0, _zone).Unix()
}

// LastWeekday get last monday to sunday
// eg. LastWeekday(7) get last sunday
func LastWeekday(day time.Weekday, timestamps ...int64) int64 {
	tm := nowTime(timestamps...)
	now := int(tm.Weekday())
	diff := (int(day) - now + 7) * 86400
	return tm.Unix() + int64(diff)
}

// NextWeekday get next monday to sunday
// eg. NextWeekda(7) get next sunday
func NextWeekday(day time.Weekday, timestamps ...int64) int64 {
	tm := nowTime(timestamps...)
	now := int(tm.Weekday())
	diff := (int(day) - now - 7) * 86400
	return tm.Unix() + int64(diff)
}

// isLeapYear checks if the given time is in a leap year
func isLeapYear(tm time.Time) bool {
	return tm.YearDay() == 366
}

// get time package from timestamp or not
func nowTime(timestamps ...int64) time.Time {
	if len(timestamps) > 0 {
		return time.Unix(timestamps[0], 0).In(_zone)
	}
	return time.Now().In(_zone)
}

// recognize the character in the php date/time format string
func recognize(c string, tm time.Time) string {
	switch c {
	// Year
	case "L": // Whether it's a leap year
		if isLeapYear(tm) {
			return "1"
		}
		return "0"
	case "o": // ISO-8601 week-numbering year. This has the same value as Y, except that if the ISO week number (W) belongs to the previous or next year, that year is used instead.
		fallthrough
	case "Y": // A full numeric representation of a year, 4 digits
		return strconv.Itoa(tm.Year())
	case "y": // A two digit representation of a year
		return tm.Format("06")

	// Month
	case "F": // A full textual representation of a month
		return tm.Month().String()
	case "m": // Numeric representation of a month, with leading zeros
		return fmt.Sprintf("%02d", tm.Month())
	case "M": // A short textual representation of a month, three letters
		return tm.Format("Jan")
	case "n": // Numeric representation of a month, without leading zeros
		return fmt.Sprintf("%d", tm.Month())
	case "t": // Number of days in the given month
		return nowTime(LastDateOfMonth(tm.Unix())).Format("2")

	// Week
	case "W": // ISO-8601 week number of year, weeks starting on Monday
		_, w := tm.ISOWeek()
		return fmt.Sprintf("%d", w)

	// Day
	case "d": // Day of the month, 2 digits with leading zeros
		return tm.Format("02")
	case "D": // A textual representation of a day, three letters
		return tm.Format("Mon")
	case "j": // Day of the month without leading zeros
		return strconv.Itoa(tm.Day())
	case "l": // A full textual representation of the day of the week
		return tm.Weekday().String()
	case "N": // ISO-8601 numeric representation of the day of the week
		return fmt.Sprintf("%d", tm.Weekday()+1)
	case "S": // English ordinal suffix for the day of the month, 2 characters
		suffix := [4]string{"", "st", "nd", "rd"}
		day := tm.Day()
		if day > 3 {
			return "th"
		}
		return suffix[int64(day)]
	case "w": // Numeric representation of the day of the week
		return fmt.Sprintf("%d", tm.Weekday())
	case "z": // The day of the year (starting from 0)
		return fmt.Sprintf("%d", tm.YearDay()-1)

	// Time
	case "a": // Lowercase Ante meridiem and Post meridiem
		return tm.Format("pm")
	case "A": // Uppercase Ante meridiem and Post meridiem
		return strings.ToUpper(tm.Format("pm"))
	case "B": // Swatch Internet time
		_, offset := tm.Zone()
		x := offset/3600 - 1
		s := tm.Second()
		h := tm.Hour()
		m := tm.Minute()
		d := h - x
		if d < 0 {
			d += 24
		}
		f := 1000 * (60*(60*(d)+m) + s) / 86400
		return strconv.Itoa(f)
	case "g": // 12-hour format of an hour without leading zeros
		return tm.Format("3")
	case "G": // 24-hour format of an hour without leading zeros
		return fmt.Sprintf("%d", tm.Hour())
	case "h": // 12-hour format of an hour with leading zeros
		return tm.Format("03")
	case "H": // 24-hour format of an hour with leading zeros
		return fmt.Sprintf("%02d", tm.Hour())
	case "i": // Minutes with leading zeros
		return fmt.Sprintf("%02d", tm.Minute())
	case "s": // Seconds, with leading zeros
		return fmt.Sprintf("%02d", tm.Second())
	case "u": // Microseconds. Note that date() will always generate 000000 since it takes an integer parameter, whereas DateTime::format() does support microseconds if DateTime was created with microseconds.
		return "000000" // todo
	case "v": // Milliseconds. Same note applies as for u.
		return "000" // todo

	// Timezone
	case "e": // Timezone identifier
		return _tz
	case "I": // Whether or not the date is in daylight saving time
		return "0" // todo
	case "O": // Difference to Greenwich time (GMT) in hours
		return tm.Format("-0700")
	case "P": // Difference to Greenwich time (GMT) with colon between hours and minutes
		return tm.Format("-07:00")
	case "T": // Timezone abbreviation
		return tm.Format("MST")
	case "Z": // Timezone offset in seconds. The offset for timezones west of UTC is negative (-43200 to 50400)
		_, offset := tm.Zone()
		return strconv.Itoa(offset)

	// Full Data/Time
	case "c": // ISO 8601 date
		return tm.Format("2006-01-02T15:04:05Z07:00")
	case "r": // RFC 2822 formatted date
		return tm.Format("Mon, 02 Jan 2006 15:04:05 -0700")
	case "U": // Seconds since the Unix Epoch (January 1 1970 00:00:00 GMT)
		return fmt.Sprintf("%v", tm.Unix())
	}
	return ""
}
