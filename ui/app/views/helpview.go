package views

import (
	"strings"

	"github.com/darkhz/bluetuith/ui/keybindings"
	"github.com/darkhz/bluetuith/ui/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

// helpView holds the help view.
type helpView struct {
	popup modalView
	page  string
	area  *tview.Flex

	topics map[string][]HelpData

	*Views
}

// Initialize initializes the help view.
func (h *helpView) Initialize() error {
	if !h.cfg.Values.NoHelpDisplay {
		h.area = tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(horizontalLine(), 1, 0, false).
			AddItem(h.status.Help, 1, 0, false)
	}

	h.initHelpData()
	h.statusHelpArea(true)

	return nil
}

// SetRootView sets the root view for the help view.
func (h *helpView) SetRootView(v *Views) {
	h.Views = v
}

// statusHelpArea shows or hides the status help text.
func (h *helpView) statusHelpArea(add bool) {
	if h.cfg.Values.NoHelpDisplay {
		return
	}

	if !add && h.area != nil {
		h.layout.RemoveItem(h.area)
		return
	}

	h.layout.AddItem(h.area, 2, 0, false)
}

// swapStatusHelp adds or removes the provided primitive from the layout and displays
// the help text below the statusbar.
func (h *helpView) swapStatusHelp(primitive tview.Primitive, add bool) {
	h.statusHelpArea(false)
	defer h.statusHelpArea(true)

	if add {
		h.layout.AddItem(primitive, 8, 0, false)
	} else {
		h.layout.RemoveItem(primitive)
	}
}

// showStatusHelp shows a condensed help text for the currently focused screen below the statusbar.
func (h *helpView) showStatusHelp(page string) {
	if h.cfg.Values.NoHelpDisplay || h.page == page {
		return
	}

	h.page = page
	pages := map[string]string{
		devicePage.String():     "设备界面",
		filePickerPage.String(): "文件选择",
		progressPage.String():   "进度视图",
	}

	items, ok := h.topics[pages[page]]
	if !ok {
		h.status.Help.Clear()
		return
	}

	groups := map[string][]HelpData{}

	for _, item := range items {
		if !item.ShowInStatus {
			continue
		}

		var group string

		for _, key := range item.Keys {
			switch key {
			case keybindings.KeyMenu, keybindings.KeySwitch:
				group = "打开"

			case keybindings.KeyFilebrowserSelect, keybindings.KeyFilebrowserInvertSelection, keybindings.KeyFilebrowserSelectAll:
				group = "选择"

			case keybindings.KeyProgressTransferSuspend, keybindings.KeyProgressTransferResume, keybindings.KeyProgressTransferCancel:
				group = "传输"

			case keybindings.KeyDeviceConnect, keybindings.KeyDevicePair, keybindings.KeyAdapterToggleScan, keybindings.KeyAdapterTogglePower:
				group = "开关"
			}
		}
		if group == "" {
			group = item.Title
		}

		helpItem := groups[group]
		if helpItem == nil {
			helpItem = []HelpData{}
		}

		helpItem = append(helpItem, item)
		groups[group] = helpItem
	}

	var text strings.Builder
	count := 0
	for group, items := range groups {
		var names, keys []string

		for _, item := range items {
			if item.Title != group {
				names = append(names, item.Title)
			}
			for _, k := range item.Keys {
				keys = append(keys, h.kb.Name(h.kb.Data(k).Kb))
			}
		}
		if names != nil {
			group += " " + strings.Join(names, "/")
		}

		helpKeys := strings.Join(keys, "/")
		if count < len(groups)-1 {
			helpKeys += ", "
		}

		title := theme.ColorWrap(theme.ThemeText, group, "::bu")
		helpKeys = theme.ColorWrap(theme.ThemeText, ": "+helpKeys)

		text.WriteString(title)
		text.WriteString(helpKeys)
		count++
	}

	h.status.Help.SetText(text.String())
}

// showHelp displays a modal with the help items for all the screens.
func (h *helpView) showHelp() {
	var row int

	helpModal := h.modals.newModalWithTable("help", "帮助", 40, 60)
	helpModal.table.SetSelectionChangedFunc(func(row, _ int) {
		if row == 1 {
			helpModal.table.ScrollToBeginning()
		}
	})
	helpModal.table.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if action == tview.MouseScrollUp {
			helpModal.table.InputHandler()(tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModNone), nil)
		}

		return action, event
	})

	for title, helpItems := range h.topics {
		helpModal.table.SetCell(
			row, 0, tview.NewTableCell("[::bu]"+title).
				SetSelectable(false).
				SetAlign(tview.AlignCenter).
				SetTextColor(theme.GetColor(theme.ThemeText)),
		)

		row++

		for _, item := range helpItems {
			var names []string

			for _, k := range item.Keys {
				names = append(names, h.kb.Name(h.kb.Data(k).Kb))
			}

			keybinding := strings.Join(names, "/")

			helpModal.table.SetCell(
				row, 0, tview.NewTableCell(theme.ColorWrap(theme.ThemeText, item.Description)).
					SetExpansion(1).
					SetAlign(tview.AlignLeft).
					SetTextColor(theme.GetColor(theme.ThemeText)).
					SetSelectedStyle(tcell.Style{}.Reverse(true)),
			)

			helpModal.table.SetCell(
				row, 1, tview.NewTableCell(theme.ColorWrap(theme.ThemeText, keybinding)).
					SetExpansion(0).
					SetAlign(tview.AlignLeft).
					SetTextColor(theme.GetColor(theme.ThemeText)).
					SetSelectedStyle(tcell.Style{}.Reverse(true)),
			)

			row++
		}

		row++

	}

	helpModal.show()
}

// HelpData describes the help item.
type HelpData struct {
	Title, Description string
	Keys               []keybindings.Key
	ShowInStatus       bool
}

// initHelpData initializes the help data for all the specified screens.
func (h *helpView) initHelpData() {
	h.topics = map[string][]HelpData{
		"设备界面": {
			{"菜单", "打开菜单", []keybindings.Key{keybindings.KeyMenu}, true},
			{"切换", "在菜单间切换", []keybindings.Key{keybindings.KeySwitch}, true},
			{"移动", "在设备/选项间移动", []keybindings.Key{keybindings.KeyNavigateUp, keybindings.KeyNavigateDown}, true},
			{"电源", "开关适配器电源", []keybindings.Key{keybindings.KeyAdapterTogglePower}, true},
			{"可被发现", "开关可被发现状态", []keybindings.Key{keybindings.KeyAdapterToggleDiscoverable}, false},
			{"可配对", "开关可配对状态", []keybindings.Key{keybindings.KeyAdapterTogglePairable}, false},
			{"扫描", "开关扫描（发现设备）", []keybindings.Key{keybindings.KeyAdapterToggleScan}, true},
			{"适配器", "切换适配器", []keybindings.Key{keybindings.KeyAdapterChange}, true},
			{"发送", "发送文件", []keybindings.Key{keybindings.KeyDeviceSendFiles}, true},
			{"网络", "连接网络", []keybindings.Key{keybindings.KeyDeviceNetwork}, false},
			{"进度", "进度视图", []keybindings.Key{keybindings.KeyProgressView}, false},
			{"播放器", "显示/隐藏播放器", []keybindings.Key{keybindings.KeyPlayerShow, keybindings.KeyPlayerHide}, false},
			{"设备信息", "显示设备信息", []keybindings.Key{keybindings.KeyDeviceInfo}, false},
			{"连接", "连接/断开选中设备", []keybindings.Key{keybindings.KeyDeviceConnect}, true},
			{"配对", "配对/取消配对选中设备", []keybindings.Key{keybindings.KeyDevicePair}, true},
			{"信任", "信任/取消信任选中设备", []keybindings.Key{keybindings.KeyDeviceTrust}, false},
			{"移除", "从适配器移除设备", []keybindings.Key{keybindings.KeyDeviceRemove}, false},
			{"取消", "取消操作", []keybindings.Key{keybindings.KeyCancel}, false},
			{"帮助", "显示帮助", []keybindings.Key{keybindings.KeyHelp}, true},
			{"退出", "退出", []keybindings.Key{keybindings.KeyQuit}, false},
		},
		"文件选择": {
			{"移动", "在目录条目间移动", []keybindings.Key{keybindings.KeyNavigateUp, keybindings.KeyNavigateDown}, true},
			{"进入/返回目录", "进入目录 / 返回上级", []keybindings.Key{keybindings.KeyNavigateRight, keybindings.KeyNavigateLeft}, true},
			{"单个", "选择单个文件", []keybindings.Key{keybindings.KeyFilebrowserSelect}, true},
			{"反选", "反选文件", []keybindings.Key{keybindings.KeyFilebrowserInvertSelection}, true},
			{"全部", "全选文件", []keybindings.Key{keybindings.KeyFilebrowserSelectAll}, true},
			{"刷新", "刷新当前目录", []keybindings.Key{keybindings.KeyFilebrowserRefresh}, false},
			{"隐藏文件", "显示/隐藏隐藏文件", []keybindings.Key{keybindings.KeyFilebrowserToggleHidden}, false},
			{"确认", "确认选择", []keybindings.Key{keybindings.KeyFilebrowserConfirmSelection}, true},
			{"退出", "退出", []keybindings.Key{keybindings.KeyClose}, false},
		},
		"进度视图": {
			{"移动", "在传输任务间移动", []keybindings.Key{keybindings.KeyNavigateUp, keybindings.KeyNavigateDown}, true},
			{"挂起", "挂起传输", []keybindings.Key{keybindings.KeyProgressTransferSuspend}, true},
			{"恢复", "恢复传输", []keybindings.Key{keybindings.KeyProgressTransferResume}, true},
			{"取消", "取消传输", []keybindings.Key{keybindings.KeyProgressTransferCancel}, true},
			{"退出", "退出", []keybindings.Key{keybindings.KeyClose}, true},
		},
		"媒体播放器": {
			{"播放/暂停", "播放/暂停", []keybindings.Key{keybindings.KeyNavigateUp, keybindings.KeyNavigateDown}, false},
			{"下一个", "下一个", []keybindings.Key{keybindings.KeyPlayerNext}, false},
			{"上一个", "上一个", []keybindings.Key{keybindings.KeyPlayerPrevious}, false},
			{"快退", "快退", []keybindings.Key{keybindings.KeyPlayerSeekBackward}, false},
			{"快进", "快进", []keybindings.Key{keybindings.KeyPlayerSeekForward}, false},
			{"停止", "停止", []keybindings.Key{keybindings.KeyPlayerStop}, false},
		},
	}
}
