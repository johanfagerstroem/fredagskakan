package main

import (
	"fmt"
//	"strconv"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type kakItem struct {
	Year int
	Week int
	Name string
	Url string
}

type scoreItem struct {
	YearWeekLocation string
	Year int
	Week int
	Location string
	ScoreSum int
	VoteCount int
}

type scoreResponse struct {
	Location string
	Score int
}

type kakResponse struct {
	Kaka string
	Year int
	Week int
	Scores []scoreResponse
}

func getScores(svc *dynamodb.DynamoDB, year, week int) []scoreResponse {
	filtYear := expression.Name("Year").Equal(expression.Value(year))
	filtWeek := expression.Name("Week").Equal(expression.Value(week))

	expr, err := expression.NewBuilder().WithFilter(filtYear.And(filtWeek)).Build()
	if err != nil {
		panic(fmt.Sprintf("Failed to build: %v", err))
	}

	params := &dynamodb.ScanInput{
		TableName: aws.String("votes"),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
	}

	result, err := svc.Scan(params)
	if err != nil {
		panic(fmt.Sprintf("Failed to scan: %v", err))
	}

	var scores []scoreItem
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &scores)

	var res []scoreResponse
	totalScoreSum := 0
	totalVoteCount := 0
	for _, s := range scores {
		res = append(res, scoreResponse{s.Location, s.ScoreSum/s.VoteCount})
		totalScoreSum = totalScoreSum + s.ScoreSum
		totalVoteCount = totalVoteCount + s.VoteCount
	}
	if totalVoteCount > 0 {
		avg := totalScoreSum / totalVoteCount
		res = append(res, scoreResponse{"all", avg})
	}

	return res
}

func requestHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	sess := session.Must(session.NewSession())
	svc  := dynamodb.New(sess)

	year := request.PathParameters["year"]
	week := request.PathParameters["week"]

	if year == "" {
		return events.APIGatewayProxyResponse{StatusCode: 400}, nil
	}

	var params *dynamodb.QueryInput
	if week == "" {
		params = &dynamodb.QueryInput{
			TableName: aws.String("kakor"),
			KeyConditionExpression: aws.String("#yr = :yyyy"),
			ExpressionAttributeNames: map[string]*string{
				"#yr": aws.String("Year"),
			},
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":yyyy": {
					N: aws.String(year),
				},
			},
		}
	} else {
		params = &dynamodb.QueryInput{
			TableName: aws.String("kakor"),
			KeyConditionExpression: aws.String("#yr = :yyyy and #wk = :week"),
			ExpressionAttributeNames: map[string]*string{
				"#yr": aws.String("Year"),
				"#wk": aws.String("Week"),
			},
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":yyyy": {
					N: aws.String(year),
				},
				":week": {
					N: aws.String(week),
				},
			},
		}
	}

	result, err := svc.Query(params)
	if err != nil {
		panic(fmt.Sprintf("Failed to build: %v", err))
	}

	var movies []kakItem
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &movies)

	if len(movies) == 1 {
		scores := getScores(svc, movies[0].Year, movies[0].Week)
		s, _ := json.Marshal(kakResponse{movies[0].Name, movies[0].Year, movies[0].Week, scores})
		return events.APIGatewayProxyResponse{Body: string(s), StatusCode: 200}, nil
	} else {
		var results []kakResponse
		for _, m := range movies {
			scores := getScores(svc, m.Year, m.Week)
			results = append(results, kakResponse{m.Name, m.Year, m.Week, scores})
		}

		s, _ := json.Marshal(results)
		return events.APIGatewayProxyResponse{Body: string(s), StatusCode: 200}, nil
	}


}


func main() {
	lambda.Start(requestHandler)
}
