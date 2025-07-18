#!/bin/bash

# 设置环境变量禁用CGO
export CGO_ENABLED=0

# 清理并创建目标目录
sudo rm -rf /usr/local/cambridge
sudo mkdir -p /usr/local/cambridge

# 编译
go build -o cambridge

# 复制二进制文件和字典文件
sudo cp cambridge /usr/local/cambridge/
sudo cp -r dict /usr/local/cambridge/

# 设置权限
sudo chmod +x /usr/local/cambridge/cambridge

# 创建systemd服务文件
sudo tee /etc/systemd/system/cambridge.service << EOF
[Unit]
Description=Cambridge Dictionary Service
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/cambridge/cambridge -addr 0.0.0.0 -port 8010 -dict /usr/local/cambridge/dict
Restart=always
User=root
WorkingDirectory=/usr/local/cambridge

[Install]
WantedBy=multi-user.target
EOF

# 重新加载systemd配置
sudo systemctl daemon-reload

echo "Build and installation completed."
echo "To start the service: sudo systemctl start cambridge"
echo "To enable service on boot: sudo systemctl enable cambridge"
echo "To check service status: sudo systemctl status cambridge"
echo "To stop the service: sudo systemctl stop cambridge"
