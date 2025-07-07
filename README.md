# View-Counter

后端练手项目。

## 项目简介

一个轻量级的网站访问统计服务，提供以下核心功能：

- **网站访问量统计**：分别统计总访问量、每日访问量
- **多网站支持**，自动识别来源域名
- **基于 SQLite 的数据持久化存储**
- **防滥用机制**，基于 IP 访问速率限制
- ~~**自动数据归档**（定期归档历史数据）~~（废案）

### 核心功能

- RESTful API 接口
- 自动识别来源域名
- 并发安全
- 每日访问量统计
- 总访问量统计

### 高级功能

- 请求频率限制
- ~~数据归档~~
- 可扩展的统计功能

## 技术栈

- **编程语言**：Go 1.24.4
- **数据库**：SQLite3
- **Web 框架**：Golang 标准库
- **并发控制**： sync.Mutex

## 部署

1. **安装依赖**
```bash
go mod download
```

2. **构建与运行**
```bash
go build -o out/view-counter
```

服务默认监听 **8081** 端口。

### API

记录访问量：
```text
POST /api/view
```

获取访问量：
```text
GET /api/view
```

### 使用示例

实现记录和获取访问量的 JS 脚本：
```js
//views.js
(function () {
  const api = "<API 路径>";
  const el = document.getElementById("site-view-counter");

  if (!el) return;

  // POST
  fetch(api, { method: "POST" });

  // GET
  fetch(api)
    .then(res => res.text())
    .then(count => {
      el.textContent = "本站总访问量" + count + " 次";
    })
    .catch(() => {
      el.textContent = "访问量数据获取失败";
    });
})();
```

并在需要记录浏览数的网页添加：
```html
<script async src="<path>/views.js"></script>
<span id="site-view-counter">本站总访问量：加载中...</span>
```

## 后续改进计划
- [x] API 防止滥用功能
- [ ] 提供统计相关的 API
- [ ] 粒度细化
- [ ] 引入缓存
- [ ] 容器化
- [ ] CI/CD
- [ ] 推送功能

## 许可证

[MIT](./LICENSE)
