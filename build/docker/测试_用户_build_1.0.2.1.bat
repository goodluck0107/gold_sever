set name=%~n0
set version=%name:~12%
echo %version%

SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
SET GO111MODULE=on

SET CUR_ROOT_PATH=%CD%

SET TOOLS_PATH=%CUR_ROOT_PATH%\tools

SET CMD_PATH=%CUR_ROOT_PATH%\..\..\cmd\usersvr

SET PRODUCT_PATH=%CUR_ROOT_PATH%\images\usersvr

go build -ldflags "-s -w" -o %PRODUCT_PATH%\App\cmd\usersvr\usersvr  %CMD_PATH%\main.go

CD %PRODUCT_PATH%
%TOOLS_PATH%\dos2unix.exe .\App\docker-entrypoint.sh
docker build -t usersvr:%version% .
docker tag usersvr:%version% crpi-h97z2sfg0o6eqwpi.cn-guangzhou.personal.cr.aliyuncs.com/ggold/usersvr:%version%
docker push crpi-h97z2sfg0o6eqwpi.cn-guangzhou.personal.cr.aliyuncs.com/ggold/usersvr:%version%

echo "Done."
pause