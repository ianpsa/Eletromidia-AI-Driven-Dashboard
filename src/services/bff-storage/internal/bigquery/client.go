package bigquery

import (
	"bff-storage/internal/models"
	"context"
	"fmt"
	"os"
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
)

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

func (c *Client) QueryGeoPoints(ctx context.Context, filters models.GeoFilters) (*models.GeoPointsResponse, error) {
	limit := filters.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	var conditions []string
	var params []bq.QueryParameter

	if filters.UfEstado != "" {
		conditions = append(conditions, "g.uf_estado = @uf_estado")
		params = append(params, bq.QueryParameter{Name: "uf_estado", Value: filters.UfEstado})
	}
	if filters.Cidade != "" {
		conditions = append(conditions, "g.cidade = @cidade")
		params = append(params, bq.QueryParameter{Name: "cidade", Value: filters.Cidade})
	}
	if filters.Endereco != "" {
		conditions = append(conditions, "g.endereco = @endereco")
		params = append(params, bq.QueryParameter{Name: "endereco", Value: filters.Endereco})
	}

	switch filters.Genero {
	case "feminino":
		conditions = append(conditions, fmt.Sprintf("gen.feminine > %f", genderThreshold))
	case "masculino":
		conditions = append(conditions, fmt.Sprintf("gen.masculine > %f", genderThreshold))
	}

	if col := ageColumn(filters.FaixaEtaria); col != "" {
		conditions = append(conditions, fmt.Sprintf("%s > %f", col, ageThreshold))
	}

	switch filters.ClasseSocial {
	case "ab":
		conditions = append(conditions, fmt.Sprintf("(sc.a_class + sc.b1_class + sc.b2_class) > %f", socialClassThreshold))
	case "c":
		conditions = append(conditions, fmt.Sprintf("(sc.c1_class + sc.c2_class) > %f", socialClassThreshold))
	case "de":
		conditions = append(conditions, fmt.Sprintf("sc.de_class > %f", socialClassThreshold))
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	sql := fmt.Sprintf(`
		SELECT
			g.id,
			SAFE_CAST(g.latitude AS FLOAT64) AS latitude,
			SAFE_CAST(g.longitude AS FLOAT64) AS longitude,
			g.uniques AS value
		FROM %s g
		JOIN %s t ON g.target_id = t.id
		JOIN %s gen ON t.gender_id = gen.id
		JOIN %s a ON t.age_id = a.id
		JOIN %s sc ON t.social_class_id = sc.id
		%s
		AND SAFE_CAST(g.latitude AS FLOAT64) IS NOT NULL
		AND SAFE_CAST(g.longitude AS FLOAT64) IS NOT NULL
		LIMIT %d`,
		c.table("geodata"), c.table("target"), c.table("gender"),
		c.table("age"), c.table("social_class"),
		where, limit,
	)

	if where == "" {
		sql = strings.Replace(sql, "AND SAFE_CAST(g.latitude", "WHERE SAFE_CAST(g.latitude", 1)
	}

	params = append(params, bq.QueryParameter{Name: "limit", Value: int64(limit)})

	q := c.client.Query(sql)
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
	var conditions []string
	var params []bq.QueryParameter

	if filters.UfEstado != "" {
		conditions = append(conditions, "g.uf_estado = @uf_estado")
		params = append(params, bq.QueryParameter{Name: "uf_estado", Value: filters.UfEstado})
	}
	if filters.Cidade != "" {
		conditions = append(conditions, "g.cidade = @cidade")
		params = append(params, bq.QueryParameter{Name: "cidade", Value: filters.Cidade})
	}
	if filters.Endereco != "" {
		conditions = append(conditions, "g.endereco = @endereco")
		params = append(params, bq.QueryParameter{Name: "endereco", Value: filters.Endereco})
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	sql := fmt.Sprintf(`
		SELECT
			COUNT(*) AS count,
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
			AVG(sc.de_class) AS de_class
		FROM %s g
		JOIN %s t ON g.target_id = t.id
		JOIN %s gen ON t.gender_id = gen.id
		JOIN %s a ON t.age_id = a.id
		JOIN %s sc ON t.social_class_id = sc.id
		%s`,
		c.table("geodata"), c.table("target"), c.table("gender"),
		c.table("age"), c.table("social_class"),
		where,
	)

	q := c.client.Query(sql)
	q.Parameters = params

	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("querying demographics: %w", err)
	}

	var row struct {
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

	if err := it.Next(&row); err != nil {
		return nil, fmt.Errorf("reading demographics row: %w", err)
	}

	return &models.DemographicSummary{
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
	}, nil
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
	if filters.UfEstado != "" {
		cidadeCondition = "uf_estado = @uf_estado"
		cidadeParams = []bq.QueryParameter{{Name: "uf_estado", Value: filters.UfEstado}}
	}
	cidades, err := c.queryDistinct(ctx, "cidade", cidadeCondition, "", cidadeParams)
	if err != nil {
		return nil, fmt.Errorf("querying cidades: %w", err)
	}
	result.Cidades = cidades

	var enderecoCondition string
	var enderecoParams []bq.QueryParameter
	if filters.Cidade != "" {
		enderecoCondition = "cidade = @cidade"
		enderecoParams = []bq.QueryParameter{{Name: "cidade", Value: filters.Cidade}}
	}
	enderecos, err := c.queryDistinct(ctx, "endereco", enderecoCondition, "", enderecoParams)
	if err != nil {
		return nil, fmt.Errorf("querying enderecos: %w", err)
	}
	result.Enderecos = enderecos

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
	case "50+":
		return "(a.x50_59 + a.x60_69 + a.x70_79 + a.x80_plus)"
	default:
		return ""
	}
}
