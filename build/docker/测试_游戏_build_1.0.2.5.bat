set name=%~n0
set version=%name:~12%
echo %version%

SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
SET GO111MODULE=on

SET CUR_ROOT_PATH=%CD%

SET TOOLS_PATH=%CUR_ROOT_PATH%\tools

SET CMD_PATH=%CUR_ROOT_PATH%\..\..\cmd\sportsvr

SET PRODUCT_PATH=%CUR_ROOT_PATH%\images\sportsvr

go build -ldflags "-s -w" -o %PRODUCT_PATH%\App\cmd\sportsvr\sportsvr  %CMD_PATH%\main.go

CD %PRODUCT_PATH%
%TOOLS_PATH%\dos2unix.exe .\App\docker-entrypoint.sh
docker build -t sportsvr:%version% .
docker tag sportsvr:%version% crpi-h97z2sfg0o6eqwpi.cn-guangzhou.personal.cr.aliyuncs.com/ggold/sportsvr:%version%
docker push crpi-h97z2sfg0o6eqwpi.cn-guangzhou.personal.cr.aliyuncs.com/ggold/sportsvr:%version%

echo "Done."
pause