package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lambda"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Create an AWS resource (S3 Bucket)
		bucket, err := s3.NewBucket(ctx, "my-bucket", nil)
		if err != nil {
			return err
		}

		// Export the name of the bucket
		ctx.Export("bucketName", bucket.ID())

		// Create an AWS resource (Lambda Function)
		_, err = lambda.NewFunction(ctx, "my-function", &lambda.FunctionArgs{
			Code:    pulumi.NewFileArchive("./lambda.zip"),
			Handler: pulumi.String("lambda.handler"),
			Runtime: pulumi.String("nodejs14.x"),
		})
		if err != nil {
			return err
		}
		return nil
	})
}
