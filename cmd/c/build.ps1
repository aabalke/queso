$env:CGO_ENABLED=1

go build -buildmode=c-archive -o queso.a c.go

# will need libqueso.h and libqueso.h with required functions
#
# gcc main.c queso.a  -o main;
