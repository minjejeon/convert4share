# Install the generator (ensures it's available, though go run might handle it)
go install github.com/saltosystems/winrt-go/cmd/winrt-go-gen@latest

# Run generation via Go tooling
Write-Host "Running go generate for internal/winrt..."
go generate ./internal/winrt

Write-Host "Done! Bindings generated in internal/winrt"
