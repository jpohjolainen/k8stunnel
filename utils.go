package main

import (
  "os"
  "fmt"
  "strconv"
  "math/rand"
)

func parseInt32OrExit(str string) int32 {
  num, err := strconv.ParseInt(str, 10, 32)
  if err != nil {
    fmt.Printf("Error parsing '%s' as port number.\n", str)
    os.Exit(3)
  }
  return int32(num)
}

func randomChars(len int) string {
  bytes := make([]byte, len)
  for i := 0; i < len; i++ {
    bytes[i] = byte(97 + rand.Intn(26))
  }
  return string(bytes)
}

func randomHighPort() int32 {
  return int32(10000 + rand.Intn(50000))
}
