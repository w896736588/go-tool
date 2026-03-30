[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

$backendJobName = 'dtool-wails-backend-dev'
$jobName = 'dtool-wails-frontend-dev'
$workspaceRoot = 'C:\work\frog\dev_tool_master'
$frontendDir = Join-Path $workspaceRoot 'web'
$frontendDevServerUrl = 'http://localhost:8080'

$existingBackendJob = Get-Job -Name $backendJobName -ErrorAction SilentlyContinue
if ($existingBackendJob) {
  Remove-Job -Name $backendJobName -Force
}

$existingJob = Get-Job -Name $jobName -ErrorAction SilentlyContinue
if ($existingJob) {
  Remove-Job -Name $jobName -Force
}

Start-Job -Name $backendJobName -ScriptBlock {
  param($workspaceRootPath)
  Set-Location $workspaceRootPath
  go run ./cmd/dtool/main.go --ConfigFile=company
} -ArgumentList $workspaceRoot | Out-Null

for ($i = 0; $i -lt 60; $i++) {
  if ((Test-NetConnection -ComputerName 127.0.0.1 -Port 17170 -WarningAction SilentlyContinue).TcpTestSucceeded) {
    break
  }
  Start-Sleep -Milliseconds 500
}

Start-Job -Name $jobName -ScriptBlock {
  param($frontendDirPath)
  Set-Location $frontendDirPath
  npm run dev
} -ArgumentList $frontendDir | Out-Null

Start-Sleep -Seconds 5

$env:FRONTEND_DEVSERVER_URL = $frontendDevServerUrl
$env:DTOOL_WAILS_DEV_EXTERNAL_BACKEND = '1'
Set-Location $workspaceRoot
go run ./cmd/dtool_wails --ConfigFile=company
