package main

import (
	"ruleEngine/goLambda"

	//"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	// e := echo.New()
	// e.POST("/test",Handler)
	// e.Logger.Fatal(e.Start(":8000"))
	golambda.Handler()

}