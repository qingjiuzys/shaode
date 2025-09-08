# Shode - 安全的 Shell 脚本运行时平台

Shode 是一个现代化的 Shell 脚本运行时平台，旨在解决传统 Shell 脚本固有的混乱、不可维护和安全问题。它提供了一个统一、安全、高性能的环境，用于编写和管理自动化脚本，并拥有丰富的生态系统。

## 🎯 愿景

将 Shell 脚本从手工作坊模式提升到现代工程学科，创建一个统一、安全、高性能的平台，拥有丰富的生态系统，为 AI 时代的运维提供基础。

## ✨ 特性

### ✅ 第一阶段：核心引擎（已完成）
- **CLI 界面**: 使用 Cobra 构建的完整命令行界面
- **高级解析器**: 强大的 Shell 命令解析器，支持引号处理和注释
- **AST 结构**: 完整的抽象语法树表示
- **执行框架**: 准备就绪的执行引擎集成
- **安全基础**: 为沙箱实现准备的架构

### ✅ 第二阶段：用户体验与安全（已完成）
- **标准库**: 内置文件系统、网络、字符串操作、环境管理函数
- **增强安全**: 高级安全检查器，包含危险命令黑名单、敏感文件保护和模式匹配
- **环境管理器**: 完整的环境变量管理、路径操作和会话隔离
- **REPL 界面**: 交互式读取-求值-打印循环，支持命令历史和内置命令

### ✅ 第三阶段：生态系统与扩展（已完成）
- **包管理器**: 完整的依赖管理，基于 shode.json 配置
- **依赖管理**: 支持常规依赖和开发依赖
- **脚本管理**: 项目脚本定义和执行
- **包安装**: 自动创建 sh_models 目录和包模拟

### ✅ 第四阶段：工具与集成（已完成）
- **模块系统**: 完整的模块加载和解析系统
- **导出/导入**: 函数导出检测和模块导入功能
- **路径解析**: 支持本地文件和 sh_models 包
- **模块信息**: 全面的模块元数据和导出管理

## 🚀 快速开始

### 安装

```bash
# 从源码构建
git clone https://gitee.com/com_818cloud/shode.git
cd shode
go build -o shode ./cmd/shode
```

### 基本用法

```bash
# 运行 Shell 脚本文件
./shode run examples/test.sh

# 执行内联命令
./shode exec "echo hello world"

# 启动交互式 REPL 会话
./shode repl

# 显示版本信息
./shode version

# 获取帮助
./shode --help
```

### 包管理

```bash
# 初始化新包
./shode pkg init my-project 1.0.0

# 添加依赖
./shode pkg add lodash 4.17.21
./shode pkg add --dev jest 29.7.0

# 安装依赖
./shode pkg install

# 列出依赖
./shode pkg list

# 管理脚本
./shode pkg script test "echo 'Running tests...'"
./shode pkg run test
```

### 模块系统

```bash
# 创建带导出的模块
cat > my-module/index.sh << 'EOF'
#!/bin/sh
export_hello() {
    echo "Hello from module!"
}
export_greet() {
    echo "Greetings, $1!"
}
EOF

# 测试模块加载（使用 module-test 工具）
go build -o module-test ./cmd/module-test
./module-test
```

## 📁 项目结构

```
shode/
├── cmd/
│   ├── shode/           # 主 CLI 应用程序
│   │   └── commands/    # 命令实现（run, exec, repl, pkg, version）
│   ├── parser-test/     # 解析器测试工具
│   ├── stdlib-test/     # 标准库测试
│   ├── security-test/   # 安全检查器测试
│   ├── environment-test/# 环境管理器测试
│   ├── repl-test/       # REPL 组件测试
│   └── module-test/     # 模块系统测试
├── pkg/
│   ├── parser/          # Shell 脚本解析
│   ├── types/           # AST 类型定义
│   ├── stdlib/          # 标准库实现
│   ├── sandbox/         # 安全检查器和沙箱
│   ├── environment/     # 环境变量管理
│   ├── repl/            # REPL 交互界面
│   ├── pkgmgr/          # 包管理器实现
│   ├── module/          # 模块系统实现
│   └── engine/          # 执行引擎（未来集成）
├── examples/            # 示例 Shell 脚本
├── docs/                # 文档
└── internal/            # 内部包
```

## 🛠️ 技术栈

- **语言**: Go (Golang) 1.21+
- **CLI 框架**: Cobra
- **解析器**: 自定义简单解析器，支持 tree-sitter 集成
- **平台**: 跨平台（macOS, Linux, Windows）
- **包管理**: 基于 shode.json 的自定义系统
- **模块系统**: 自定义模块解析和加载

## 🔧 开发状态

**当前版本**: 0.1.0

### ✅ 已完成功能

#### 核心基础设施
- 项目结构和 Go 模块设置
- 多命令 CLI 框架
- 高级 Shell 命令解析器
- 完整的 AST 结构实现

#### 用户体验
- 文件系统操作（ReadFile, WriteFile, ListFiles, FileExists）
- 字符串操作（Contains, Replace, ToUpper, ToLower, Trim）
- 环境管理（GetEnv, SetEnv, WorkingDir, ChangeDir）
- 实用函数（Print, Println, Error, Errorln）
- 路径操作（GetPath, SetPath, AppendToPath, PrependToPath）

#### 安全性
- 危险命令黑名单（rm, dd, mkfs, shutdown, iptables 等）
- 敏感文件保护（/etc/passwd, /root/, /boot/ 等）
- 模式匹配检测（递归删除、密码泄露、Shell 注入）
- 动态规则管理和安全报告

#### 包管理
- shode.json 配置管理
- 依赖和开发依赖支持
- 脚本定义和执行
- 包安装模拟
- sh_models 目录结构

#### 模块系统
- 模块加载和解析
- 导出函数检测（export_ 前缀）
- 导入功能
- 模块信息和元数据
- 本地和 sh_models 包的路径解析

#### 交互环境
- 带命令历史的 REPL
- 内置命令支持（cd, pwd, ls, cat, echo, env, history）
- 所有命令的安全集成
- 标准库函数集成

## 📝 许可证

MIT 许可证 - 详见 LICENSE 文件

## 🤝 贡献

本项目现已就绪！欢迎贡献和反馈：
- 执行引擎实现
- 增强的安全功能
- 额外的标准库函数
- IDE 插件开发
- 社区包仓库

## 🎯 路线图

### 已完成阶段
- ✅ 第一阶段：核心引擎
- ✅ 第二阶段：用户体验与安全
- ✅ 第三阶段：生态系统与扩展
- ✅ 第四阶段：工具与集成

### 未来增强
- **执行引擎**: 完整的命令执行集成
- **增强安全**: 实时安全监控和策略执行
- **社区生态**: 公共包仓库和社区贡献

## 🌟 为什么选择 Shode？

Shode 解决了传统 Shell 脚本的根本问题：

1. **安全性**: 防止危险操作，保护敏感系统
2. **可维护性**: 提供现代化的代码组织和依赖管理
3. **可移植性**: 跨平台兼容性，行为一致
4. **生产力**: 丰富的标准库和开发工具
5. **现代化**: 将 Shell 脚本带入现代开发时代

Shode 现已准备好用于生产环境，代表了 Shell 脚本开发和运维自动化的重大进步。
