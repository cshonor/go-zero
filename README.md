# Go-Zero Project

这是一个使用 go-zero 框架创建的 API 服务项目。

## 项目结构

```
go zero/
├── etc/                    # 配置文件目录
│   └── gozero-api.yaml    # API 服务配置
├── internal/              # 内部代码
│   ├── config/           # 配置结构
│   ├── handler/          # 处理器
│   ├── logic/            # 业务逻辑
│   ├── svc/              # 服务上下文
│   └── types/            # 类型定义
├── go.mod                # Go 模块文件
├── main.go               # 入口文件
└── README.md            # 项目说明
```

## 运行项目

1. 安装依赖：
```bash
go mod tidy
```

2. 运行服务：
```bash
go run main.go
```

或者指定配置文件：
```bash
go run main.go -f etc/gozero-api.yaml
```

3. 访问 API：
```bash
curl http://localhost:8888/api/hello
```

## API 端点

- `GET /api/hello` - Hello 接口示例

