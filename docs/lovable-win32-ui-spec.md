# RelayPane Win32 UI 需求规格（Lovable → Cursor 交接版）

> **用途**：交给 [Lovable](https://lovable.dev) 做视觉与布局设计；完成后把本文件 + 设计稿 + 切图一并带回 Cursor，映射到 Go `walk`（Win32 原生）实现。  
> **产品**：RelayPane — 双栏 SFTP 文件管理器（类似 WinSCP / 经典 Windows 资源管理器）。  
> **硬性约束**：必须是 **Windows 原生桌面风格**，不是 Web / SaaS / Material / Fluent 2 圆角风。

---

## 1. 平台与风格（必读）

| 项 | 要求 |
|---|---|
| 目标平台 | Windows 10/11 桌面程序 |
| 实现栈（Cursor 侧） | Go + [walk](https://github.com/lxn/walk)（Win32 原生控件） |
| 视觉参考 | **WinSCP**、**Windows 资源管理器（经典列表视图）**、**记事本/系统对话框** |
| 禁止 | 大圆角卡片、渐变 Hero、Web Tab、侧边抽屉、Tailwind 阴影、移动端布局 |
| 字体 | 界面：`Segoe UI` 9pt；路径/命令输出：`Consolas` 9pt |
| 配色 | 系统默认灰白（`#F0F0F0` 窗口底、白底列表、标准按钮）；仅状态点用绿/黄/红 |
| DPI | 以 **96 DPI** 为基准标注像素；图标提供 16×16 与 32×32 两套 |
| 语言 | 简体中文为主稿，英文为对照列（见 §8 文案表） |

---

## 2. 主窗口结构（必须严格按此层级）

```
MainWindow 1280×760（最小 960×600）
├── MenuBar                    [设置] [功能] [关于]
├── SessionTabBar              高 22–26px，单行，左对齐
│   ├── [Tab: 服务器名 ×] ...  每项 max 160px，关闭图标贴文字
│   └── [+] 新建连接           14×14 图标
├── ToolBar                    高 24–28px
│   └── [连接] [刷新] [上传] [下载]   ← 文字按钮，非大图标条
├── HSplitter（左右 50%）
│   ├── LocalPane
│   └── RemotePane
└── StatusBar                  高 28–36px
    ├── ● 连接状态文字
    ├── （可选）ProgressBar 细条
    ├── HSpacer
    ├── 传输速度
    └── [重新连接]（断线时显示）
```

### 2.1 布局不变量（实现时必须遵守）

1. **Tab 栏、ToolBar、StatusBar 高度固定**，不可被内容撑高；只有中间 `HSplitter` 占满剩余高度。  
2. **左右面板必须顶部对齐**：标题行、导航行、文件列表上沿在同一水平线。  
3. **远程面板导航行**与本地同宽同高：本地多出的「磁盘 / 盘符 / 收藏 / 下拉」区域，远程侧用**等宽空白占位**，使路径框左缘对齐。  
4. **窗口尺寸**仅由用户拖边或双击标题栏改变；菜单/右键/刷新不得触发窗口缩放。

---

## 3. 双栏文件面板（LocalPane / RemotePane）

两栏结构**完全一致**（仅本地多控件、远程用占位对齐）：

```
VBox（顶部对齐 AlignTop）
├── Label 粗体          "本地" / "远程"
├── NavRow 高 22px
│   ├── ToolButton 16×16   up.png
│   ├── ToolButton 16×16   refresh.png
│   ├── [本地] ImageView 14×14 disk.png + ComboBox 宽56（盘符 C: D: …）
│   ├── [本地] ImageView 14×14 like.png + ComboBox 宽100（常用目录）
│   ├── [远程] 等宽占位（56+14+100+14 共 184px 左右）
│   └── LineEdit 路径（占满剩余宽度）
├── RenameBar（默认隐藏）  "重命名:" + LineEdit
└── TableView（Stretch 占满）
    Columns: Name 220 | Size 80 | Modified 140
    行高 ≈ 系统 Small Icons 列表
    本地/远程：Shell 图标 + 文件名（远程目录统一文件夹图标）
```

### 3.1 文件列表交互（设计稿需体现状态）

| 状态 | 视觉 |
|---|---|
| 普通行 | 白底 / 交替浅灰行 |
| 选中行 | 系统标准蓝色高亮 |
| 多选 | 多行蓝色高亮 |
| 父目录 | 可选 `..` 或返回箭头 |
| 拖拽中 | 自定义 copy.cur 光标（仅按住左键时） |
| 加载中 | 状态栏或列表内联 "Loading…"，**不要**在远程栏单独插一行把布局顶歪 |

### 3.2 右键菜单（上下文菜单）

**本地 / 远程**各自一套，标准 Win32 弹出菜单，无图标或仅系统小图标：

- 刷新、粘贴、新建文件夹、新建文件  
- 上传（本地）/ 下载（远程）  
- 复制、删除、重命名  
- 编辑（单文件文本/图片）  
- 分隔线分组符合 Windows 习惯  

---

## 4. 会话 Tab 栏

| 元素 | 规格 |
|---|---|
| Tab 标签 | `PushButton` 风格或扁平 Tab；文字截断 18 字 + `…`；max 宽 160px |
| 激活 Tab | 浅蓝底 `#DCEBFC` 或系统选中色；禁用点击 |
| 连接中 | 标签后 `…` |
| 断线 | 标签后 `!` |
| 关闭 | `close.png` 14×14，**紧贴** Tab 文字右侧，间距 0 |
| 新建 | `new.png` 14×14，位于所有 Tab 右侧，Tab 栏**左对齐**（右侧弹性空白） |

---

## 5. 工具栏与菜单

### 5.1 ToolBar（主窗口）

文字按钮：`连接` | `刷新` | `上传` | `下载`  
（与菜单功能重复即可，保持 WinSCP 式简单工具栏。）

### 5.2 MenuBar

**设置**
- 语言 → English / 简体中文  
- 我的服务器…  
- 云端同步…  
- ——  
- 退出  

**功能**
- 系统信息、网络信息、磁盘空间、目录占用、CPU/内存  
- ——  
- 同步 → 本地→服务器 / 服务器→本地  
- ——  
- 执行远程命令（菜单**不要**写 Ctrl+E）  

**关于**
- 检查更新、关于我们  

### 5.3 全局快捷键（设计稿脚注即可，不做菜单项）

| 快捷键 | 行为 |
|---|---|
| Ctrl+E | 打开「执行远程命令」对话框 |
| F5 | 刷新两侧列表 |
| F2 | 内联重命名 |
| Delete | 删除选中 |
| Ctrl+A | 全选当前面板 |
| Esc | 取消重命名 |

---

## 6. 对话框清单（每个需单独画一屏 96DPI mock）

### 6.1 连接服务器 `480×320+`
- 说明文字 + `ListBox` 已保存服务器（双击连接）  
- 底栏：`取消` | `新建连接…` | （弹性空白） | `连接`  

### 6.2 新建/编辑服务器 `~520×480`
- 表单：名称、主机、端口、用户名、认证方式、密钥路径、…  
- `仅连接` / `保存并连接`  

### 6.3 执行远程命令 `760×560`
```
[ LineEdit 命令输入                    ] [执行]
历史: [ComboBox 下拉________________] [保存置顶] [删除选中] [全部删除]
提示: 可选中输出文本后 Ctrl+C 复制
[ TextEdit 只读输出区 - Consolas，占满 ]
                                    [确定]
```
- 命令行 **Enter** → 系统 Beep + 执行  
- 「保存置顶」：命令写入置顶列表，历史过多时不删除置顶项  

### 6.4 冲突处理 `Overwrite / Skip / Rename / Cancel`
- 标准 `MessageBox` 或自定义小对话框  

### 6.5 功能对话框（统一风格）
- **系统信息 / 网络 / 磁盘 / 目录占用 / CPU内存**：`760×560` 左右，顶部 Refresh，内容区 `TextEdit` 或 `TableView`  
- **网络**：流量统计 + 端口列表 +「每 5 秒自动刷新」勾选  
- **目录占用**：树形表格，可 Up/Refresh  

### 6.6 云端同步、检查更新、关于、文本/图片编辑器
- 均为标准模态 `Dialog`，遵循系统按钮顺序（确定/取消，主按钮在右）  

---

## 7. 图标资产（Lovable 需导出 PNG）

路径约定：`internal/assets/*.png`（透明底 RGBA）

| 文件名 | 尺寸 | 用途 |
|---|---|---|
| `logo.png` | 256×256（含 32/16 衍生） | 窗口/任务栏图标 |
| `up.png` | 32×32 源图 | 上一级 |
| `refresh.png` | 32×32 | 刷新 |
| `close.png` | 32×32 | 关闭 Tab |
| `new.png` | 32×32 | 新建连接 |
| `disk.png` | 32×32 | 盘符区 |
| `like.png` | 32×32 | 常用目录 |
| `copy.cur` | 光标 | 跨栏拖拽（已有，勿改热区） |

风格：线性或扁平 **Windows 工具栏小图标**，16px 显示清晰；避免过大色块（曾出现整屏绿色 + 按钮问题）。

---

## 8. 关键文案（zh / en）

| Key | 中文 | English |
|---|---|---|
| AppTitle | RelayPane | RelayPane |
| Local / Remote | 本地 / 远程 | Local / Remote |
| Connect | 连接 | Connect |
| Upload / Download | 上传 / 下载 | Upload / Download |
| NotConnected | 未连接 | Not connected |
| Reconnect | 重新连接 | Reconnect |
| TransferIdle | — MB/s | — MB/s |
| CloseTab | 关闭会话 | Close session |
| NewTabConnect | 新建连接 | Connect |
| FeatShell | 执行远程命令 | Remote Command |
| FeatShellPin | 保存置顶 | Pin |
| ColName / Size / Modified | 名称 / 大小 / 修改时间 | Name / Size / Modified |

完整键名见仓库 `internal/i18n/i18n.go`。

---

## 9. 状态与主题色

| 语义 | 色值 | 控件 |
|---|---|---|
| 已连接 | `#2ECC71` | 状态栏 `●` |
| 连接中 | `#F1C40F` | 状态栏 `●` |
| 未连接/断线 | `#E74C4C` | 状态栏 `●` |
| 激活 Tab 背景 | `#DCEBFC` | Tab 容器 |
|  inactive Tab | `#F0F0F0` | Tab 容器 |

其余一律用 **Visual Styles 默认**，不要自造深色主题（除非另开需求）。

---

## 10. Lovable 交付物（带回 Cursor 用）

请按以下格式交付，便于直接落地 `internal/walkui/`：

1. **PDF/PNG 设计稿**  
   - 主窗口：已连接 / 未连接 / 断线 / 传输中（进度条可见）  
   - 主窗口：本地有文件 + 远程空列表（验证左右对齐）  
   - Tab 栏：0 个 Tab、1 个 Tab、多个 Tab  
   - 全部对话框（§6）  
   - 右键菜单展开态  

2. **`design-tokens.json`**（示例结构）  
```json
{
  "fontUI": "Segoe UI 9pt",
  "fontMono": "Consolas 9pt",
  "tabBarHeight": 24,
  "navRowHeight": 22,
  "navIconSize": 16,
  "tabIconSize": 14,
  "driveComboWidth": 56,
  "placesComboWidth": 100,
  "colors": { "connected": "#2ECC71", "tabActive": "#DCEBFC" }
}
```

3. **`component-tree.md`** — 与 §2、§3 一致的控件树 + 每个控件 **Margin/Padding/Min/Max Size**（96 DPI 像素）。

4. **`assets/` 目录** — §7 全部 PNG，32×32 源文件。

5. **（可选）HTML 静态 mock**  
   - 若 Lovable 只能出 Web：用 `gray` 系统色、`border: 1px solid #ACA899`、小图标 16px，**明确标注「仅视觉参考，非 Web 产品」**。

---

## 11. Cursor 集成说明（给未来自己）

Lovable 稿到位后，在 Cursor 中：

1. 对照 `component-tree.md` 改 `internal/walkui/mainwindow.go` 的 `declarative` 布局。  
2. 图标放入 `internal/assets/`，更新 `ui_icons.go` 的 embed。  
3. 颜色/高度常量写入 `uiicons.go` / `theme.go`。  
4. **不要**用 Web 代码；只取尺寸、间距、图标与状态图。  
5. 回归检查 §2.1 布局不变量（尤其 Tab 栏高度、左右对齐、窗口不因右键缩小）。

---

## 12. 验收标准（UI 层面）

- [ ] 第一眼看是 **Win32 原生工具**，不是网页套壳  
- [ ] Tab 栏单行 ~24px，`+` 与 Tab 左对齐，不居中浮空  
- [ ] 左右面板标题、导航、列表上沿对齐；路径框左缘对齐  
- [ ] 工具栏图标 16px，Tab 图标 14px，无 oversized 图标  
- [ ] 菜单栏无孤立「Ctrl+E」项；Ctrl+E 仍可打开远程命令  
- [ ] 远程命令：Enter 有反馈、历史+置顶+删除按钮齐全  
- [ ] 中英文案位置预留足够（Combo/按钮宽度）  
- [ ] 所有对话框按钮顺序符合 Windows 习惯  

---

*文档版本：2026-06-07 · 对应当前分支 `feat/walk-ui` / walk 实现*
