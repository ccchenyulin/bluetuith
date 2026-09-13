package keybindings

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Key describes the application keybinding type.
type Key string

// The different application keybinding types.
const (
	KeyMenu                        Key = "Menu"
	KeySelect                      Key = "Select"
	KeyCancel                      Key = "Cancel"
	KeySuspend                     Key = "Suspend"
	KeyQuit                        Key = "Quit"
	KeySwitch                      Key = "Switch"
	KeyClose                       Key = "Close"
	KeyHelp                        Key = "Help"
	KeyAdapterChange               Key = "AdapterChange"
	KeyAdapterTogglePower          Key = "AdapterTogglePower"
	KeyAdapterToggleDiscoverable   Key = "AdapterToggleDiscoverable"
	KeyAdapterTogglePairable       Key = "AdapterTogglePairable"
	KeyAdapterToggleScan           Key = "AdapterToggleScan"
	KeyDeviceSendFiles             Key = "DeviceSendFiles"
	KeyDeviceNetwork               Key = "DeviceNetwork"
	KeyDeviceConnect               Key = "DeviceConnect"
	KeyDevicePair                  Key = "DevicePair"
	KeyDeviceTrust                 Key = "DeviceTrust"
	KeyDeviceBlock                 Key = "DeviceBlock"
	KeyDeviceAudioProfiles         Key = "DeviceAudioProfiles"
	KeyDeviceInfo                  Key = "DeviceInfo"
	KeyDeviceRemove                Key = "DeviceRemove"
	KeyPlayerShow                  Key = "PlayerShow"
	KeyPlayerHide                  Key = "PlayerHide"
	KeyFilebrowserDirForward       Key = "FilebrowserDirForward"
	KeyFilebrowserDirBack          Key = "FilebrowserDirBack"
	KeyFilebrowserSelect           Key = "FilebrowserSelect"
	KeyFilebrowserInvertSelection  Key = "FilebrowserInvertSelection"
	KeyFilebrowserSelectAll        Key = "FilebrowserSelectAll"
	KeyFilebrowserRefresh          Key = "FilebrowserRefresh"
	KeyFilebrowserToggleHidden     Key = "FilebrowserToggleHidden"
	KeyFilebrowserConfirmSelection Key = "FilebrowserConfirmSelection"
	KeyProgressView                Key = "ProgressView"
	KeyProgressTransferSuspend     Key = "ProgressTransferSuspend"
	KeyProgressTransferResume      Key = "ProgressTransferResume"
	KeyProgressTransferCancel      Key = "ProgressTransferCancel"
	KeyPlayerTogglePlay            Key = "PlayerTogglePlay"
	KeyPlayerNext                  Key = "PlayerNext"
	KeyPlayerPrevious              Key = "PlayerPrevious"
	KeyPlayerSeekForward           Key = "PlayerSeekForward"
	KeyPlayerSeekBackward          Key = "PlayerSeekBackward"
	KeyPlayerStop                  Key = "PlayerStop"
	KeyNavigateUp                  Key = "NavigateUp"
	KeyNavigateDown                Key = "NavigateDown"
	KeyNavigateRight               Key = "NavigateRight"
	KeyNavigateLeft                Key = "NavigateLeft"
	KeyNavigateTop                 Key = "NavigateTop"
	KeyNavigateBottom              Key = "NavigateBottom"
)

// Context describes the context where the keybinding is
// supposed to be applied in.
type Context string

// The different context types for keybindings.
const (
	ContextApp      Context = "App"
	ContextDevice   Context = "Device"
	ContextFiles    Context = "Files"
	ContextProgress Context = "进度"
)

// KeyData stores the metadata for the key.
type KeyData struct {
	Title   string
	Context Context
	Kb      Keybinding
	Global  bool
}

// Keybinding stores the keybinding.
type Keybinding struct {
	Key  tcell.Key
	Rune rune
	Mod  tcell.ModMask
}

// Keybindings contains an entire list of keybindings and its associated contexts and other data.
type Keybindings struct {
	keyData        map[Key]*KeyData
	contextKeys    map[Context]map[Keybinding]Key
	navigationKeys map[Key]Keybinding
	translateKeys  map[string]string
}

// NewKeybindings returns a new keybindings configuration.
func NewKeybindings() *Keybindings {
	k := &Keybindings{}

	k.initData()
	k.initKeys()

	return k
}

// Data returns the key data associated with
// the provided keyID and operation name.
func (k *Keybindings) Data(key Key) *KeyData {
	return k.keyData[key]
}

// Initialize initializes all the keybindings by context.
func (k *Keybindings) Initialize() {
	for keyName, key := range k.keyData {
		if k.contextKeys[key.Context] == nil {
			k.contextKeys[key.Context] = make(map[Keybinding]Key)
		}

		k.contextKeys[key.Context][key.Kb] = keyName
	}
}

// Key returns the operation name for the provided keyID
// and the keyboard event.
func (k *Keybindings) Key(event *tcell.EventKey, keyContexts ...Context) Key {
	ch := event.Rune()
	if event.Key() != tcell.KeyRune {
		ch = ' '
	}

	mod := event.Modifiers()
	if unicode.IsUpper(ch) && mod&tcell.ModShift != 0 {
		mod &^= tcell.ModShift
	}

	kb := Keybinding{event.Key(), ch, mod}

	if key, ok := k.checkContexts(kb, keyContexts); ok {
		return key
	}

	if key, ok := k.checkContexts(kb, []Context{
		ContextApp,
		ContextDevice,
	}); ok {
		return key
	}

	return ""
}

// Name formats and returns the key's name.
func (k *Keybindings) Name(kb Keybinding) string {
	if kb.Key == tcell.KeyRune {
		keyname := string(kb.Rune)
		if kb.Rune == ' ' {
			keyname = "Space"
		}

		if kb.Mod&tcell.ModAlt != 0 {
			keyname = "Alt+" + keyname
		}

		return keyname
	}

	return tcell.NewEventKey(kb.Key, kb.Rune, kb.Mod).Name()
}

// IsNavigation checks whether the provided key is a navigation key.
func (k *Keybindings) IsNavigation(pressed Key, event *tcell.EventKey) (*tcell.EventKey, bool) {
	kb := Keybinding{event.Key(), event.Rune(), event.Modifiers()}
	if kb.Key != tcell.KeyRune {
		kb.Rune = ' '
	}

	n, ok := k.navigationKeys[pressed]
	if !ok || n == kb {
		return nil, false
	}

	return tcell.NewEventKey(n.Key, n.Rune, n.Mod), true
}

// Validate validates the keybindings from the configuration.
func (k *Keybindings) Validate(kbMap map[string]string) error {
	if len(kbMap) == 0 {
		return nil
	}

	keyNames := make(map[string]tcell.Key)
	for key, names := range tcell.KeyNames {
		keyNames[names] = key
	}

	for keyType, key := range kbMap {
		k.checkBindings(keyType, key, keyNames)
	}

	keyErrors := make(map[Keybinding]string)

	for keyType, keydata := range k.keyData {
		for existing, data := range k.keyData {
			if data.Kb == keydata.Kb && data.Title != keydata.Title {
				if data.Context == keydata.Context || data.Global || keydata.Global {
					goto KeyError
				}

				continue

			KeyError:
				if _, ok := keyErrors[keydata.Kb]; !ok {
					keyErrors[keydata.Kb] = fmt.Sprintf("- %s will override %s (%s)", keyType, existing, k.Name(keydata.Kb))
				}
			}
		}
	}

	if len(keyErrors) > 0 {
		err := "Config: The following keybindings will conflict:\n"
		for _, ke := range keyErrors {
			err += ke + "\n"
		}

		return errors.New(strings.TrimRight(err, "\n"))
	}

	return nil
}

// checkContexts checks whether a keybinding exists within the provided keybinding context.
func (k *Keybindings) checkContexts(kb Keybinding, contexts []Context) (Key, bool) {
	for _, context := range contexts {
		if operation, ok := k.contextKeys[context][kb]; ok {
			return operation, true
		}

	}

	return "", false
}

// checkBindings validates the provided keybinding.
//
//gocyclo:ignore
func (k *Keybindings) checkBindings(keyType, key string, keyNames map[string]tcell.Key) error {
	var runes []rune
	var keys []tcell.Key

	if _, ok := k.keyData[Key(keyType)]; !ok {
		return fmt.Errorf("config: Invalid key type %s", keyType)
	}

	keybinding := Keybinding{
		Key:  tcell.KeyRune,
		Rune: ' ',
		Mod:  tcell.ModNone,
	}

	tokens := strings.FieldsFunc(key, func(c rune) bool {
		return unicode.IsSpace(c) || c == '+'
	})

	for _, token := range tokens {
		length := runewidth.StringWidth(token)
		if length > 1 {
			token = cases.Title(language.Und, cases.NoLower).String(token)
		} else if length == 1 {
			c, _ := utf8.DecodeRuneInString(token)

			keybinding.Rune = rune(c)
			runes = append(runes, keybinding.Rune)

			continue
		}

		if translated, ok := k.translateKeys[token]; ok {
			token = translated
		}

		switch token {
		case "Ctrl":
			keybinding.Mod |= tcell.ModCtrl

		case "Alt":
			keybinding.Mod |= tcell.ModAlt

		case "Shift":
			keybinding.Mod |= tcell.ModShift

		case "Space", "Plus":
			keybinding.Rune = ' '
			if token == "Plus" {
				keybinding.Rune = '+'
			}

			runes = append(runes, keybinding.Rune)

		default:
			if key, ok := keyNames[token]; ok {
				keybinding.Key = key
				keybinding.Rune = ' '
				keys = append(keys, keybinding.Key)
			}
		}
	}

	if keys != nil && runes != nil || len(runes) > 1 || len(keys) > 1 {
		return fmt.Errorf("config: More than one key entered for %s (%s)", keyType, key)
	}

	if keybinding.Mod&tcell.ModShift != 0 {
		keybinding.Rune = unicode.ToUpper(keybinding.Rune)

		if unicode.IsLetter(keybinding.Rune) {
			keybinding.Mod &^= tcell.ModShift
		}
	}

	if keybinding.Mod&tcell.ModCtrl != 0 {
		var modKey string

		switch {
		case len(keys) > 0:
			if key, ok := tcell.KeyNames[keybinding.Key]; ok {
				modKey = key
			}

		case len(runes) > 0:
			if keybinding.Rune == ' ' {
				modKey = "Space"
			} else {
				modKey = string(unicode.ToUpper(keybinding.Rune))
			}
		}

		if modKey != "" {
			modKey = "Ctrl-" + modKey
			if key, ok := keyNames[modKey]; ok {
				keybinding.Key = key
				keybinding.Rune = ' '
				keys = append(keys, keybinding.Key)
			}
		}
	}

	if keys == nil && runes == nil {
		return fmt.Errorf("config: No key specified or invalid keybinding for %s (%s)", keyType, key)
	}

	k.keyData[Key(keyType)].Kb = keybinding

	return nil
}

// initKeys initializes and stores the key types and contexts.
func (k *Keybindings) initKeys() {
	k.contextKeys = make(map[Context]map[Keybinding]Key)

	k.navigationKeys = map[Key]Keybinding{
		KeyNavigateUp:     {tcell.KeyUp, ' ', tcell.ModNone},
		KeyNavigateDown:   {tcell.KeyDown, ' ', tcell.ModNone},
		KeyNavigateRight:  {tcell.KeyRight, ' ', tcell.ModNone},
		KeyNavigateLeft:   {tcell.KeyLeft, ' ', tcell.ModNone},
		KeyNavigateTop:    {tcell.KeyPgUp, ' ', tcell.ModNone},
		KeyNavigateBottom: {tcell.KeyPgDn, ' ', tcell.ModNone},
	}

	k.translateKeys = map[string]string{
		"Pgup":      "PgUp",
		"Pgdn":      "PgDn",
		"Pageup":    "PgUp",
		"Pagedown":  "PgDn",
		"Upright":   "UpRight",
		"Downright": "DownRight",
		"Upleft":    "UpLeft",
		"Downleft":  "DownLeft",
		"Prtsc":     "Print",
		"Backspace": "Backspace2",
	}
}

// initData initializes and stores the keybindings configuration.
func (k *Keybindings) initData() {
	k.keyData = map[Key]*KeyData{
		KeySwitch: {
			Title:   "切换",
			Context: ContextApp,
			Kb:      Keybinding{tcell.KeyTab, ' ', tcell.ModNone},
			Global:  true,
		},
		KeyClose: {
			Title:   "关闭",
			Context: ContextApp,
			Kb:      Keybinding{tcell.KeyEscape, ' ', tcell.ModNone},
			Global:  true,
		},
		KeyQuit: {
			Title:   "退出",
			Context: ContextApp,
			Kb:      Keybinding{tcell.KeyRune, 'Q', tcell.ModNone},
			Global:  true,
		},
		KeyMenu: {
			Title:   "菜单",
			Context: ContextApp,
			Kb:      Keybinding{tcell.KeyRune, 'm', tcell.ModAlt},
		},
		KeySelect: {
			Title:   "选择",
			Context: ContextApp,
			Kb:      Keybinding{tcell.KeyEnter, ' ', tcell.ModNone},
			Global:  true,
		},
		KeyCancel: {
			Title:   "取消",
			Context: ContextApp,
			Kb:      Keybinding{tcell.KeyCtrlX, ' ', tcell.ModCtrl},
			Global:  true,
		},
		KeySuspend: {
			Title:   "挂起",
			Context: ContextApp,
			Kb:      Keybinding{tcell.KeyCtrlZ, ' ', tcell.ModCtrl},
			Global:  true,
		},
		KeyHelp: {
			Title:   "帮助",
			Context: ContextApp,
			Kb:      Keybinding{tcell.KeyRune, '?', tcell.ModShift},
			Global:  true,
		},
		KeyNavigateUp: {
			Title:   "上移",
			Context: ContextApp,
			Kb:      Keybinding{tcell.KeyUp, ' ', tcell.ModNone},
		},
		KeyNavigateDown: {
			Title:   "下移",
			Context: ContextApp,
			Kb:      Keybinding{tcell.KeyDown, ' ', tcell.ModNone},
		},
		KeyNavigateRight: {
			Title:   "右移",
			Context: ContextApp,
			Kb:      Keybinding{tcell.KeyRight, ' ', tcell.ModNone},
		},
		KeyNavigateLeft: {
			Title:   "左移",
			Context: ContextApp,
			Kb:      Keybinding{tcell.KeyLeft, ' ', tcell.ModNone},
		},
		KeyNavigateTop: {
			Title:   "移到顶部",
			Context: ContextApp,
			Kb:      Keybinding{tcell.KeyPgUp, ' ', tcell.ModNone},
		},
		KeyNavigateBottom: {
			Title:   "移到底部",
			Context: ContextApp,
			Kb:      Keybinding{tcell.KeyPgDn, ' ', tcell.ModNone},
		},
		KeyAdapterTogglePower: {
			Title:   "电源",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'o', tcell.ModNone},
		},
		KeyAdapterToggleDiscoverable: {
			Title:   "可被发现",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'S', tcell.ModNone},
		},
		KeyAdapterTogglePairable: {
			Title:   "可配对",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'P', tcell.ModNone},
		},
		KeyAdapterToggleScan: {
			Title:   "扫描",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 's', tcell.ModNone},
		},
		KeyAdapterChange: {
			Title:   "切换",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'a', tcell.ModNone},
		},
		KeyDeviceConnect: {
			Title:   "连接",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'c', tcell.ModNone},
		},
		KeyDevicePair: {
			Title:   "配对",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'p', tcell.ModNone},
		},
		KeyDeviceTrust: {
			Title:   "信任",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 't', tcell.ModNone},
		},
		KeyDeviceBlock: {
			Title:   "阻止",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'b', tcell.ModNone},
		},
		KeyDeviceSendFiles: {
			Title:   "发送",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'f', tcell.ModNone},
		},
		KeyDeviceNetwork: {
			Title:   "网络选项",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'n', tcell.ModNone},
		},
		KeyDeviceAudioProfiles: {
			Title:   "音频配置",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'A', tcell.ModNone},
		},
		KeyDeviceInfo: {
			Title:   "Info",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'i', tcell.ModNone},
		},
		KeyDeviceRemove: {
			Title:   "移除",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'd', tcell.ModNone},
		},
		KeyPlayerShow: {
			Title:   "显示播放器",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'm', tcell.ModNone},
		},
		KeyPlayerHide: {
			Title:   "隐藏播放器",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'M', tcell.ModNone},
		},
		KeyPlayerTogglePlay: {
			Title:   "播放/暂停",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, ' ', tcell.ModNone},
		},
		KeyPlayerNext: {
			Title:   "下一个",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, '>', tcell.ModNone},
		},
		KeyPlayerPrevious: {
			Title:   "上一个",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, '<', tcell.ModNone},
		},
		KeyPlayerSeekForward: {
			Title:   "快进",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRight, ' ', tcell.ModNone},
		},
		KeyPlayerSeekBackward: {
			Title:   "快退",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyLeft, ' ', tcell.ModNone},
		},
		KeyPlayerStop: {
			Title:   "停止",
			Context: ContextDevice,
			Kb:      Keybinding{tcell.KeyRune, ']', tcell.ModNone},
		},
		KeyFilebrowserConfirmSelection: {
			Title:   "确认选择",
			Context: ContextFiles,
			Kb:      Keybinding{tcell.KeyCtrlS, ' ', tcell.ModCtrl},
		},
		KeyFilebrowserDirForward: {
			Title:   "进入目录",
			Context: ContextFiles,
			Kb:      Keybinding{tcell.KeyRight, ' ', tcell.ModNone},
		},
		KeyFilebrowserDirBack: {
			Title:   "返回上级",
			Context: ContextFiles,
			Kb:      Keybinding{tcell.KeyLeft, ' ', tcell.ModNone},
		},
		KeyFilebrowserSelect: {
			Title:   "选择",
			Context: ContextFiles,
			Kb:      Keybinding{tcell.KeyRune, ' ', tcell.ModNone},
		},
		KeyFilebrowserInvertSelection: {
			Title:   "反选",
			Context: ContextFiles,
			Kb:      Keybinding{tcell.KeyRune, 'a', tcell.ModNone},
		},
		KeyFilebrowserSelectAll: {
			Title:   "全选",
			Context: ContextFiles,
			Kb:      Keybinding{tcell.KeyRune, 'A', tcell.ModNone},
		},
		KeyFilebrowserRefresh: {
			Title:   "刷新",
			Context: ContextFiles,
			Kb:      Keybinding{tcell.KeyCtrlR, ' ', tcell.ModCtrl},
		},
		KeyFilebrowserToggleHidden: {
			Title:   "隐藏文件",
			Context: ContextFiles,
			Kb:      Keybinding{tcell.KeyRune, 'h', tcell.ModCtrl},
		},
		KeyProgressTransferResume: {
			Title:   "恢复传输",
			Context: ContextProgress,
			Kb:      Keybinding{tcell.KeyRune, 'g', tcell.ModNone},
		},
		KeyProgressTransferCancel: {
			Title:   "取消传输",
			Context: ContextProgress,
			Kb:      Keybinding{tcell.KeyRune, 'x', tcell.ModNone},
		},
		KeyProgressView: {
			Title:   "查看下载",
			Context: ContextProgress,
			Kb:      Keybinding{tcell.KeyRune, 'v', tcell.ModNone},
		},
		KeyProgressTransferSuspend: {
			Title:   "挂起传输",
			Context: ContextProgress,
			Kb:      Keybinding{tcell.KeyRune, 'z', tcell.ModNone},
		},
	}
}
