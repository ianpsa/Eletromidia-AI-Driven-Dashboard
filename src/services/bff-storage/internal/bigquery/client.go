package bigquery

import (
	"bff-storage/internal/models"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	bq "cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

const (
	genderThreshold      = 0.5
	ageThreshold         = 0.15
	socialClassThreshold = 0.3

	defaultLimit = 5000
	maxLimit     = 10000

	demographicColumns = `
		AVG(a.x18_19) AS x18_19,
		AVG(a.x20_29) AS x20_29,
		AVG(a.x30_39) AS x30_39,
		AVG(a.x40_49) AS x40_49,
		AVG(a.x50_59) AS x50_59,
		AVG(a.x60_69) AS x60_69,
		AVG(a.x70_79) AS x70_79,
		AVG(a.x80_plus) AS x80_plus,
		AVG(gen.feminine) AS feminine,
		AVG(gen.masculine) AS masculine,
		AVG(sc.a_class) AS a_class,
		AVG(sc.b1_class) AS b1_class,
		AVG(sc.b2_class) AS b2_class,
		AVG(sc.c1_class) AS c1_class,
		AVG(sc.c2_class) AS c2_class,
		AVG(sc.de_class) AS de_class`
)

type demographicRow struct {
	Count   int64   `bigquery:"count"`
	X18_19  float64 `bigquery:"x18_19"`
	X20_29  float64 `bigquery:"x20_29"`
	X30_39  float64 `bigquery:"x30_39"`
	X40_49  float64 `bigquery:"x40_49"`
	X50_59  float64 `bigquery:"x50_59"`
	X60_69  float64 `bigquery:"x60_69"`
	X70_79  float64 `bigquery:"x70_79"`
	X80Plus float64 `bigquery:"x80_plus"`
	Fem     float64 `bigquery:"feminine"`
	Masc    float64 `bigquery:"masculine"`
	AClass  float64 `bigquery:"a_class"`
	B1Class float64 `bigquery:"b1_class"`
	B2Class float64 `bigquery:"b2_class"`
	C1Class float64 `bigquery:"c1_class"`
	C2Class float64 `bigquery:"c2_class"`
	DEClass float64 `bigquery:"de_class"`
}

func rowToSummary(row demographicRow) models.DemographicSummary {
	return models.DemographicSummary{
		Count: int(row.Count),
		Age: models.AgeSummary{
			X18_19:  row.X18_19,
			X20_29:  row.X20_29,
			X30_39:  row.X30_39,
			X40_49:  row.X40_49,
			X50_59:  row.X50_59,
			X60_69:  row.X60_69,
			X70_79:  row.X70_79,
			X80Plus: row.X80Plus,
		},
		Gender: models.GenderSummary{
			Feminine:  row.Fem,
			Masculine: row.Masc,
		},
		SocialClass: models.SocialClassSummary{
			AClass:  row.AClass,
			B1Class: row.B1Class,
			B2Class: row.B2Class,
			C1Class: row.C1Class,
			C2Class: row.C2Class,
			DEClass: row.DEClass,
		},
	}
}

type Client struct {
	ProjectID string
	DatasetID string
	client    *bq.Client
}

func NewClient(ctx context.Context, projectID, datasetID, credentialsFile string) (*Client, error) {
	var opts []option.ClientOption
	if credentialsFile != "" {
		credBytes, err := os.ReadFile(credentialsFile)
		if err != nil {
			return nil, fmt.Errorf("reading BQ credentials: %w", err)
		}
		opts = append(opts, option.WithCredentialsJSON(credBytes))
	}

	client, err := bq.NewClient(ctx, projectID, opts...)
	if err != nil {
		return nil, fmt.Errorf("creating BigQuery client: %w", err)
	}

	return &Client{
		ProjectID: projectID,
		DatasetID: datasetID,
		client:    client,
	}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) CheckDataset(ctx context.Context) error {
	_, err := c.client.Dataset(c.DatasetID).Metadata(ctx)
	return err
}

func (c *Client) table(name string) string {
	return fmt.Sprintf("`%s.%s.%s`", c.ProjectID, c.DatasetID, name)
}

func buildFilters(filters models.GeoFilters) ([]string, []bq.QueryParameter) {
	var conditions []string
	var params []bq.QueryParameter

	if len(filters.UfEstado) == 1 {
		conditions = append(conditions, "g.uf_estado = @uf_estado")
		params = append(params, bq.QueryParameter{Name: "uf_estado", Value: filters.UfEstado[0]})
	} else if len(filters.UfEstado) > 1 {
		conditions = append(conditions, "g.uf_estado IN UNNEST(@uf_estado)")
		params = append(params, bq.QueryParameter{Name: "uf_estado", Value: filters.UfEstado})
	}

	if len(filters.Cidade) == 1 {
		conditions = append(conditions, "g.cidade = @cidade")
		params = append(params, bq.QueryParameter{Name: "cidade", Value: filters.Cidade[0]})
	} else if len(filters.Cidade) > 1 {
		conditions = append(conditions, "g.cidade IN UNNEST(@cidade)")
		params = append(params, bq.QueryParameter{Name: "cidade", Value: filters.Cidade})
	}

	if len(filters.Endereco) == 1 {
		conditions = append(conditions, "g.endereco = @endereco")
		params = append(params, bq.QueryParameter{Name: "endereco", Value: filters.Endereco[0]})
	} else if len(filters.Endereco) > 1 {
		conditions = append(conditions, "g.endereco IN UNNEST(@endereco)")
		params = append(params, bq.QueryParameter{Name: "endereco", Value: filters.Endereco})
	}

	if len(filters.Horario) > 0 {
		hours := parseHours(filters.Horario)
		if len(hours) == 1 {
			conditions = append(conditions, "SAFE_CAST(g.impression_hour AS INT64) = @horario")
			params = append(params, bq.QueryParameter{Name: "horario", Value: hours[0]})
		} else if len(hours) > 1 {
			conditions = append(conditions, "SAFE_CAST(g.impression_hour AS INT64) IN UNNEST(@horario)")
			params = append(params, bq.QueryParameter{Name: "horario", Value: hours})
		}
	}

	if len(filters.Genero) > 0 {
		var genderConds []string
		for _, g := range filters.Genero {
			switch g {
			case "feminino":
				genderConds = append(genderConds, fmt.Sprintf("gen.feminine > %f", genderThreshold))
			case "masculino":
				genderConds = append(genderConds, fmt.Sprintf("gen.masculine > %f", genderThreshold))
			}
		}
		if len(genderConds) > 0 {
			conditions = append(conditions, "("+strings.Join(genderConds, " OR ")+")")
		}
	}

	if len(filters.FaixaEtaria) > 0 {
		var ageConds []string
		for _, faixa := range filters.FaixaEtaria {
			if col := ageColumn(faixa); col != "" {
				ageConds = append(ageConds, fmt.Sprintf("%s > %f", col, ageThreshold))
			}
		}
		if len(ageConds) > 0 {
			conditions = append(conditions, "("+strings.Join(ageConds, " OR ")+")")
		}
	}

	if len(filters.ClasseSocial) > 0 {
		var classConds []string
		for _, cs := range filters.ClasseSocial {
			switch cs {
			case "a":
				classConds = append(classConds, fmt.Sprintf("sc.a_class > %f", socialClassThreshold))
			case "b1":
				classConds = append(classConds, fmt.Sprintf("sc.b1_class > %f", socialClassThreshold))
			case "b2":
				classConds = append(classConds, fmt.Sprintf("sc.b2_class > %f", socialClassThreshold))
			case "c1":
				classConds = append(classConds, fmt.Sprintf("sc.c1_class > %f", socialClassThreshold))
			case "c2":
				classConds = append(classConds, fmt.Sprintf("sc.c2_class > %f", socialClassThreshold))
			case "de":
				classConds = append(classConds, fmt.Sprintf("sc.de_class > %f", socialClassThreshold))
			case "ab":
				classConds = append(classConds, fmt.Sprintf("(sc.a_class + sc.b1_class + sc.b2_class) > %f", socialClassThreshold))
			case "c":
				classConds = append(classConds, fmt.Sprintf("(sc.c1_class + sc.c2_class) > %f", socialClassThreshold))
			}
		}
		if len(classConds) > 0 {
			conditions = append(conditions, "("+strings.Join(classConds, " OR ")+")")
		}
	}

	return conditions, params
}

func whereClause(conditions []string) string {
	if len(conditions) > 0 {
		return "WHERE " + strings.Join(conditions, " AND ")
	}
	return ""
}

func (c *Client) joinTables() string {
	return fmt.Sprintf(`FROM %s g
		JOIN %s t ON g.target_id = t.id
		JOIN %s gen ON t.gender_id = gen.id
		JOIN %s a ON t.age_id = a.id
		JOIN %s sc ON t.social_class_id = sc.id`,
		c.table("geodata"), c.table("target"), c.table("gender"),
		c.table("age"), c.table("social_class"))
}

func (c *Client) QueryGeoPoints(ctx context.Context, filters models.GeoFilters) (*models.GeoPointsResponse, error) {
	limit := filters.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	conditions, params := buildFilters(filters)
	where := whereClause(conditions)

	latNotNull := "SAFE_CAST(g.latitude AS FLOAT64) IS NOT NULL AND SAFE_CAST(g.longitude AS FLOAT64) IS NOT NULL"
	if where == "" {
		where = "WHERE " + latNotNull
	} else {
		where += " AND " + latNotNull
	}

	sql := fmt.Sprintf(`
		SELECT
			g.id,
			SAFE_CAST(g.latitude AS FLOAT64) AS latitude,
			SAFE_CAST(g.longitude AS FLOAT64) AS longitude,
			g.uniques AS value
		%s
		%s
		LIMIT %d`,
		c.joinTables(), where, limit,
	)

	q := c.client.Query(sql)
	q.DefaultProjectID = c.ProjectID
	q.Parameters = params

	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("querying geo points: %w", err)
	}

	var points []models.GeoPoint
	for {
		var row struct {
			ID        string  `bigquery:"id"`
			Latitude  float64 `bigquery:"latitude"`
			Longitude float64 `bigquery:"longitude"`
			Value     float64 `bigquery:"value"`
		}
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading geo point row: %w", err)
		}
		points = append(points, models.GeoPoint{
			ID:        row.ID,
			Latitude:  row.Latitude,
			Longitude: row.Longitude,
			Value:     row.Value,
		})
	}

	return &models.GeoPointsResponse{
		Count:  len(points),
		Points: points,
	}, nil
}

func (c *Client) QueryDemographics(ctx context.Context, filters models.GeoFilters) (*models.DemographicSummary, error) {
	conditions, params := buildFilters(filters)
	where := whereClause(conditions)

	sql := fmt.Sprintf(`
		SELECT
			COUNT(*) AS count,
			%s
		%s
		%s`, demographicColumns, c.joinTables(), where,
	)

	q := c.client.Query(sql)
	q.DefaultProjectID = c.ProjectID
	q.Parameters = params

	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("querying demographics: %w", err)
	}

	var row demographicRow
	if err := it.Next(&row); err != nil {
		return nil, fmt.Errorf("reading demographics row: %w", err)
	}

	result := rowToSummary(row)
	return &result, nil
}

func (c *Client) QueryCompare(ctx context.Context, filters models.GeoFilters, groupBy string) (*models.CompareResponse, error) {
	conditions, params := buildFilters(filters)
	where := whereClause(conditions)

	caseExpr, err := groupByCase(groupBy)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`
		SELECT
			%s AS label,
			COUNT(*) AS count,
			%s
		%s
		%s
		GROUP BY label
		ORDER BY label`,
		caseExpr, demographicColumns, c.joinTables(), where,
	)

	q := c.client.Query(sql)
	q.DefaultProjectID = c.ProjectID
	q.Parameters = params

	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("querying compare: %w", err)
	}

	var groups []models.CompareGroup
	for {
		var row struct {
			Label string `bigquery:"label"`
			demographicRow
		}
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading compare row: %w", err)
		}
		groups = append(groups, models.CompareGroup{
			Label:              row.Label,
			DemographicSummary: rowToSummary(row.demographicRow),
		})
	}

	return &models.CompareResponse{Groups: groups}, nil
}

func groupByCase(groupBy string) (string, error) {
	switch groupBy {
	case "genero":
		return fmt.Sprintf(`CASE
			WHEN gen.feminine > %f THEN 'feminino'
			ELSE 'masculino'
		END`, genderThreshold), nil
	case "faixa_etaria":
		return `CASE
			WHEN a.x18_19 >= GREATEST(a.x20_29, a.x30_39, a.x40_49, a.x50_59, a.x60_69, a.x70_79, a.x80_plus) THEN '18-19'
			WHEN a.x20_29 >= GREATEST(a.x18_19, a.x30_39, a.x40_49, a.x50_59, a.x60_69, a.x70_79, a.x80_plus) THEN '20-29'
			WHEN a.x30_39 >= GREATEST(a.x18_19, a.x20_29, a.x40_49, a.x50_59, a.x60_69, a.x70_79, a.x80_plus) THEN '30-39'
			WHEN a.x40_49 >= GREATEST(a.x18_19, a.x20_29, a.x30_39, a.x50_59, a.x60_69, a.x70_79, a.x80_plus) THEN '40-49'
			WHEN a.x50_59 >= GREATEST(a.x18_19, a.x20_29, a.x30_39, a.x40_49, a.x60_69, a.x70_79, a.x80_plus) THEN '50-59'
			WHEN a.x60_69 >= GREATEST(a.x18_19, a.x20_29, a.x30_39, a.x40_49, a.x50_59, a.x70_79, a.x80_plus) THEN '60-69'
			WHEN a.x70_79 >= GREATEST(a.x18_19, a.x20_29, a.x30_39, a.x40_49, a.x50_59, a.x60_69, a.x80_plus) THEN '70-79'
			ELSE '80+'
		END`, nil
	case "classe_social":
		return `CASE
			WHEN sc.a_class >= GREATEST(sc.b1_class, sc.b2_class, sc.c1_class, sc.c2_class, sc.de_class) THEN 'a'
			WHEN sc.b1_class >= GREATEST(sc.a_class, sc.b2_class, sc.c1_class, sc.c2_class, sc.de_class) THEN 'b1'
			WHEN sc.b2_class >= GREATEST(sc.a_class, sc.b1_class, sc.c1_class, sc.c2_class, sc.de_class) THEN 'b2'
			WHEN sc.c1_class >= GREATEST(sc.a_class, sc.b1_class, sc.b2_class, sc.c2_class, sc.de_class) THEN 'c1'
			WHEN sc.c2_class >= GREATEST(sc.a_class, sc.b1_class, sc.b2_class, sc.c1_class, sc.de_class) THEN 'c2'
			ELSE 'de'
		END`, nil
	default:
		return "", fmt.Errorf("invalid group_by value: %s", groupBy)
	}
}

func (c *Client) QueryFilterOptions(ctx context.Context, filters models.GeoFilters) (*models.FilterOptions, error) {
	result := &models.FilterOptions{}

	estados, err := c.queryDistinct(ctx, "uf_estado", "", "", nil)
	if err != nil {
		return nil, fmt.Errorf("querying estados: %w", err)
	}
	result.Estados = estados

	var cidadeCondition string
	var cidadeParams []bq.QueryParameter
	if len(filters.UfEstado) == 1 {
		cidadeCondition = "uf_estado = @uf_estado"
		cidadeParams = []bq.QueryParameter{{Name: "uf_estado", Value: filters.UfEstado[0]}}
	} else if len(filters.UfEstado) > 1 {
		cidadeCondition = "uf_estado IN UNNEST(@uf_estado)"
		cidadeParams = []bq.QueryParameter{{Name: "uf_estado", Value: filters.UfEstado}}
	}
	cidades, err := c.queryDistinct(ctx, "cidade", cidadeCondition, "", cidadeParams)
	if err != nil {
		return nil, fmt.Errorf("querying cidades: %w", err)
	}
	result.Cidades = cidades

	var enderecoCondition string
	var enderecoParams []bq.QueryParameter
	if len(filters.Cidade) == 1 {
		enderecoCondition = "cidade = @cidade"
		enderecoParams = []bq.QueryParameter{{Name: "cidade", Value: filters.Cidade[0]}}
	} else if len(filters.Cidade) > 1 {
		enderecoCondition = "cidade IN UNNEST(@cidade)"
		enderecoParams = []bq.QueryParameter{{Name: "cidade", Value: filters.Cidade}}
	}
	enderecos, err := c.queryDistinct(ctx, "endereco", enderecoCondition, "", enderecoParams)
	if err != nil {
		return nil, fmt.Errorf("querying enderecos: %w", err)
	}
	result.Enderecos = enderecos

	result.Horarios = []string{
		"00", "01", "02", "03", "04", "05",
		"06", "07", "08", "09", "10", "11",
		"12", "13", "14", "15", "16", "17",
		"18", "19", "20", "21", "22", "23",
	}

	ageCols, err := c.queryTableColumns(ctx, "age")
	if err != nil {
		return nil, fmt.Errorf("querying age columns: %w", err)
	}
	for _, col := range ageCols {
		result.FaixasEtarias = append(result.FaixasEtarias, ageColumnToLabel(col))
	}

	genderCols, err := c.queryTableColumns(ctx, "gender")
	if err != nil {
		return nil, fmt.Errorf("querying gender columns: %w", err)
	}
	for _, col := range genderCols {
		result.Generos = append(result.Generos, genderColumnToLabel(col))
	}

	socialClassCols, err := c.queryTableColumns(ctx, "social_class")
	if err != nil {
		return nil, fmt.Errorf("querying social_class columns: %w", err)
	}
	for _, col := range socialClassCols {
		result.ClassesSociais = append(result.ClassesSociais, socialClassColumnToLabel(col))
	}

	return result, nil
}

func (c *Client) queryDistinct(ctx context.Context, column, condition, _ string, params []bq.QueryParameter) ([]string, error) {
	where := ""
	if condition != "" {
		where = "WHERE " + condition
	}

	sql := fmt.Sprintf(
		"SELECT DISTINCT %s FROM %s %s ORDER BY %s",
		column, c.table("geodata"), where, column,
	)

	q := c.client.Query(sql)
	q.DefaultProjectID = c.ProjectID
	q.Parameters = params

	it, err := q.Read(ctx)
	if err != nil {
		return nil, err
	}

	var values []string
	for {
		var row []bq.Value
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(row) > 0 && row[0] != nil {
			values = append(values, fmt.Sprintf("%v", row[0]))
		}
	}

	return values, nil
}

func parseHours(raw []string) []int64 {
	var hours []int64
	for _, h := range raw {
		if v, err := strconv.ParseInt(h, 10, 64); err == nil && v >= 0 && v <= 23 {
			hours = append(hours, v)
		}
	}
	return hours
}

func ageColumn(faixa string) string {
	switch faixa {
	case "18-19":
		return "a.x18_19"
	case "20-29":
		return "a.x20_29"
	case "30-39":
		return "a.x30_39"
	case "40-49":
		return "a.x40_49"
	case "50-59":
		return "a.x50_59"
	case "60-69":
		return "a.x60_69"
	case "70-79":
		return "a.x70_79"
	case "80+":
		return "a.x80_plus"
	case "50+":
		return "(a.x50_59 + a.x60_69 + a.x70_79 + a.x80_plus)"
	default:
		return ""
	}
}

func (c *Client) queryTableColumns(ctx context.Context, tableName string) ([]string, error) {
	sql := fmt.Sprintf(
		"SELECT column_name FROM `%s.%s.INFORMATION_SCHEMA.COLUMNS` WHERE table_name = @table AND column_name != 'id' ORDER BY ordinal_position",
		c.ProjectID, c.DatasetID,
	)

	q := c.client.Query(sql)
	q.DefaultProjectID = c.ProjectID
	q.Parameters = []bq.QueryParameter{{Name: "table", Value: tableName}}

	it, err := q.Read(ctx)
	if err != nil {
		return nil, err
	}

	var columns []string
	for {
		var row []bq.Value
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(row) > 0 && row[0] != nil {
			columns = append(columns, fmt.Sprintf("%v", row[0]))
		}
	}

	return columns, nil
}

func ageColumnToLabel(col string) string {
	s := strings.TrimPrefix(col, "x")
	s = strings.Replace(s, "_plus", "+", 1)
	s = strings.Replace(s, "_", "-", 1)
	return s
}

func genderColumnToLabel(col string) string {
	switch col {
	case "feminine":
		return "feminino"
	case "masculine":
		return "masculino"
	default:
		return col
	}
}

func socialClassColumnToLabel(col string) string {
	return strings.TrimSuffix(col, "_class")
}
