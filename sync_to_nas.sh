#!/bin/bash
# ZeroQuant2 代码同步到 NAS
# 运行前确保 NAS 已挂载

echo "===== ZeroQuant2 NAS 同步 ====="

# 检查 NAS 挂载
if [ ! -d "/tmp/nas_home" ]; then
    echo "挂载 NAS..."
    mount -t cifs //100.65.205.77/home /tmp/nas_home -o username=zero,password=zq215007,vers=3.0
fi

NAS_DIR="/tmp/nas_home/ZeroQuant2"

# 同步修改的文件
echo "1. 同步 config.go..."
cp -f src/backend/internal/config/config.go "$NAS_DIR/src/backend/internal/config/"

echo "2. 同步 middleware.go..."
cp -f src/backend/internal/middleware/middleware.go "$NAS_DIR/src/backend/internal/middleware/"

echo "3. 同步 websocket.go..."
cp -f src/backend/internal/websocket/websocket.go "$NAS_DIR/src/backend/internal/websocket/"

echo "4. 同步 strategy_service.go..."
cp -f src/backend/internal/service/strategy_service.go "$NAS_DIR/src/backend/internal/service/"

echo "5. 同步 main.go..."
cp -f src/backend/cmd/server/main.go "$NAS_DIR/src/backend/cmd/server/"

echo "6. 同步 config.yaml..."
cp -f src/backend/config.yaml "$NAS_DIR/src/backend/"

echo "7. 同步开发记录..."
cp -f docs/开发记录/2026-05-13_Go后端代码审查.md "$NAS_DIR/docs/开发记录/"

echo "8. 同步开发进度看板..."
cp -f docs/开发进度看板.md "$NAS_DIR/docs/"

echo "===== 完成 ====="
echo "2026-05-13 同步完成"