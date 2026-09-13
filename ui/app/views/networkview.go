package views

import (
	"errors"
	"fmt"
	"strings"

	"go.uber.org/atomic"

	"github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	"github.com/darkhz/bluetuith/ui/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

// networkView holds the network selector view.
type networkView struct {
	isSupported atomic.Bool

	*Views
}

// Initialize initializes the network selector view.
func (n *networkView) Initialize() error {
	n.isSupported.Store(true)

	return nil
}

// SetRootView sets the root view for the network selector view.
func (n *networkView) SetRootView(v *Views) {
	n.Views = v
}

// networkSelect shows a popup to select the network type.
func (n *networkView) networkSelect() {
	if !n.isSupported.Load() {
		n.status.ErrorMessage(errors.New("this operation is not supported"))
		return
	}

	type nwTypeDesc struct {
		connType bluetooth.NetworkType
		desc     string
	}

	var connTypes []nwTypeDesc

	device := n.device.getSelection(false)
	if device.IsNil() {
		return
	}

	if device.HaveService(bluetooth.PanuServiceClass) {
		connTypes = append(connTypes, nwTypeDesc{
			bluetooth.NetworkPanu,
			"个人区域网（PAN）",
		})
	}
	if device.HaveService(bluetooth.DialupNetServiceClass) {
		connTypes = append(connTypes, nwTypeDesc{
			bluetooth.NetworkDun,
			"拨号网络（DUN）",
		})
	}

	deviceName := getDeviceDisplayName(device.DeviceEventData)

	if connTypes == nil {
		n.status.InfoMessage("没有可用的网络选项："+deviceName, false)
		return
	}

	n.menu.drawContextMenu(
		menuDeviceName.String(),
		func(networkMenu *tview.Table) {
			row, _ := networkMenu.GetSelection()

			cell := networkMenu.GetCell(row, 0)
			if cell == nil {
				return
			}

			connType, ok := cell.GetReference().(bluetooth.NetworkType)
			if !ok {
				return
			}

			go n.networkConnect(device, connType)
		}, nil,
		func(networkMenu *tview.Table) (int, int) {
			var width int

			for row, nw := range connTypes {
				ctype := nw.connType
				description := nw.desc

				if len(description) > width {
					width = len(description)
				}

				networkMenu.SetCell(
					row, 0, tview.NewTableCell(description).
						SetExpansion(1).
						SetReference(ctype).
						SetAlign(tview.AlignLeft).
						SetTextColor(theme.GetColor(theme.ThemeText)).
						SetSelectedStyle(
							tcell.Style{}.Reverse(true),
						),
				)
				networkMenu.SetCell(
					row, 1, tview.NewTableCell("("+strings.ToUpper(ctype.String())+")").
						SetAlign(tview.AlignRight).
						SetTextColor(theme.GetColor(theme.ThemeText)).
						SetSelectedStyle(
							tcell.Style{}.Reverse(true),
						),
				)
			}

			return width, 0
		},
	)
}

// networkConnect connects to the network with the selected network type.
func (n *networkView) networkConnect(device bluetooth.DeviceData, connType bluetooth.NetworkType) {
	info := fmt.Sprintf(
		"%s (%s)",
		getDeviceDisplayName(device.DeviceEventData), strings.ToUpper(connType.String()),
	)

	deviceName := getDeviceDisplayName(device.DeviceEventData)

	n.op.startOperation(
		func() {
			n.status.InfoMessage("正在连接 "+info, true)
			err := n.app.Session().Network(device.DeviceAddress).Connect(deviceName, connType)
			if err != nil {
				n.status.ErrorMessage(err)
				return
			}
			n.status.InfoMessage("已连接到 "+info, false)
		},
		func() {
			err := n.app.Session().Network(device.DeviceAddress).Disconnect()
			if err != nil {
				n.status.ErrorMessage(err)
				return
			}
			n.status.InfoMessage("已取消连接到 "+info, false)
		},
	)
}
