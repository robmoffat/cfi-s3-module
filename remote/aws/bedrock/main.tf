module "bedrock" {
  source  = "aws-ia/bedrock/aws"
  version = "0.0.20"
  foundation_model = "anthropic.claude-v2"
  instruction = "You are an automotive assisant who can provide detailed information about cars to a customer."
}