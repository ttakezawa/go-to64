# go-to64

## About

**go-to64** analyzes Golang main package to convert `int`/`uint` to `int64`/`uint64`. This is an experiment tool, so be very careful.

In a 32-bit environment such as CodeForces, this tool may be useful if you want to force the source code of main package to always use 64-bit integers.

## Installation

```bash
go install github.com/ttakezawa/go-to64@latest
```

## Usage

```bash
go-to64 -fix .
```

## Experimental result

This tool will replace type and cast integer literal to int64 or uint64 as shown in the following example.

Before:

```go
package main

import (
	"fmt"
	"math/bits"
)

const (
	A, B int  = 1, 2
	C    uint = 3
	D         = 4
	E         = 5 + int(C)
)

var (
	a, b int  = 1, 2
	c    uint = 3
	d         = 4
	e         = 5 + add(6+a, int(c))
)

func main() {
	s := 15
	s = bits.OnesCount(uint(s))
	for i := 0; i < add(a, b); i++ {
		fmt.Println(i + 100 + add(a, b))
	}
}

func add(a, b int) int {
	return a + b
}
```

After:

```go
package main

import (
	"fmt"
	"math/bits"
)

const (
	A, B int64  = 1, 2
	C    uint64 = 3
	D           = 4
	E           = 5 + int64(C)
)

var (
	a, b int64  = 1, 2
	c    uint64 = 3
	d           = int64(4)
	e           = int64(5) + add(int64(6)+a, int64(c))
)

func main() {
	s := int64(15)
	s = int64(bits.OnesCount64(uint64(s)))
	for i := int64(0); i < add(a, b); i++ {
		fmt.Println(i + int64(100) + add(a, b))
	}
}

func add(a, b int64) int64 {
	return a + b
}
```