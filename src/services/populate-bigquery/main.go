package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"populate/config"
	"populate/service"
)

// Lembrar de fechar os Clientes

func main() {

	ctx := context.Background()
	_ = godotenv.Load()

	_, csClient, file, err := config.InitCloudStorage(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer func() {
		if err := csClient.Close(); err != nil {
			log.Printf("error closing storage client: %v", err)
		}
	}()

	bqClient, dataset, err := config.InitBigQuery(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer func() {
		if err := bqClient.Close(); err != nil {
			log.Printf("error closing bigquery client: %v", err)
		}
	}()

	err = service.LoadCsvIntoBigQuery(dataset, file, ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("")

}
