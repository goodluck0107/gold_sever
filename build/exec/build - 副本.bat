@echo off

rem 此文件用于在WINDOWS平台下编译程序

echo (1/9): [INIT FIRST]

rd /s/q bin && echo [SUCCEED] || echo remove bin error, this is a warn... ignore it....

SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64

echo.

echo (2/9): [BUILD backendsvr] ...

go build -ldflags "-w -s" -o ./bin/cmd/backendsvr/backendsvr ../../cmd/backendsvr/main.go && echo [SUCCEED] || goto F

echo.

echo (3/9): [BUILD apisvr] ...

go build -ldflags "-w -s" -o ./bin/cmd/apisvr/apisvr ../../cmd/apisvr/main.go && echo [SUCCEED] || goto F

echo.

echo (4/9): [BUILD usersvr] ...

go build -ldflags "-w -s" -o ./bin/cmd/usersvr/usersvr ../../cmd/usersvr/main.go && echo [SUCCEED] || goto F

echo.

echo (5/9): [BUILD centersvr] ...

go build -ldflags "-w -s" -o ./bin/cmd/centersvr/centersvr ../../cmd/centersvr/main.go && echo [SUCCEED] || goto F

echo.

echo (6/9): [BUILD sportsvr] ...

go build -ldflags "-w -s" -o ./bin/cmd/sportsvr/sportsvr ../../cmd/sportsvr/main.go && echo [SUCCEED] || goto F

echo.

echo (7/9): [COPY ETC FILES]

xcopy table\\fanRule bin\\cmd\\sportsvr\\fanRule /e/i && echo [SUCCEED 1/4] || goto F
xcopy table\\tbl bin\\cmd\\sportsvr\\tbl /e/i && echo [SUCCEED 2/4] || goto F
xcopy table\\tbl_ini bin\\cmd\\sportsvr\\tbl_ini /e/i && echo [SUCCEED 3/4] || goto F

xcopy ..\\..\\configs bin\\configs /e/i && echo [SUCCEED 4/4] || goto F

echo.

echo (8/9): [ZIP MATERIALS]

set filename=%date:~0,4%%date:~5,2%%date:~8,2%%time:~0,2%%time:~3,2%%time:~6,2%
set "filename=%filename: =0%"

..\\..\\configs\\7z a bin_%filename%.7z .//bin//* -r -mx=9 && echo [SUCCEED] || goto F

echo (9/9): [FREE FILES]
rd /s/q bin && echo [SUCCEED]
echo.

echo.

echo [BUILD OK]

echo.

pause
exit

:F

echo.
echo [BUILD FAILED]
echo.
echo =================================
echo Compile failed, check for error!
echo =================================
echo.

pause