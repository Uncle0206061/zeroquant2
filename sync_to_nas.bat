@echo off
REM ZeroQuant2 代码同步脚本
REM 运行前确保 NAS 已挂载为 Z: 盘符

echo ===== ZeroQuant2 NAS 同步 =====

echo 挂载 NAS（如果未挂载）:
echo    net use Z: \\100.65.205.77\home robot00.00 rb123456 /persistent:yes
net use Z: \\100.65.205.77\home robot00.00 rb123456 /persistent:yes

echo.
echo 1. 同步修改的文件...
copy /Y "D:\ZeroQuant2\src\backend\internal\config\config.go" "Z:\ZeroQuant2\src\backend\internal\config\"
copy /Y "D:\ZeroQuant2\src\backend\internal\middleware\middleware.go" "Z:\ZeroQuant2\src\backend\internal\middleware\"
copy /Y "D:\ZeroQuant2\src\backend\internal\websocket\websocket.go" "Z:\ZeroQuant2\src\backend\internal\websocket\"
copy /Y "D:\ZeroQuant2\src\backend\internal\service\strategy_service.go" "Z:\ZeroQuant2\src\backend\internal\service\"
copy /Y "D:\ZeroQuant2\src\backend\cmd\server\main.go" "Z:\ZeroQuant2\src\backend\cmd\server\"
copy /Y "D:\ZeroQuant2\src\backend\config.yaml" "Z:\ZeroQuant2\src\backend\"

echo 2. 同步配置文件和文档...
copy /Y "D:\ZeroQuant2\docs\开发进度看板.md" "Z:\ZeroQuant2\docs\"
copy /Y "D:\ZeroQuant2\docs\开发记录\2026-05-13_Go后端代码审查.md" "Z:\ZeroQuant2\docs\开发记录\"

echo 3. 同步 CLAUDE.md...
copy /Y "D:\ZeroQuant2\CLAUDE.md" "Z:\ZeroQuant2\"

echo ===== 完成 =====
pause