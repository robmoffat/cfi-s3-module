module "bedrock" {
  source  = "git::https://github.com/aws-ia/terraform-aws-bedrock.git//examples/agent-with-guardrails?ref=0.0.29"
}