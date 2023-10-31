cp $(tinygo env TINYGOROOT)/targets/wasm_exec.js . && tinygo build -o main.wasm -target wasm --no-debug index.go 
file_to_modify="wasm_exec.js"
string_to_replace="console.error('syscall/js.finalizeRef not implemented');"
sed -i "s@$string_to_replace@@" "$file_to_modify"