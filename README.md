# go-php

```
package main

import (
	"fmt"
	"php"
)

func main() {
	php.DateDefaultTimezoneSet("Asia/Shanghai")
	php.DateDefaultTimezoneGet()
	fmt.Println(php.Date("a A b B c C d D e E f F g G h H i I j J k K l L m M n N o O p P q Q r R s S t T u U v V w W x X y Y z Z", 1536223000))
	fmt.Println(php.Date("Y-m-d H:i:s", php.LastDateOfMonth()))
	fmt.Println(php.Date("Y-m-d H:i:s", php.FirstDateOfMonth()))
	fmt.Println(php.Date("Y-m-d H:i:s", php.LastWeekday(3)))
	fmt.Println(php.Date("Y-m-d H:i:s", php.Mktime(5, 50, 4, 1, 7, 2014)))
	fmt.Println(php.Date("Y-m-d H:i:s", php.Strtotime("+1 week 2 hours 3day")))
	fmt.Println(php.Time())
	fmt.Println(php.Microtime())
	fmt.Println(php.ArrayKeys([]int{1, 2, 3, 4}))
	fmt.Println(php.ArrayKeys([]string{"123"}))
	fmt.Println(php.ArrayKeys(map[string]string{"321": "123"}))
	fmt.Println(php.ArrayKeys(map[int]string{0: "123", 1: "321"}))
	a := php.ArrayKeys([]string{"123", "321"}).([]interface{})
	fmt.Println(a[0])
	for k, v := range a {
		fmt.Println(k, v)
	}
	fmt.Println(php.ArrayValues(php.ArrayKeys([]int{1, 2, 3, 4})))
	fmt.Println(php.ArrayValues(map[int]string{0: "123", 2: "321"}))
	fmt.Println(php.ArrayKeyExists("321", map[string]string{"321": "123"}))
	fmt.Println(php.ArrayKeyExists(321, map[string]string{"321": "123"}))
	fmt.Println(php.ArrayKeyExists(6, []int{1, 2, 3, 4}))
	fmt.Println(php.InArray(6, []int{1, 2, 3, 4}))
	fmt.Println(php.ArrayFilp(map[int]string{0: "123", 2: "321"}))
	fmt.Println(php.ArrayFilp([]int{1, 2, 3, 4}))
	fmt.Println(php.ArrayUnique([]int{1, 2, 3, 2, 3, 4}))
	b := []int{1, 2, 3, 2, 3, 4}
	php.Sort(b)
	fmt.Println(b)
}
```
