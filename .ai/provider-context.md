# AI Assistant Quick Reference - Apple App Store Connect Terraform Provider

## Quick Commands

```bash
# Build
go build -o terraform-provider-appleappstoreconnect

# Test
go test ./...                          # Unit tests
TF_ACC=1 go test ./... -timeout 30m   # Acceptance tests

# Quality
make fmt        # Format code
make lint       # Run linter
make generate   # Generate docs

# Install locally
go install
```

## Project Structure

```
.
├── internal/provider/
│   ├── client.go                    # JWT API client
│   ├── provider.go                  # Main provider
│   ├── pass_type_id_*.go          # Pass Type ID resource/datasource
│   ├── certificate_*.go           # Certificate resource/datasource
│   └── certificates_*.go          # Multiple certificates datasource
├── examples/                       # Usage examples
├── docs/                          # Generated documentation
├── templates/                     # Documentation templates
└── tools/                         # Build tools
```

## Key Patterns

### Resource Implementation

```go
type MyResource struct {
    client *Client
}

func (r *MyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse)
func (r *MyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse)
func (r *MyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse)
func (r *MyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse)
func (r *MyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse)
```

### API Client Usage

```go
resp, err := r.client.Do(ctx, Request{
    Method:   http.MethodPost,
    Endpoint: "/passTypeIds",
    Body:     requestBody,
})
```

### Type Conversion

```go
// Terraform types to Go
identifier := data.Identifier.ValueString()

// Go to Terraform types
data.ID = types.StringValue(apiResponse.Data.ID)
```

## Common Tasks for AI Assistants

### Adding a New Field to a Resource

1. Add to model struct with `tfsdk` tag
2. Add to schema with description
3. Update CRUD operations to handle field
4. Add validation if needed
5. Update tests
6. Regenerate documentation

### Debugging API Issues

1. Check `tflog.Debug()` statements
2. Verify JWT token generation
3. Check API response structure
4. Validate request body format

### Writing Tests

- Use `resource.Test` for acceptance tests
- Mock API responses for unit tests
- Always clean up test resources
- Use meaningful test names

## API Endpoints

- `/v1/passTypeIds` - Pass Type IDs
- `/v1/certificates` - Certificates
- Relationships via included data

## Validation Rules

- Pass Type ID: `pass.` prefix, reverse DNS format
- Certificate Type: Must be from allowed list
- CSR Content: Valid PEM format
- Relationships: Pass certs need Pass Type ID

## Error Handling

```go
if err != nil {
    resp.Diagnostics.AddError(
        "Client Error",
        fmt.Sprintf("Unable to create Pass Type ID, got error: %s", err),
    )
    return
}
```

## Important Notes

- Always run linter after changes
- JWT tokens expire in 20 minutes
- API returns max 200 items per request
- All times in RFC3339 format
- Sensitive fields marked in schema
