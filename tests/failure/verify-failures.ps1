param(
  [string]$BaseUrl = "http://localhost:8080",
  [string]$JwtToken = ""
)
$headers = @{ "Content-Type" = "application/json" }
if ($JwtToken) { $headers.Authorization = "Bearer $JwtToken" }

function Assert-Status($label, $script, $expected) {
  $status = $null
  try { & $script | Out-Null } catch { if ($_.Exception.Response) { $status = $_.Exception.Response.StatusCode.value__ } }
  if ($status -ne $expected) { throw "$label returned $status, expected $expected" }
}Assert-Status "missing agent id" { Invoke-RestMethod -Method Post -Uri "$BaseUrl/v1/runs" -Headers $headers -Body '{"input":"bad"}' } 400
Assert-Status "missing run" { Invoke-RestMethod -Method Get -Uri "$BaseUrl/v1/runs/does-not-exist" -Headers $headers } 404
Write-Output "Failure checks passed against $BaseUrl"