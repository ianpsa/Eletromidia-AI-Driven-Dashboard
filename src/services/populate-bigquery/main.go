package main

import (
    "populate/service"
    "populate/config"
    "context"
    "os"
    "fmt"
    "github.com/joho/godotenv"
)

// Lembrar de fechar os Clientes

func main() {

    ctx := context.Background()
    godotenv.Load()

    _, csClient, file, err := config.InitCloudStorage(ctx)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    defer csClient.Close()

    bqClient, dataset, err := config.InitBigQuery(ctx)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    defer bqClient.Close()

    err = service.LoadCsvIntoBigQuery(dataset, file, ctx)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    fmt.Println("")


}