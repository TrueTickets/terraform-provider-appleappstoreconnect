Your job is to create a Terraform provider for Apple's App Store Connect API (https://developer.apple.com/documentation/appstoreconnectapi).
Initially the provider will implement resources and datasources only for Pass Type IDs and Certificates.

You can do the initial scaffolding of the provider using Hashicorp's quick start repository: https://github.com/hashicorp/terraform-provider-scaffolding-framework

The code must adhere to a few rules:
- Think step-by-step before writing the code, approach it logically
- Must be written in Go using the Terraform Plugin Framework:
	- https://github.com/hashicorp/terraform-plugin-framework
	- https://developer.hashicorp.com/terraform/plugin/framework
- Avoid hardcoded values like project IDs
- Code written should be as parallel as possible enabling the fastest and most optimal execution
- Code should handle errors gracefully, especially when doing multiple API calls
- Each error should be handled and logged with a reason
- Code should have a comprehensive unit test suite
- Examples of Terraform code for use of the provider and the different resources and datasources should also be created
- Code should be extensively commented
- Clear and comprehensive docs in the expected markdown format and location have to be generated so the provided can be published in public registries
- Any configuration parameters the provider needs (e.g. API key) have to be accepted both as parameters in the provider definition and as env vars
- Resources need to support import using the API's unique identifier for that resource
- Generate the corresponding context files so that new executions of AI agents can understand the project and be more productive. Also generate any custom tools or commands that you think would be optimal for you or other AI agents

Be concise, professional and to the point. Do not give generic advice, and do not make assumptions, always ask for any additional information you need.

Start with making a detailed plan for the steps you will take to create the provider. Then prompt me before starting each step so we can either proceed or adjust course.
