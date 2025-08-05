# One API 的前端界面 / Frontend Templates

> 每个文件夹代表一个主题，欢迎提交你的主题
> Each folder represents a theme/template, and we welcome your theme submissions

> [!WARNING]
> 不是每一个主题都及时同步了所有功能，由于精力有限，优先更新默认主题，其他主题欢迎 & 期待 PR
> Not every theme is synchronized with all features in a timely manner. Due to limited resources, the default theme is updated first. PRs for other themes are welcomed & expected.

## 开发指南 / Development Guide

### 可用模板 / Available Templates

| 模板 / Template | UI框架 / Framework | 端口 / Port | 目录 / Directory |
|----------------|-------------------|------------|------------------|
| **Default** | Semantic UI React | 3001 | `./default/` |
| **Air** | Semi UI | 3002 | `./air/` |
| **Berry** | Material UI | 3003 | `./berry/` |

### 快速开发启动 / Quick Development Start

#### 1. 启动Go后端 / Start Go Backend (Required)

```bash
# 从项目根目录 / From project root
go run main.go
```

#### 2. 选择模板并启动开发 / Choose Template and Start Development

```bash
# 从项目根目录 / From project root
./dev.sh default dev    # Default template on port 3001
./dev.sh air dev        # Air template on port 3002
./dev.sh berry dev      # Berry template on port 3003

# 或使用make / Or use make:
make dev-default        # Default template
make dev-air           # Air template
make dev-berry         # Berry template
```

### 开发地址 / Development URLs

- **Default Template**: http://localhost:3001
- **Air Template**: http://localhost:3002
- **Berry Template**: http://localhost:3003

所有模板自动代理API调用到Go后端: `http://100.113.170.10:3000`
All templates automatically proxy API calls to Go backend: `http://100.113.170.10:3000`

### 生产构建 / Production Build

```bash
# 构建单个模板 / Build individual template
./dev.sh default build     # or: make build-frontend-default
./dev.sh air build         # or: make build-frontend-air
./dev.sh berry build       # or: make build-frontend-berry

# 构建所有模板 / Build all templates
make build-all-templates
```

详细开发指南请参阅: [`../docs/DEVELOPMENT.md`](../docs/DEVELOPMENT.md)
For detailed development guide, see: [`../docs/DEVELOPMENT.md`](../docs/DEVELOPMENT.md)

## 提交新的主题

> 欢迎在页面底部保留你和 One API 的版权信息以及指向链接

1. 在 `web` 文件夹下新建一个文件夹，文件夹名为主题名。
2. 把你的主题文件放到这个文件夹下。
3. 修改你的 `package.json` 文件，把 `build` 命令改为：`"build": "react-scripts build && mv -f build ../build/default"`，其中 `default` 为你的主题名。
4. 修改 `common/config/config.go` 中的 `ValidThemes`，把你的主题名称注册进去。
5. 修改 `web/THEMES` 文件，这里也需要同步修改。

## 主题列表

### 主题：default

默认主题，由 [JustSong](https://github.com/songquanpeng) 开发。

预览：
|![image](https://github.com/songquanpeng/one-api/assets/39998050/ccfbc668-3a7f-4bc1-87da-7eacfd7bf371)|![image](https://github.com/songquanpeng/one-api/assets/39998050/a63ed547-44b9-45db-b43a-ecea07d60840)|
|:---:|:---:|

### 主题：berry

由 [MartialBE](https://github.com/MartialBE) 开发。

预览：
|||
|:---:|:---:|
|![image](https://github.com/songquanpeng/one-api/assets/42402987/36aff5c6-c5ff-4a90-8e3d-33d5cff34cbf)|![image](https://github.com/songquanpeng/one-api/assets/42402987/9ac63b36-5140-4064-8fad-fc9d25821509)|
|![image](https://github.com/songquanpeng/one-api/assets/42402987/fb2b1c64-ef24-4027-9b80-0cd9d945a47f)|![image](https://github.com/songquanpeng/one-api/assets/42402987/b6b649ec-2888-4324-8b2d-d5e11554eed6)|
|![image](https://github.com/songquanpeng/one-api/assets/42402987/6d3b22e0-436b-4e26-8911-bcc993c6a2bd)|![image](https://github.com/songquanpeng/one-api/assets/42402987/eef1e224-7245-44d7-804e-9d1c8fa3f29c)|

### 主题：air
由 [Calon](https://github.com/Calcium-Ion) 开发。
|![image](https://github.com/songquanpeng/songquanpeng.github.io/assets/39998050/1ddb274b-a715-4e81-858b-857d520b6ff4)|![image](https://github.com/songquanpeng/songquanpeng.github.io/assets/39998050/163b0b8e-1f73-49cb-b632-3dcb986b56d5)|
|:---:|:---:|


#### 开发说明

请查看 [web/berry/README.md](https://github.com/songquanpeng/one-api/tree/main/web/berry/README.md)
