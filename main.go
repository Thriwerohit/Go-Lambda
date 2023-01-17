package main

import (
	golambda "ruleEngine/goLambda"

	//"github.com/aws/aws-lambda-go/lambda"
)

func main() {
//lambda.Start(golambda.Handler)
golambda.Handler()
}
