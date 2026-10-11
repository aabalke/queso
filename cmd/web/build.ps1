cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" .

$env:GOOS="js"
$env:GOARCH="wasm"

go build -o main.wasm

#py -m http.server 8080
