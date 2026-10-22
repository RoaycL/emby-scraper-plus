# Emby Scraper Plus 🍑

<p align="center">
  <strong>专为 Emby / Jellyfin 打造的增强型媒体元数据刮削服务端</strong><br>
  结合 MetaTube 协议架构与 Peach 智能路由及极速 GraphQL 数据源
</p>

---

## 🌟 项目亮点

1. **协议 100% 兼容 MetaTube 插件**
   - 原生兼容 `jellyfin-plugin-metatube`（在 Emby 中安装该插件后，直接将服务器地址指向本项目即可，无需改动插件）。
2. **极速 DMM / FANZA 官方 GraphQL 直连**
   - 采用官方 GraphQL 接口（`https://api.video.dmm.co.jp/graphql`），单次请求仅需 0.3 ~ 1 秒，免翻墙、免验证码，全面覆盖当月新作与独立小厂。
3. **r18.dev 镜像多语言与日文汉字修正**
   - 自动获取日文原题与演员汉字正名，避免英文罗马音与机翻标题。
4. **AVBase `__NEXT_DATA__` 单次结构化提取**
   - 告别繁琐易碎的 HTML DOM 正则，直接从 Next.js 服务端渲染树中解析多店铺（FANZA / MGS / DUGA）商品数据。
5. **Peach 启发式智能路由与字段短路**
   - 自动区分普通有码、素人系、无码与 FC2 体系，优先走高可靠度官方接口，必要标量齐全即自动短路终止，避免频繁请求被目标站风控封 IP。
6. **Gfriends 演员头像自动补齐**
   - 内置 Gfriends CDN 自动索引，演员头像秒速匹配呈现。
7. **自带现代化 Web 测试控制台**
   - 浏览器打开 `http://<IP>:8080` 即可实时测试任意番号刮削效果、预览海报大图，并支持**一键下载标准 Kodi/Emby NFO 边车文件**。

---

## 🚀 fnOS（飞牛云）/ NAS 极速部署指南

### 方式 1：Docker Compose 部署（推荐）

在 fnOS 的应用中心或终端创建 `docker-compose.yml`：

```yaml
services:
  emby-scraper-plus:
    image: roaycl/emby-scraper-plus:latest
    container_name: emby-scraper-plus
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      # 若 NAS 需要配置科学代理访问部分外网，可取消以下注释：
      # - HTTP_PROXY=http://192.168.1.100:7890
      # - HTTPS_PROXY=http://192.168.1.100:7890
```

启动容器：
```bash
docker compose up -d
```

### 方式 2：Docker 命令行运行

```bash
docker run -d \
  --name emby-scraper-plus \
  --restart unless-stopped \
  -p 8080:8080 \
  roaycl/emby-scraper-plus:latest
```

---

## 📺 在 fnOS Emby 中配置使用

1. 打开 Emby 管理控制台 → **插件** → 安装 `MetaTube`（或导入 `Jellyfin.Plugin.MetaTube.dll`）。
2. 进入 **MetaTube 插件设置**：
   - **服务器地址 (Server)**：填写 `http://<你的fnOS_IP>:8080`（例如 `http://192.168.1.10:8080`）。
   - **Token**：若未在服务端配置 `TOKEN` 环境变量，保持留空即可。
3. 进入 Emby **媒体库** 设置：
   - 在电影媒体库中勾选 **MetaTube** 为元数据与图片下载器。
4. 扫描媒体库或点击“识别”，系统将通过智能路由秒级入库高清封面、简介与演员头像！

---

## 🛠 本地源码编译与运行

需要 Go 1.22+：

```bash
git clone https://github.com/RoaycL/emby-scraper-plus.git
cd emby-scraper-plus

# 编译二进制文件
go build -o bin/emby-scraper-plus ./cmd/server

# 启动运行
./bin/emby-scraper-plus -port 8080
```

打开浏览器访问 `http://localhost:8080` 即可使用可视化测试看板。

---

## 🔌 API 接口概览

兼容 MetaTube RESTful v1 标准：

| 端点 | 方法 | 说明 |
| --- | --- | --- |
| `/v1/movies/search?q={keyword}` | GET | 智能路由番号检索 |
| `/v1/movies/:provider/:id` | GET | 获取影片详细元数据（`:provider` 可传 `SMART`） |
| `/v1/actors/search?q={name}` | GET | 演员搜索 |
| `/v1/actors/:provider/:id` | GET | 演员信息与 Gfriends 头像检索 |
| `/v1/images/primary/:provider/:id` | GET | 封面海报图片代理与缓存 |
| `/v1/nfo/:provider/:id` | GET | 直接导出为标准 Kodi/Emby `.nfo` 边车文件 |
| `/healthz` | GET | 健康检查探针 |

---

## 📜 开源许可

本项目基于 [MIT License](LICENSE) 开源。感谢 [metatube-community](https://github.com/metatube-community)、[peach-mitao](https://github.com/peach-mitao/peach) 与 [gfriends](https://github.com/gfriends/gfriends) 社区提供的优秀灵感与开源组件。
