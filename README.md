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
	fmt.Println(php.Date("Y-m-d H:i:s", php.Strtotime("+1 week 2 hours 3day")))
	fmt.Println(php.Mktime(5, 50, 4, 1, 7, 2014))
}
```
