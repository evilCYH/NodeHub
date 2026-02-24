# NodeHub 节点重命名模板指南

NodeHub 节点重命名功能允许用户自定义节点名称的显示格式，通过灵活的模板语法实现个性化的节点管理体验。

---

## 📋 可用变量

重命名模板支持以下变量：

| 变量                    | 说明                | 示例               |
|-----------------------|-------------------|------------------|
| `{{.Count}}`          | 节点序号 (必填，从1开始)    | 1, 2, 3          |
| `{{.SpeedUp}}`        | 上行速度 (平均，单位：KB/s) | 102400, 51200    |
| `{{.SpeedDown}}`      | 下行速度 (平均，单位：KB/s) | 102400, 51200    |
| `{{.Delay}}`          | 延迟 (平均，单位：毫秒)     | 45, 120          |
| `{{.Risk}}`           | 风险等级 (数字越小越好)     | 1, 2, 3          |
| `{{.Country.NameEn}}` | 国家/地区代码           | JP, US, SG       |
| `{{.Country.NameZh}}` | 国家/地区中文名称         | 日本, 美国, 新加坡      |
| `{{.Country.Emoji}}`  | 国家/地区旗帜表情符号       | 🇯🇵, 🇺🇸, 🇸🇬 |
| `{{.SubName}}`        | 订阅名称              | 未知订阅             |
| `{{.SubTags}}`        | 订阅标签              | \<Tag1\|Tag2\>   |
| `{{.SubTagsOrigin}}`  | 订阅标签（原始数组）        | ["Tag1", "Tag2"] |

> 注意：`.SubTagsOrigin` 类型为 `[]string`，因此不能直接在重命名模板中使用

---

## 🚀 快速开始

### 立即可用的推荐模板

#### 简洁美观格式
```go
{{.Country.Emoji}}{{.Country.NameZh}}-{{.Delay}}ms
```
输出示例：`🇯🇵日本-45ms`, `🇺🇸美国-120ms`

#### 游戏玩家格式 (低延迟优先)
```go
{{if le .Delay 50}}🚀{{else if le .Delay 100}}⚡{{else}}🐌{{end}}{{.Country.NameZh}}
```
输出示例：`🚀日本`, `⚡美国`, `🐌其他`

#### 下载专用格式 (高速度优先)
```go
{{div .SpeedDown 1024}}MB/s-{{.Country.Emoji}}{{.Country.NameZh}}
```
输出示例：`100MB/s-🇯🇵日本`, `50MB/s-🇺🇸美国`

---

## 🎨 模板库

### 基础模板

#### 简单格式
```go
节点{{.Count}}
```
输出示例：`节点1`, `节点2`, `节点3`

#### 带国家信息
```go
{{.Country.NameZh}}-节点{{.Count}}
```
输出示例：`日本-节点1`, `美国-节点2`

#### 带国旗格式
```go
{{.Country.Emoji}}{{.Country.NameZh}}-{{.Count}}
```
输出示例：`🇯🇵日本-1`, `🇺🇸美国-2`

#### 基础性能格式
```go
{{.Country.Emoji}}{{.Country.NameZh}}-{{.Delay}}ms
```
输出示例：`🇯🇵日本-45ms`, `🇺🇸美国-120ms`

### 推荐模板

#### 格式1：简洁信息
```go
{{.Country.Emoji}}{{.Country.NameZh}}-{{.Delay}}ms
```
输出示例：`🇯🇵日本-45ms`, `🇺🇸美国-120ms`

#### 格式2：速度优先
```go
{{div .SpeedDown 1024}}MB/s-{{.Country.Emoji}}{{.Country.NameZh}}
```
输出示例：`100MB/s-🇯🇵日本`, `50MB/s-🇺🇸美国`

#### 格式3：完整信息
```go
{{printf "%03d" .Count}}-{{.Country.Emoji}}{{.Country.NameZh}}-{{.Delay}}ms-{{div .SpeedDown 1024}}MB/s
```
输出示例：`001-🇯🇵日本-45ms-100MB/s`

#### 格式4：质量评级
```go
{{.Country.Emoji}}{{.Country.NameZh}}-{{if le .Delay 50}}极速{{else if le .Delay 100}}快速{{else}}普通{{end}}
```
输出示例：`🇯🇵日本-极速`, `🇺🇸美国-快速`

### 高级模板

#### 数字格式化
```go
{{printf "%03d" .Count}}-{{.Country.NameZh}}
```
输出示例：`001-日本`, `002-美国`, `010-新加坡`

#### 条件判断
```go
{{.Country.NameZh}}{{if eq .Risk 1}}✅{{else if eq .Risk 2}}⚠️{{else}}❌{{end}}
```
输出示例：`日本✅`, `美国⚠️`, `其他❌`

#### 速度质量标识
```go
{{if ge .SpeedDown 51200}}🚀{{else if ge .SpeedDown 10240}}⚡{{else}}🐌{{end}}{{.Country.NameZh}}
```
输出示例：`🚀日本`, `⚡美国`, `🐌其他`

#### 延迟分级
```go
{{if le .Delay 50}}极速{{else if le .Delay 100}}快速{{else if le .Delay 200}}普通{{else}}较慢{{end}}-{{.Country.NameZh}}
```
输出示例：`极速-日本`, `快速-美国`, `普通-新加坡`

### 场景专用模板

#### 游戏玩家专用
```go
🎮{{printf "%03d" .Count}}-{{.Country.Emoji}}{{.Country.NameZh}}-{{if le .Delay 50}}4星{{else if le .Delay 100}}3星{{else}}2星{{end}}
```

#### 工作办公专用
```go
💼{{.Country.NameZh}}-{{if le .Delay 100}}稳定{{else}}一般{{end}}-{{div .SpeedDown 1024}}MB
```

#### 视频流媒体专用
```go
📺{{.Country.Emoji}}{{.Country.NameZh}}-{{if ge .SpeedDown 25600}}4K{{else if ge .SpeedDown 10240}}1080P{{else}}720P{{end}}
```

#### 开发者专用
```go
{{.Country.NameEn}}{{printf "%03d" .Count}}|{{.Delay}}ms|{{div .SpeedDown 1024}}MB|Risk{{.Risk}}
```

#### 专业监控风格
```go
[{{.Country.NameEn}}] Ping:{{.Delay}}ms|Down:{{div .SpeedDown 1024}}MB/s|Risk:{{.Risk}}
```
输出示例：`[JP] Ping:45ms|Down:100MB/s|Risk:1`

#### 质量评分系统
```go
{{.Country.Emoji}}{{.Country.NameZh}}-{{if ge .SpeedDown 51200}}{{if le .Delay 50}}S{{else if le .Delay 100}}A{{else}}B{{end}}{{else if ge .SpeedDown 10240}}{{if le .Delay 50}}A{{else if le .Delay 100}}B{{else}}C{{end}}{{else}}C{{end}}
```
输出示例：`🇯🇵日本-S`, `🇺🇸美国-B`, `🇸🇬新加坡-C`

---

## 📚 函数参考

### 数学运算
- `add x y` - 加法：`{{add .Count 1}}`
- `sub x y` - 减法：`{{sub 100 .Delay}}`
- `div x y` - 除法：`{{div .SpeedDown 1024}}`
- `mod x y` - 取余：`{{mod .Count 2}}`

### 比较运算
- `eq x y` - 等于：`{{if eq .Risk 1}}安全{{end}}`
- `ne x y` - 不等于：`{{if ne .Delay 0}}正常{{end}}`
- `lt x y` - 小于：`{{if lt .Delay 50}}极速{{end}}`
- `le x y` - 小于等于：`{{if le .Delay 100}}快速{{end}}`
- `gt x y` - 大于：`{{if gt .SpeedDown 51200}}高速{{end}}`
- `ge x y` - 大于等于：`{{if ge .SpeedDown 10240}}可用{{end}}`

### 逻辑运算
- `and x y` - 逻辑与：`{{if and (ge .SpeedDown 51200) (le .Delay 50)}}极品{{end}}`
- `or x y` - 逻辑或：`{{if or (le .Delay 50) (ge .SpeedDown 51200)}}推荐{{end}}`
- `not x` - 逻辑非：`{{if not (eq .Risk 3)}}安全{{end}}`

### 字符串处理
- `printf format args...` - 格式化：`{{printf "%03d" .Count}}`
- `slice s start end` - 切片：`{{slice .Country.NameEn 0 2}}`

### 数组处理

+ `for index item` - 循环：`<{{range $i, $v := .SubTags}}{{if $i}}|{{end}}{{$v}}{{end}}>`

---

## 💡 使用技巧

### 单位转换
- KB/s 转 MB/s：`{{div .SpeedDown 1024}}MB/s`
- ms 转 s：`{{div .Delay 1000}}.{{mod .Delay 1000}}s`
- 智能速度单位：`{{if ge .SpeedDown 1024}}{{div .SpeedDown 1024}}MB/s{{else}}{{.SpeedDown}}KB/s{{end}}`

### 条件组合
```go
{{if and (ge .SpeedDown 51200) (le .Delay 50)}}🚀{{else if and (ge .SpeedDown 10240) (le .Delay 100)}}⚡{{else}}🐌{{end}}{{.Country.NameZh}}
```

### 性能分级
```go
{{.Country.NameZh}}-{{if le .Delay 30}}S+{{else if le .Delay 50}}S{{else if le .Delay 100}}A{{else if le .Delay 200}}B{{else}}C{{end}}
```

### 风险标识
```go
{{.Country.Emoji}}{{.Country.NameZh}}{{if eq .Risk 1}}🟢{{else if eq .Risk 2}}🟡{{else if eq .Risk 3}}🟠{{else}}🔴{{end}}
```

---

## 📖 快速参考

### 常用阈值
- **延迟等级**：
  - 极速：≤ 50ms
  - 快速：≤ 100ms
  - 普通：≤ 200ms
  - 较慢：> 200ms

- **速度等级** (KB/s)：
  - 极速：≥ 51200 (50MB/s)
  - 快速：≥ 10240 (10MB/s)
  - 普通：≥ 2048 (2MB/s)
  - 较慢：< 2048

- **视频质量要求**：
  - 4K：≥ 25600 KB/s
  - 1080P：≥ 10240 KB/s
  - 720P：≥ 5120 KB/s

### 颜色建议
- 🟢 安全：风险等级 1
- 🟡 注意：风险等级 2
- 🟠 警告：风险等级 3
- 🔴 危险：风险等级 ≥ 4

---

## ⚠️ 注意事项

- **必填项**：每个模板都必须包含 `{{.Count}}` 变量
- **单位说明**：速度变量单位为 KB/s，延迟变量单位为毫秒
- **语法规范**：使用 Go 语言的 `text/template` 语法
- **大小写敏感**：所有变量名区分大小写，请确保使用正确的变量名
- **字符转义**：模板中的引号需要转义，如 `\"`

---

## ❓ 常见问题

**Q: 模板中必须包含哪些变量？**
A: `{{.Count}}` 是必填项，其他变量可根据需要选择使用

**Q: 如何将速度单位从 KB/s 转换为 MB/s？**
A: 使用 `{{div .SpeedDown 1024}}MB/s` 进行单位转换

**Q: 如何根据延迟给节点进行分级显示？**
A: 使用条件判断语句，如 `{{if le .Delay 50}}极速{{end}}`

**Q: 模板支持哪些数学运算？**
A: 支持加减乘除、取余等基本数学运算

**Q: 模板语法错误应该如何排查？**
A: 请检查变量名大小写、括号匹配和条件语句完整性

**Q: 为什么我的模板没有生效？**
A: 请检查是否包含了必填的 `{{.Count}}` 变量

**Q: 如何将节点序号显示为三位数格式？**
A: 使用 `{{printf "%03d" .Count}}` 可以格式化为 001, 002...

**Q: 如何根据不同的速度显示对应的图标？**
A: 使用条件判断：`{{if ge .SpeedDown 51200}}🚀{{else}}⚡{{end}}`

