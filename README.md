# Azure API VM Size Lister

This is a small utility that extracts VM Sizes from all azure regions and stores the data in a json file iand pushes the resulting files to Azure Blob Storage.

### Environment Variables
This application expects the following environment variables.

```bash
"AZURE_TENANT_ID": "<-- Tenant Id -->",
"AZURE_CLIENT_ID": "<-- Service Principal Client Id -->",
"AZURE_CLIENT_SECRET": "<-- Service Principal Secret -->",
"AZURE_SUBSCRIPTION_ID": "<-- Subscrpition Id -->",
"AZURE_STORAGE_ACCOUNT" : "<-- Sorage Account Name -->",
"AZURE_STORAGE_ACCESS_KEY" : "<-- Storage Key -->"
```

## Azure Resources
This application requires the creation of an Azure Storage account in order to store the final output files.

This will be used to support an azure rule in [https://github.com/terraform-linters/tflint-ruleset-azurerm](https://github.com/terraform-linters/tflint-ruleset-azurerm)