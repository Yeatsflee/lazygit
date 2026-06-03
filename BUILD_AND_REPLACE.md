# 自定义构建与替换说明

本文档记录本仓库的一处自定义改动、如何从源码编译出新的 `lazygit`
二进制，以及如何替换本机已安装的版本（含回退方法）。

## 本次改动：用一个开关永久禁用 custom patch

新增配置项 `gui.disableCustomPatch`（默认 `false`，即完全保持 lazygit 原行为）。
设为 `true` 后，custom patch 功能从 UI 彻底进不去，且**误按 `Enter` 也不会再弹出
"Custom Patch" 面板**。

custom patch 在整个 lazygit 里只有三个启动入口，全部集中在 commit 文件列表
（在 Commits 面板选中某个 commit 按 `Enter` 进入的文件列表）。开关开启时：

- **`Enter` 文件** —— 不再构建 custom patch，而是直接聚焦主面板显示该文件的
  diff（"Patch"），**等同于按 `0`（Focus main view）**。
- **`<space>` 选择 / toggle-all 键** —— 构建 patch 的这两个快捷键**不再注册**
  （底部快捷键栏也不会显示它们）。
- **在聚焦的 diff（主面板 / 0 区）里鼠标点击某一行** —— 也不会再启动 custom patch。

由于以上三个入口被堵死后 patch 永远无法进入 active 状态，`<ctrl+p>` 的 patch
选项菜单、"Custom Patch" 副面板、模式指示器等全部随之自动失效，无需逐个处理。

未改动的行为：

- 在 Commits 面板对某个 commit 按 `Enter` 进入其文件列表 —— 不变。
- 在文件列表对**目录**按 `Enter` 折叠/展开 —— 不变。
- 其它任何面板的 `Enter` —— 不变。
- 开关保持默认 `false` 时，上述一切都与上游 lazygit 完全一致。

在 `~/.config/lazygit/config.yml` 里启用：

```yaml
gui:
  disableCustomPatch: true
```

涉及文件：

- `pkg/config/user_config.go` —— 新增 `DisableCustomPatch` 字段及默认值。
- `pkg/gui/controllers/commits_files_controller.go` —— 在三个入口处按开关
  early-return / 不注册键位 / 改为聚焦 diff。
- `docs-master/Config.md`、`schema-master/config.json` —— 由 `go generate`
  根据上面的配置项自动生成、保持同步。

## 前置条件：Go 工具链

构建需要 Go（版本见 `go.mod` 中的 `go` 指令，当前为 `1.25.0` 起）。

如果本机没装 Go，可以临时下载一份工具链到 `/tmp`（不污染系统）：

```bash
# 下载并解压一份独立的 Go 工具链
curl -sSL -o /tmp/go.tgz "https://go.dev/dl/$(curl -sS https://go.dev/VERSION?m=text | head -1).linux-amd64.tar.gz"
mkdir -p /tmp/goroot && tar -C /tmp/goroot -xzf /tmp/go.tgz
/tmp/goroot/go/bin/go version
```

> 注意：`/tmp` 在重启后通常会被清空，临时工具链与产物都会消失。
> 需要长期使用请把 Go 装到固定位置，或把编译好的二进制按下文复制到 `/usr/local/bin`。

## 编译新的二进制

在仓库根目录执行（若 Go 已在 `PATH` 中，可省略前面的 `export`）：

```bash
cd <仓库根目录>            # 例如 /home/xiaolongz/zxl/projects/tmp/lazygit

# 仅在使用 /tmp 临时工具链时需要这几行
export GOROOT=/tmp/goroot/go
export PATH=$GOROOT/bin:$PATH
export GOPATH=/tmp/gopath
export GOCACHE=/tmp/gocache
export GOFLAGS=-mod=vendor

# 编译，产物输出到 /tmp/lazygit-new
go build -o /tmp/lazygit-new ./

# 验证产物可运行
/tmp/lazygit-new --version
```

也可以用项目自带的 `make build`（同样需要 Go 在 `PATH` 中）：

```bash
make build      # 产物为 ./lazygit
```

### 快速验证本次改动

```bash
/tmp/lazygit-new
```

进入任意 git 仓库后：Commits 面板 → 在某个 commit 上按 `Enter` 进入文件列表 →
在某个**文件**上按 `Enter`，应直接看到主面板的 Patch diff，**不再弹出 Custom Patch**
（与按 `0` 效果一致）。

## 替换本机已安装的 lazygit

本机已安装路径通常是 `/usr/local/bin/lazygit`（该目录为 root 所有，需要 `sudo`）。
建议先备份再替换：

```bash
# 1. 备份现有版本（便于回退）
sudo cp /usr/local/bin/lazygit /usr/local/bin/lazygit.bak

# 2. 用新编译的二进制替换
sudo cp /tmp/lazygit-new /usr/local/bin/lazygit

# 3. 验证
lazygit --version
```

> 提示：从最新源码编译出的二进制，其 `--version` 显示的是 commit 短哈希而非
> 正式发布号（如 `0.60.0`）。这是正常现象。

## 回退到原版

```bash
sudo cp /usr/local/bin/lazygit.bak /usr/local/bin/lazygit
lazygit --version
```
