$executable = (Get-Item -path $PSScriptRoot).Name

go build

if ($lastexitcode -eq 0) {
 &"./$executable.exe" @args
}
