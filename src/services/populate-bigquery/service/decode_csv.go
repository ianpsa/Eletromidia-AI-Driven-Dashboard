package service

import(
    "encoding/csv"
    "io"
    "populate/models"
    "fmt"
    "cloud.google.com/go/bigquery"
    "cloud.google.com/go/storage"
    "context"
    "github.com/google/uuid"
    "strconv"
)


func LoadCsvIntoBigQuery(dataset *bigquery.Dataset, file *storage.ObjectHandle, ctx context.Context) error {

    geodataInserter := dataset.Table("geodata").Inserter()
    targetInserter := dataset.Table("target").Inserter()
    ageInserter := dataset.Table("age").Inserter()
    genderInserter := dataset.Table("gender").Inserter()
    socialClassInserter := dataset.Table("social_class").Inserter()

    reader, err := file.NewReader(ctx)
    if err != nil {
        return fmt.Errorf("[service/decode_csv] Aconteceu um erro ao ler o csv: %v", err)
    }
    defer reader.Close()

    csvReader := csv.NewReader(reader)
    csvReader.LazyQuotes = true

    csvReader.Read()
    var AgeRows         []*models.Age
    var GenderRows      []*models.Gender
    var SociaClassRows  []*models.SocialClass
    var TargetRows      []*models.Target
    var GeodataRows     []*models.Geodata

    IteratingCsvLines:
    for {

        row, err := csvReader.Read()
        if err == io.EOF {
            break IteratingCsvLines
        }

        if err != nil {
            return fmt.Errorf("[service/decode_csv] Ocorreu um erro ao tentar iterar nas linhas do csv: %v", err)
        }

        impressionHour := row[0]
        locationID := row[1]
        uniques := row[2]
        latitude := row[3]
        longitude := row[4]
        ufEstado := row[5]
        cidade := row[6]
        endereco := row[7]
        numero := row[8]
        target := row[9]


        targetStruct, err := ParseTarget(target)
        if err != nil {
            continue IteratingCsvLines
        }

        var ageSchema models.Age
        var genderSchema models.Gender
        var socialClassSchema models.SocialClass
        var targetSchema models.Target

        // Função para preencher uma lista com Structs de Idade
        AgeRows, ageSchema = ageFiller(targetStruct, AgeRows)

        // Função para preencher uma lista com Structs de Genêro
        GenderRows, genderSchema = genderFiller(targetStruct, GenderRows)

        // Função para preencher uma lista com Structs de Classe Social
        SociaClassRows, socialClassSchema = socialClassFiller(targetStruct, SociaClassRows)

        // Função para preencher uma lista com Structs de Target
        TargetRows, targetSchema = targetFiller(TargetRows, ageSchema, genderSchema, socialClassSchema)

        // Função para preencher uma lista com Structs de Geodata
        GeodataRows = geodataFiller(GeodataRows, targetSchema, impressionHour, locationID, uniques, latitude, longitude, ufEstado, cidade, endereco, numero)
        fmt.Println("")
        fmt.Println("")

    }

    fmt.Printf("Tamanho de Geodata Rows: %d\n", len(GeodataRows))
    fmt.Printf("Tamanho de Target Rows: %d\n", len(TargetRows))
    fmt.Printf("Tamanho de Age Rows: %d\n", len(AgeRows))
    fmt.Printf("Tamanho de Gender Rows: %d\n", len(GenderRows))
    fmt.Printf("Tamanho de Social Class Rows: %d\n", len(SociaClassRows))

    if err := ageInserter.Put(ctx, AgeRows); err != nil{
        return fmt.Errorf("[service/decode_csv] Erro ao inserir age no BigQuery: %v, err")
    }

    if err := genderInserter.Put(ctx, GenderRows); err != nil{
        return fmt.Errorf("[service/decode_csv] Erro ao inserir age no BigQuery: %v, err")
    }

    if err := socialClassInserter.Put(ctx, SociaClassRows); err != nil{
        return fmt.Errorf("[service/decode_csv] Erro ao inserir age no BigQuery: %v, err")
    }

    if err := targetInserter.Put(ctx, TargetRows); err != nil{
        return fmt.Errorf("[service/decode_csv] Erro ao inserir age no BigQuery: %v, err")
    }

    if err := geodataInserter.Put(ctx, GeodataRows); err != nil{
        return fmt.Errorf("[service/decode_csv] Erro ao inserir age no BigQuery: %v, err")
    }

    fmt.Println("O BigQuery foi populado :D")
    return nil

}

func ageFiller(t Target, a []*models.Age) ([]*models.Age, models.Age) {
        ageStruct := models.Age{}
        for k, v := range t.Idade {
            if _, ok := t.Idade[k]; ok {
                switch k {
                    case "18-19":
                        ageStruct.X18_19 = v
                    
                    case "20-29":
                        ageStruct.X20_29 = v
                    
                    case "30-39":
                        ageStruct.X30_39 = v
                    
                    case "40-49":
                        ageStruct.X40_49 = v
                    
                    case "50-59":
                        ageStruct.X50_59 = v
                    
                    case "60-69":
                        ageStruct.X60_69 = v
                    
                    case "70-79":
                        ageStruct.X70_79 = v
                    
                    case "80+":
                        ageStruct.X80_plus = v
                    
                }
            }
        }
        ageStruct.ID = uuid.New().String()
        a = append(a, &ageStruct)
        return a, ageStruct

}

func genderFiller(t Target, g []*models.Gender) ([]*models.Gender, models.Gender) {
    genderStruct := models.Gender{}
    for k, v := range t.Genero {
        if _, ok := t.Genero[k]; ok {
            switch k {
                case "F":
                    genderStruct.Feminine = v
                case "M":
                    genderStruct.Masculine = v
            }
        }
    }
    genderStruct.ID = uuid.New().String()
    g = append(g, &genderStruct)
    return g, genderStruct
}

func socialClassFiller(t Target, s []*models.SocialClass) ([]*models.SocialClass, models.SocialClass) {
    socialClassStruct := models.SocialClass{}
    for k, v := range t.Classe_Social {
        if _, ok := t.Classe_Social[k]; ok {
            switch k {
                case "A":
                    socialClassStruct.A_Class = v
                case "B1":
                    socialClassStruct.B1_Class = v
                case "B2":
                    socialClassStruct.B2_Class = v
                case "C1":
                    socialClassStruct.C1_Class = v
                case "C2":
                    socialClassStruct.C2_Class = v 
                case "DE":
                    socialClassStruct.DE_Class = v
            }
        }
    }

    socialClassStruct.ID = uuid.New().String()
    s = append(s, &socialClassStruct)
    fmt.Printf("uuid da Classe Social: %s\n", socialClassStruct.ID)
    return s, socialClassStruct
}

func targetFiller(t []*models.Target, age models.Age, gender models.Gender, socialClass models.SocialClass) ([]*models.Target, models.Target) {
    targetStruct := models.Target{
        ID: uuid.New().String(),
        AgeID: age.ID,
        GenderID: gender.ID,
        SocialClassID: socialClass.ID,
    }

    fmt.Printf("ID de social_class definido em Target: %s\n", targetStruct.SocialClassID)
    fmt.Printf("ID do Target %s\n", targetStruct.ID)
    t = append(t, &targetStruct)
    return t, targetStruct

}

func geodataFiller(geo []*models.Geodata, target models.Target, imp string, locID string, uni string, lat string, long string, uf string, cid string, end string, num string) []*models.Geodata {
    locationIDInt, err := strconv.Atoi(locID)
    if err != nil {
        fmt.Printf("erro ao converter location_id: %v", err)
    }

    uniquesFloat, err := strconv.ParseFloat(uni, 64)
    if err != nil {
        fmt.Printf("erro ao converter uniques: %v", err)
    }

    numeroInt, err := strconv.Atoi(num)
    if err != nil {
        fmt.Printf("erro ao converter numero: %v", err)
    }

    geodataStruct := models.Geodata{
        ImpressionHour: imp,
        Uniques:    uniquesFloat,
        LocationID: locationIDInt,
        Latitude: lat,
        Longitude: long,
        UFEstado: uf,
        Cidade: cid,
        Endereco: end,
        Numero: numeroInt,
        TargetID: target.ID,
    }

    geo = append(geo, &geodataStruct)
    fmt.Printf("Segue o Geodata da linha atual: %v\n", geodataStruct)
    return geo
}
