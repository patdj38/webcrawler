package main

import (
	"fmt"
	//"github.com/spf13/cobra"
	//"github.com/spf13/viper"
	"os"
	"pat/rest"
)

var cfgFile = ""

func initConfig() {
	if cfgFile != "" {
		// enable ability to specify config file via flag
	}
}

func init() {
	//cobra.OnInitialize(initConfig)
}

func main() {
	shutdownCh := make(chan struct{})
	err := rest.RunServer(shutdownCh)
	if err != nil {
		fmt.Errorf(err.Error())
		os.Exit(1)
	}
	fmt.Printf("Exiting main...\n")
}
