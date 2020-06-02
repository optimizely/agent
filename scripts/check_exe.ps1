Write-Host "Checking if bin\optimizely.exe works..." -ForegroundColor Green

$OPTLY_PID = Start-Process -Filepath "bin\optimizely.exe" -NoNewWindow -PassThru
Start-Sleep -s 5

if( Invoke-RestMethod -uri 127.0.0.1:8080/health -Method Get | Where-Object { $_.status -match "ok" } ){
  Write-Host "Success: /health endpoint is working"
}else{
  Write-Host "Stopping bin\optimizely.exe..." -ForegroundColor Green
  Stop-Process $OPTLY_PID
  throw "Failure: /health endpoint is not working"
}

Write-Host "Stopping bin\optimizely.exe..." -ForegroundColor Green
Stop-Process $OPTLY_PID
