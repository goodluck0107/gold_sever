set name=%~n0
set version=%name:~12%
echo %version%

SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
SET GO111MODULE=on

SET CUR_ROOT_PATH=%CD%

SET TOOLS_PATH=%CUR_ROOT_PATH%\tools

SET CMD_PATH=%CUR_ROOT_PATH%\..\..\cmd\apisvr

SET PRODUCT_PATH=%CUR_ROOT_PATH%\images\apisvr

go build -ldflags "-s -w" -o %PRODUCT_PATH%\App\cmd\apisvr\apisvr  %CMD_PATH%\main.go

CD %PRODUCT_PATH%
%TOOLS_PATH%\dos2unix.exe .\App\docker-entrypoint.sh
docker build -t apisvr:%version% .
docker tag apisvr:%version% crpi-h97z2sfg0o6eqwpi.cn-guangzhou.personal.cr.aliyuncs.com/ggold/apisvr:%version%
docker push crpi-h97z2sfg0o6eqwpi.cn-guangzhou.personal.cr.aliyuncs.com/ggold/apisvr:%version%

echo "Done."
pause