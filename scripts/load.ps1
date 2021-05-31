$project = "bad_apple"

Get-ChildItem ".\metrics\$project\" | ForEach-Object {
    Write-Host "Loading $_"
    promtool tsdb create-blocks-from openmetrics $_.FullName
}
