package iam

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

// AWSIAMService implements IAMService for AWS
type AWSIAMService struct {
	client *iam.Client
	ctx    context.Context
}

// NewAWSIAMService creates a new AWS IAM service using default credentials
func NewAWSIAMService(ctx context.Context) (*AWSIAMService, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &AWSIAMService{
		client: iam.NewFromConfig(cfg),
		ctx:    ctx,
	}, nil
}

// ProvisionUser creates a new IAM user with access keys
func (s *AWSIAMService) ProvisionUser(userName string) (*Identity, error) {
	// Create the IAM user
	createUserOutput, err := s.client.CreateUser(s.ctx, &iam.CreateUserInput{
		UserName: aws.String(userName),
		Tags: []types.Tag{
			{
				Key:   aws.String("Purpose"),
				Value: aws.String("CCC-Testing"),
			},
			{
				Key:   aws.String("ManagedBy"),
				Value: aws.String("CCC-CFI-Compliance-Framework"),
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create IAM user %s: %w", userName, err)
	}

	// Create access key for the user
	createKeyOutput, err := s.client.CreateAccessKey(s.ctx, &iam.CreateAccessKeyInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		// Cleanup: delete the user if key creation fails
		s.client.DeleteUser(s.ctx, &iam.DeleteUserInput{
			UserName: aws.String(userName),
		})
		return nil, fmt.Errorf("failed to create access key for user %s: %w", userName, err)
	}

	// Create identity with credentials in map
	identity := &Identity{
		UserName:    userName,
		Provider:    "aws",
		Credentials: make(map[string]string),
	}

	// Store AWS-specific fields in Credentials map
	identity.Credentials["arn"] = aws.ToString(createUserOutput.User.Arn)
	identity.Credentials["user_id"] = aws.ToString(createUserOutput.User.UserId)
	identity.Credentials["access_key_id"] = aws.ToString(createKeyOutput.AccessKey.AccessKeyId)
	identity.Credentials["secret_access_key"] = aws.ToString(createKeyOutput.AccessKey.SecretAccessKey)

	// Extract and store account ID from ARN (format: arn:aws:iam::123456789012:user/username)
	if createUserOutput.User.Arn != nil {
		arn := aws.ToString(createUserOutput.User.Arn)
		parts := splitARN(arn)
		if len(parts) > 4 {
			identity.Credentials["account_id"] = parts[4]
		}
	}

	return identity, nil
}

// SetAccess grants an identity access to a specific AWS service/resource at the specified level
func (s *AWSIAMService) SetAccess(identity *Identity, serviceID string, level AccessLevel) error {
	// Generate policy document based on access level and service ID
	policyDocument, err := s.generatePolicyDocument(serviceID, level)
	if err != nil {
		return fmt.Errorf("failed to generate policy: %w", err)
	}

	// Create a unique policy name
	policyName := fmt.Sprintf("CCC-Test-%s-%s", sanitizeForPolicyName(serviceID), level)

	// Attach inline policy to user
	_, err = s.client.PutUserPolicy(s.ctx, &iam.PutUserPolicyInput{
		UserName:       aws.String(identity.UserName),
		PolicyName:     aws.String(policyName),
		PolicyDocument: aws.String(policyDocument),
	})
	if err != nil {
		return fmt.Errorf("failed to attach policy to user %s: %w", identity.UserName, err)
	}

	return nil
}

// DestroyUser removes an IAM user and all associated resources
func (s *AWSIAMService) DestroyUser(identity *Identity) error {
	userName := identity.UserName

	// List and delete access keys
	listKeysOutput, err := s.client.ListAccessKeys(s.ctx, &iam.ListAccessKeysInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		return fmt.Errorf("failed to list access keys for user %s: %w", userName, err)
	}

	for _, key := range listKeysOutput.AccessKeyMetadata {
		_, err := s.client.DeleteAccessKey(s.ctx, &iam.DeleteAccessKeyInput{
			UserName:    aws.String(userName),
			AccessKeyId: key.AccessKeyId,
		})
		if err != nil {
			return fmt.Errorf("failed to delete access key %s: %w", aws.ToString(key.AccessKeyId), err)
		}
	}

	// List and delete inline policies
	listPoliciesOutput, err := s.client.ListUserPolicies(s.ctx, &iam.ListUserPoliciesInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		return fmt.Errorf("failed to list user policies for %s: %w", userName, err)
	}

	for _, policyName := range listPoliciesOutput.PolicyNames {
		_, err := s.client.DeleteUserPolicy(s.ctx, &iam.DeleteUserPolicyInput{
			UserName:   aws.String(userName),
			PolicyName: aws.String(policyName),
		})
		if err != nil {
			return fmt.Errorf("failed to delete policy %s: %w", policyName, err)
		}
	}

	// List and detach managed policies
	listAttachedOutput, err := s.client.ListAttachedUserPolicies(s.ctx, &iam.ListAttachedUserPoliciesInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		return fmt.Errorf("failed to list attached policies for %s: %w", userName, err)
	}

	for _, policy := range listAttachedOutput.AttachedPolicies {
		_, err := s.client.DetachUserPolicy(s.ctx, &iam.DetachUserPolicyInput{
			UserName:  aws.String(userName),
			PolicyArn: policy.PolicyArn,
		})
		if err != nil {
			return fmt.Errorf("failed to detach policy %s: %w", aws.ToString(policy.PolicyArn), err)
		}
	}

	// Finally, delete the user
	_, err = s.client.DeleteUser(s.ctx, &iam.DeleteUserInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete user %s: %w", userName, err)
	}

	return nil
}

// generatePolicyDocument creates an IAM policy document for the given resource and access level
func (s *AWSIAMService) generatePolicyDocument(resourceARN string, level AccessLevel) (string, error) {
	var actions []string

	// Determine actions based on service type and access level
	// This is a simplified version - in production, you'd want more sophisticated logic
	switch level {
	case AccessLevelRead:
		actions = []string{
			"s3:GetObject",
			"s3:ListBucket",
			"rds:DescribeDBInstances",
			"ec2:Describe*",
		}
	case AccessLevelWrite:
		actions = []string{
			"s3:GetObject",
			"s3:PutObject",
			"s3:DeleteObject",
			"s3:ListBucket",
			"rds:DescribeDBInstances",
			"rds:ModifyDBInstance",
			"ec2:Describe*",
		}
	case AccessLevelAdmin:
		actions = []string{"*"}
	default:
		return "", fmt.Errorf("unsupported access level: %s", level)
	}

	// Build policy document
	policy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect":   "Allow",
				"Action":   actions,
				"Resource": resourceARN,
			},
		},
	}

	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return "", fmt.Errorf("failed to marshal policy: %w", err)
	}

	return string(policyJSON), nil
}

// Helper functions

func splitARN(arn string) []string {
	// Simple ARN splitter: arn:partition:service:region:account-id:resource
	result := make([]string, 0)
	current := ""
	for _, char := range arn {
		if char == ':' {
			result = append(result, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func sanitizeForPolicyName(s string) string {
	// Replace characters that aren't valid in policy names
	result := ""
	for _, char := range s {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
			result += string(char)
		} else if char == '-' || char == '_' {
			result += string(char)
		}
	}
	if len(result) > 64 {
		result = result[:64]
	}
	return result
}
