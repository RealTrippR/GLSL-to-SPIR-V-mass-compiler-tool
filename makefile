SRC = glsl_spirv.go

OUT_WIN = glsl_spirv.exe
OUT_LIN = glsl_spirv.o

build_windows_amd64:
	set GOOS=windows&& set GOARCH=amd64&& go build -o $(OUT_WIN) $(SRC)

build_linux_amd64:
	set GOOS=linux&& set GOARCH=amd64&& go build -o $(OUT_LIN) $(SRC)
