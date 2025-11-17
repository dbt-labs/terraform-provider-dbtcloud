package validators

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type S3PathValidator struct{}

func (v S3PathValidator) Description(ctx context.Context) string {
	return "Validates that the S3 path starts with 's3://' and follows the correct S3 URI format."
}

func (v S3PathValidator) MarkdownDescription(ctx context.Context) string {
	return "Validates that the S3 path starts with `s3://` and follows the correct S3 URI format."
}

func (v S3PathValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Skip validation if the value is unknown or null
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	// Get the value
	s3Path := req.ConfigValue.ValueString()

	// Check if the path starts with s3://
	if !strings.HasPrefix(s3Path, "s3://") {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid S3 Path Format",
			fmt.Sprintf("The S3 path must start with 's3://'. Got: %s\nExample: s3://my-bucket/my-path/", s3Path),
		)
		return
	}

	// Check if there's content after s3://
	bucketAndPath := strings.TrimPrefix(s3Path, "s3://")
	if bucketAndPath == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid S3 Path Format",
			fmt.Sprintf("The S3 path must include a bucket name after 's3://'. Got: %s\nExample: s3://my-bucket/my-path/", s3Path),
		)
		return
	}

	// Check if path ends with a slash (best practice for S3 directory paths)
	if !strings.HasSuffix(s3Path, "/") {
		resp.Diagnostics.AddAttributeWarning(
			req.Path,
			"S3 Path Format Recommendation",
			fmt.Sprintf("The S3 path should typically end with a trailing slash for directory paths. Got: %s\nRecommended: %s/", s3Path, s3Path),
		)
	}
}
