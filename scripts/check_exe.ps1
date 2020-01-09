Write-Host "Checking if bin\optimizely.exe  runs..." -ForegroundColor Green
$pid = Start-Process -Filepath "bin\optimizely.exe" -NoNewWindow -PassThru
Start-Sleep -s 5
Write-Host "Stopping bin\optimizely.exe..." -ForegroundColor Green
Stop-Process $pid
