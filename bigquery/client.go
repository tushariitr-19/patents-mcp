package bigquery

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"github.com/tushariitr-19/patents-mcp/logger"
	"github.com/tushariitr-19/patents-mcp/models"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
)

type Client struct {
	bq        *bigquery.Client
	projectID string
}

func NewClient(ctx context.Context, projectID string) (*Client, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create BigQuery client: %w", err)
	}
	return &Client{bq: client, projectID: projectID}, nil
}

func (c *Client) SearchPatents(ctx context.Context, query string, limit int) ([]models.Patent, error) {
	logger.Log.Info("searching patents", zap.String("query", query), zap.Int("limit", limit))

	sql := fmt.Sprintf(`
    SELECT
        publication_number,
        (SELECT text FROM UNNEST(title_localized) WHERE language = 'en' LIMIT 1) AS title,
        (SELECT text FROM UNNEST(abstract_localized) WHERE language = 'en' LIMIT 1) AS abstract,
        filing_date,
        grant_date,
        inventor_harmonized,
        assignee_harmonized,
        country_code
    FROM
        patents-public-data.patents.publications
    WHERE
        LOWER((SELECT text FROM UNNEST(abstract_localized) WHERE language = 'en' LIMIT 1)) LIKE LOWER(@query)
    LIMIT %d
	`, limit)

	q := c.bq.Query(sql)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "query", Value: "%" + query + "%"},
	}

	it, err := q.Read(ctx)
	if err != nil {
		logger.Log.Error("BigQuery query failed", zap.Error(err))
		return nil, fmt.Errorf("query failed: %w", err)
	}

	var patents []models.Patent
	for {
		var row struct {
			PublicationNumber  string `bigquery:"publication_number"`
			Title              string `bigquery:"title"`
			Abstract           string `bigquery:"abstract"`
			FilingDate         int    `bigquery:"filing_date"`
			GrantDate          int    `bigquery:"grant_date"`
			InvertorHarmonized []struct {
				Name string `bigquery:"name"`
			} `bigquery:"inventor_harmonized"`
			AssigneeHarmonized []struct {
				Name string `bigquery:"name"`
			} `bigquery:"assignee_harmonized"`
			CountryCode string `bigquery:"country_code"`
		}

		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			logger.Log.Error("error reading BigQuery row", zap.Error(err))
			return nil, err
		}

		assignee := ""
		if len(row.AssigneeHarmonized) > 0 {
			assignee = row.AssigneeHarmonized[0].Name
		}

		inventors := make([]string, 0, len(row.InvertorHarmonized))
		for _, inv := range row.InvertorHarmonized {
			inventors = append(inventors, inv.Name)
		}

		patents = append(patents, models.Patent{
			PatentID:   row.PublicationNumber,
			Title:      row.Title,
			Abstract:   row.Abstract,
			FilingDate: fmt.Sprintf("%d", row.FilingDate),
			GrantDate:  fmt.Sprintf("%d", row.GrantDate),
			Inventors:  inventors,
			Assignee:   assignee,
			Country:    row.CountryCode,
			URL:        fmt.Sprintf("https://patents.google.com/patent/%s", row.PublicationNumber),
		})
	}

	logger.Log.Info("search complete", zap.Int("results", len(patents)))
	return patents, nil
}

func (c *Client) Close() error {
	return c.bq.Close()
}

func (c *Client) GetPatent(ctx context.Context, publicationNumber string) (*models.Patent, error) {
	logger.Log.Info("fetching patent", zap.String("publication_number", publicationNumber))

	sql := `
    SELECT
        publication_number,
        (SELECT text FROM UNNEST(title_localized) WHERE language = 'en' LIMIT 1) AS title,
        (SELECT text FROM UNNEST(abstract_localized) WHERE language = 'en' LIMIT 1) AS abstract,
        filing_date,
        grant_date,
        inventor_harmonized,
        assignee_harmonized,
        country_code,
        cpc,
        citation
    FROM
        patents-public-data.patents.publications
    WHERE
        publication_number = @publication_number
    LIMIT 1
	`

	q := c.bq.Query(sql)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "publication_number", Value: publicationNumber},
	}

	it, err := q.Read(ctx)
	if err != nil {
		logger.Log.Error("BigQuery query failed", zap.Error(err))
		return nil, fmt.Errorf("query failed: %w", err)
	}

	var row struct {
		PublicationNumber  string `bigquery:"publication_number"`
		Title              string `bigquery:"title"`
		Abstract           string `bigquery:"abstract"`
		FilingDate         int    `bigquery:"filing_date"`
		GrantDate          int    `bigquery:"grant_date"`
		InvertorHarmonized []struct {
			Name string `bigquery:"name"`
		} `bigquery:"inventor_harmonized"`
		AssigneeHarmonized []struct {
			Name string `bigquery:"name"`
		} `bigquery:"assignee_harmonized"`
		CountryCode string `bigquery:"country_code"`
		CPC         []struct {
			Code string `bigquery:"code"`
		} `bigquery:"cpc"`
		Citation []struct {
			PublicationNumber string `bigquery:"publication_number"`
		} `bigquery:"citation"`
	}

	err = it.Next(&row)
	if err == iterator.Done {
		return nil, fmt.Errorf("patent not found: %s", publicationNumber)
	}
	if err != nil {
		return nil, err
	}

	assignee := ""
	if len(row.AssigneeHarmonized) > 0 {
		assignee = row.AssigneeHarmonized[0].Name
	}

	inventors := make([]string, 0, len(row.InvertorHarmonized))
	for _, inv := range row.InvertorHarmonized {
		inventors = append(inventors, inv.Name)
	}

	cpcCodes := make([]string, 0, len(row.CPC))
	for _, c := range row.CPC {
		cpcCodes = append(cpcCodes, c.Code)
	}

	citations := make([]string, 0, len(row.Citation))
	for _, c := range row.Citation {
		citations = append(citations, c.PublicationNumber)
	}

	return &models.Patent{
		PatentID:   row.PublicationNumber,
		Title:      row.Title,
		Abstract:   row.Abstract,
		FilingDate: fmt.Sprintf("%d", row.FilingDate),
		GrantDate:  fmt.Sprintf("%d", row.GrantDate),
		Inventors:  inventors,
		Assignee:   assignee,
		Country:    row.CountryCode,
		URL:        fmt.Sprintf("https://patents.google.com/patent/%s", row.PublicationNumber),
		CPCCodes:   cpcCodes,
		Citations:  citations,
	}, nil
}
