> ## ⚠️ 本仓库说明
>
> **这是 [bluetuith-org/bluetuith](https://github.com/bluetuith-org/bluetuith) 的简体中文汉化版本（个人自用，非官方项目）。**
>
> | 项 | 说明 |
> |---|---|
> | 上游项目 | https://github.com/bluetuith-org/bluetuith |
> | 本仓库性质 | 个人汉化，未向上游提交 PR |
> | 汉化分支 | `chinese-i18n` |
> | 汉化方式 | 直接翻译源码内界面字面量（上游无 i18n 框架） |
> | 汉化范围 | 仅界面文字，未改动任何业务逻辑 |
> | 原许可证 | 遵循上游 LICENSE（见仓库内 LICENSE 文件） |
>
> 上游更新后同步方法：
> ```bash
> git fetch upstream
> git merge upstream/master
> ```
>
> 另修正一处上游缺陷：`go.mod` 声明的 module 路径为 `github.com/bluetuith-org/bluetuith`，
> 但全部 import 仍写 `github.com/darkhz/bluetuith`，导致上游代码无法 `go build`；
> 本分支将 module 名改回实际 import 路径以恢复可编译。
>
> ---

[![Go Report Card](https://goreportcard.com/badge/github.com/darkhz/bluetuith)](https://goreportcard.com/report/github.com/darkhz/bluetuith) [![Packaging status](https://repology.org/badge/tiny-repos/bluetuith.svg)](https://repology.org/project/bluetuith/versions)

![demo](demo/demo.gif)

# bluetuith
bluetuith is a TUI-based bluetooth connection manager, which can interact with bluetooth adapters and devices.
It aims to be a replacement to most bluetooth managers, like blueman.

This project is currently in the alpha stage.

## Funding

This project is funded through [NGI Zero Core](https://nlnet.nl/core), a fund established by [NLnet](https://nlnet.nl) with financial support from the European Commission's [Next Generation Internet](https://ngi.eu) program. Learn more at the [NLnet project page](https://nlnet.nl/project/bluetuith).

[<img src="https://nlnet.nl/logo/banner.png" alt="NLnet foundation logo" width="20%" />](https://nlnet.nl)
[<img src="https://nlnet.nl/image/logos/NGI0_tag.svg" alt="NGI Zero Logo" width="20%" />](https://nlnet.nl/core)

## Features
### UI
- Adapter selector and status indicators
- Mouse support
- File transfer progress view
- Authentication view for pairing and file transfers
- Device interaction menus and options
- Media player

### Platform-specific
The available features per-platform are [here](https://github.com/bluetuith-org/bluetooth-classic?tab=readme-ov-file#feature-matrix).
Please view the documentation for a list of operating requirements per platform.

## Documentation
The documentation is now hosted [here](https://bluetuith-org.github.io/bluetuith/).

The wiki is out-of-date.


[![Packaging status](https://repology.org/badge/vertical-allrepos/bluetuith.svg)](https://repology.org/project/bluetuith/versions)


