Example usage of the mod_management API

```go
client, err := modmanagement.NewClient(modmanagement.ModManagementClientOpts{})
if err != nil {
    log.Fatal(err)
}
```

```go
client, err := modmanagement.NewClient(modmanagement.ModManagementClientOpts{
    Port: lo.ToPtr(8080),
})
if err != nil {
    log.Fatal(err)
}
```
