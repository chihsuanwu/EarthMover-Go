# Frontend Modernization

前端程式碼從 2017 年的 C++ 時代原封不動搬過來，本次升級在不改變功能和 UI 的前提下，將技術棧現代化。

## 變更摘要

18 個檔案，刪除 1674 行，新增 675 行（淨減 999 行）。

## 移除的依賴

### Firebase SDK (v4.1.3)
- 移除 SDK `<script>` 和初始化 config（含 API key）
- 移除所有 `firebase.database()` 呼叫（5 處）
- **Watch Game 功能暫時移除**：此功能完全依賴 Firebase Realtime Database 來同步棋局，移除 Firebase 後無法運作
- 移除僅供 Watch Game 使用的 `Board.put()` 方法

### jQuery (v3.2.1)
- 移除 CDN `<script>`
- 約 40 處 jQuery 呼叫全部替換為原生 DOM API：

| jQuery | 替代方案 |
|--------|---------|
| `$('#id')` | `document.getElementById()` |
| `$('.x input').prop('disabled', true)` | `setDisabled()` helper（新增） |
| `$('#x').toggle()` | `classList.toggle('hidden')` |
| `$('input[name="x"]:checked').val()` | `document.querySelector().value` |
| `$(el).click(fn)` | `el.addEventListener('click', fn)` |
| `span.children().eq(i).find(...)` | `span.children[i].querySelector(...)` |

## JavaScript 現代化

### XHR → Fetch API
`post()` 函數從手動 Promise + XMLHttpRequest 改為 `async/await` + `fetch()`：
```js
// Before (25 行)
function post(params, path) {
  return new Promise(function(resolve, reject) {
    var req = new XMLHttpRequest();
    ...
  });
}

// After (8 行)
async function post(params, path) {
  params = params || {};
  params.sessionId = sessionManager.getSessionID();
  const res = await fetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(params),
  });
  if (res.status === 204) return;
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}
```

### ES5 → ES2020+

| 舊寫法 | 新寫法 |
|--------|--------|
| `var Board = function() { ... }` + `Board.prototype.x = ...` | `class Board { ... }` |
| `var` | `const` / `let` |
| `function(x) { return x; }` | `(x) => x` |
| `.then(function onSuccess() { ... })` | `.then(() => { ... })` |
| `'string ' + var + ' string'` | `` `string ${var} string` `` |

涉及檔案：`board.js`、`dialog.js`、`timer.js`、`main.js`。

### Tree Visualization (`treevisualize.js`)
- 原本有兩個版本：`treevisualize.js`（C++ 時代原版）和 `treevisualize-min.js`（Go 移植後修改版）。HTML 載入的是 min 版
- 將 min 版的邏輯合併回 `treevisualize.js`，刪除 min 版
- 修正 JSON 欄位名稱，對齊 Go 後端的縮寫 tag（`i`, `tc`, `wr`, `wol`, `wt`, `ch`）
- 加入 `d3.hierarchy(treeData, d => d.ch)` — 告訴 D3 children 欄位叫 `ch`
- XHR 改用共用的 `post()` 函數
- JS 語法升級（`const/let`、arrow functions、template literals）

## HTML 清理

- 加入 `<meta charset="UTF-8">`、`<meta name="viewport">`、`<title>`、`lang="en"`
- `checked="true"` → `checked`（boolean attribute）
- `disabled="true"` → `disabled`
- 移除 `type="text/javascript"`（HTML5 預設）
- 移除舊的 GitHub ribbon（指向 C++ 舊 repo）
- 移除未使用的 `#prompt-game-id` div
- Script 標籤移至 `</body>` 前
- Google Fonts URL 升級為 `css2` API

## CSS 清理

- **移除 55 處 vendor prefix**：`-webkit-`、`-moz-`、`-o-`、`-ms-`、`-khtml-`，這些 prefix 針對 2010 年代瀏覽器，現代瀏覽器已不需要
- 移除 `@-webkit-keyframes`、`@-moz-keyframes` 重複定義，只保留 `@keyframes`
- 新增 `.hidden { display: none !important; }` 工具類別
- `.tl-check` 改用 `.hidden` class 控制可見性
- Toolbar game submenu 高度從 `10.6vh` 調整為 `5.4vh`（移除 watch game 後只剩一個項目）

## 刪除的檔案

| 檔案 | 原因 |
|------|------|
| `dashboard.css` | 未被任何頁面引用 |
| `dashboard.js` | 依賴已排除的 `proc_stat` 模組 |
| `treevisualize-min.js` | 邏輯已合併回 `treevisualize.js` |

## 後端修正

`internal/mcts/tree.go` 的 `TreeJSON` struct 無變更，但在排查過程中確認前端必須使用縮寫欄位名（`i`, `tc`, `wr`, `wol`, `wt`, `ch`）對齊後端 JSON tag。

## 未升級項目

| 項目 | 原因 |
|------|------|
| D3.js（維持 v4） | v7 的 transition API 與現有的 nested transition pattern 不相容，會導致視覺異常。D3 是純視覺化 library，無安全疑慮 |
| 不引入 React/Vue | 單頁 Canvas 遊戲，vanilla JS 足夠 |
| 不加 build 工具 | 只有 5 個 JS 檔 + 1 個 CDN，不需要打包 |
| 不轉 TypeScript | 程式碼量小（~700 行 JS），收益不大 |
