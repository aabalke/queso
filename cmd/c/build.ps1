$env:CGO_ENABLED=1
go build -buildmode=c-archive -o queso.a c.go

# will need libqueso.h and libqueso.h with required functions
# gcc main.c queso.a  -o main;

# ./build.ps1; gcc main.c libqueso.h queso.a -o main.exe; ./main.exe
