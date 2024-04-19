package mailsenderservice

import (
	"context"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/nguyentrunghieu15/vcs-common-prj/apu/mail_sender"
)

type ElasticService struct {
	client *elasticsearch.TypedClient
	index  string
}

func NewElasticService(config elasticsearch.Config) *ElasticService {
	typedClient, err := elasticsearch.NewTypedClient(config)
	if err != nil {
		log.Fatalln("Can't create elastic client", err)
	}
	return &ElasticService{client: typedClient, index: "server_statistic"}
}

func (e *ElasticService) GetAverageUpTimeServer(req *mail_sender.RequestSendStatisticServerToEmail) (*search.Response, error) {
	termRequest := "ID.keyword"
	sumField := "In"
	result, err := e.client.Search().Index(e.index).
		Request(&search.Request{
			Query: &types.Query{
				Bool: &types.BoolQuery{
					Must: []types.Query{
						types.Query{
							Match: map[string]types.MatchQuery{
								"Status": types.MatchQuery{
									Query: "on",
								},
							},
						},
					},
					Filter: []types.Query{
						types.Query{
							Range: map[string]types.RangeQuery{
								"At": types.DateRangeQuery{
									Gte: &req.From,
									Lte: &req.To,
								},
							},
						},
					},
				},
			},
			Aggregations: map[string]types.Aggregations{
				"statis": types.Aggregations{
					Terms: &types.TermsAggregation{
						Field: &termRequest,
					},
					Aggregations: map[string]types.Aggregations{
						"sum_time_up": types.Aggregations{
							Sum: &types.SumAggregation{
								Field: &sumField,
							},
						},
					},
				},
				"avg_time_up": {
					AvgBucket: &types.AverageBucketAggregation{
						BucketsPath: "statis>sum_time_up",
					},
				},
			},
		}).Do(context.TODO())
	return result, err
}
