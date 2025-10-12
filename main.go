package main

import (
	"fmt"

	"github.com/slajuwomi/gator/internal/config"
)

func main() {
	configStruct := config.Read()
	configStruct.SetUser("stephen")
	testConfigStruct := config.Read()
	fmt.Printf("%+v", testConfigStruct)
}
