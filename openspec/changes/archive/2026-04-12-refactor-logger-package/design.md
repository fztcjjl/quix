## Context

quix 框架的 Logger 组件位于 `core/logger/` 包，提供 Logger 接口和三个适配器（slog、zerolog、zap）。当前使用方式为 `app.Logger().Info(ctx, msg, args...)`，需要通过 App 实例获取 Logger。

目标：包名从 `logger` 改为 `log`，增加全局默认实例，支持 `log.Info(ctx, msg, args...)` 开箱即用。

## Goals / Non-Goals

**Goals:**
- 包重命名 `core/logger` → `core/log`，使用 `log.Info()` 更简洁
- 全局默认 Logger，创建 App 后项目任何地方可直接使用
- App.New() 自动设置全局默认

**Non-Goals:**
- 修改 Logger 接口定义
- 修改三个适配器的实现逻辑
- 修改默认使用 zerolog 的决策

## Decisions

### 1. 全局默认 Logger 实现

```go
// core/log/log.go
var defaultLogger Logger = &noopLogger{}

func SetDefault(l Logger) { defaultLogger = l }

func Info(ctx context.Context, msg string, args ...any)  { defaultLogger.Info(ctx, msg, args...) }
func Error(ctx context.Context, msg string, args ...any) { defaultLogger.Error(ctx, msg, args...) }
func Warn(ctx context.Context, msg string, args ...any)  { defaultLogger.Warn(ctx, msg, args...) }
func Debug(ctx context.Context, msg string, args ...any) { defaultLogger.Debug(ctx, msg, args...) }
func With(args ...any) Logger                             { return defaultLogger.With(args...) }
```

**noopLogger**：初始默认值，所有方法为空操作，避免 nil panic。App.New() 创建真实 Logger 后替换。

### 2. 包名 log 遮蔽标准库

Go 标准库有 `log` 包，重命名后会遮蔽。影响：
- 使用 quix 的项目本身使用结构化日志，几乎不会同时需要 stdlib `log`
- 如确实需要，import 时用别名 `import stdlog "log"`

**理由**：`log.Info()` 比 `logger.Info()` 简洁，与行业惯例一致（slog、logrus 等都用短名）。

### 3. 包级函数放在 log.go 而非修改现有接口文件

现有文件 `logger.go`（或改名后 `log.go`）定义 Logger 接口。全局变量和包级函数追加到同一文件中，不单独创建文件，因为它们是同一个包的公共 API。

### 4. App.New() 自动设置全局默认

```go
func New(opts ...Option) *App {
    // ... create logger ...
    log.SetDefault(app.logger)
    // ...
}
```

**理由**：用户创建 App 后无需额外操作，全局 Logger 即可使用。

### 5. App.Logger() 方法保留

虽然全局 Logger 已经可用，但 `App.Logger()` 保留，用于需要特定 Logger 实例的场景（如测试中注入 mock）。

## Risks / Trade-offs

- [全局状态] 多个 App 实例会互相覆盖全局 Logger → 实际项目通常只有一个 App 实例，可接受
- [遮蔽 stdlib] → stdlib `log` 在使用结构化日志的项目中几乎不需要
- [noopLogger 测试隔离] → 测试中可通过 `log.SetDefault(mockLogger)` 重置
