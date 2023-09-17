package main

import (
	"apple-store-helper/common"
	"apple-store-helper/services"
	"apple-store-helper/theme"
	"apple-store-helper/view"
	"errors"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
)

// main 主函数 (Main function)
func main() {
	initMP3Player()
	initFyneApp()

	// 默认地区 (Default Area)
	defaultArea := services.Listen.Area.Title

	// 门店选择器 (Store Selector)
	storeWidget := widget.NewSelect(services.Store.ByAreaTitleForOptions(defaultArea), nil)
	storeWidget.PlaceHolder = "Please select a pick-up store"

	// 型号选择器 (Product Selector)
	productWidget := widget.NewSelect(services.Product.ByAreaTitleForOptions(defaultArea), nil)
	productWidget.PlaceHolder = "Please select an iPhone model"

	// Bark 通知输入框
	barkWidget := widget.NewEntry()
	barkWidget.SetPlaceHolder("https://api.day.app/你的BarkKey")

	// 地区选择器 (Area Selector)
	areaWidget := widget.NewRadioGroup(services.Area.ForOptions(), func(value string) {
		storeWidget.Options = services.Store.ByAreaTitleForOptions(value)
		storeWidget.ClearSelected()

		productWidget.Options = services.Product.ByAreaTitleForOptions(value)
		productWidget.ClearSelected()

		services.Listen.Area = services.Area.GetArea(value)
		services.Listen.Clean()
	})

	areaWidget.Horizontal = true

	help := `1. Add the model you wish to purchase to your cart on the Apple official website.
2. Select your region, store, and model, then click the "Add" button to add the models you'd like to monitor to the tracking list.
3. Click the "Start" button to begin monitoring. The cart page will automatically open when stock is available.
`

	loadUserSettingsCache(areaWidget, storeWidget, productWidget, barkWidget)

	// 初始化 GUI 窗口内容 (Initialize GUI)
	view.Window.SetContent(container.NewVBox(
		widget.NewLabel(help),
		container.New(layout.NewFormLayout(), widget.NewLabel("Select Region:"), areaWidget),
		container.New(layout.NewFormLayout(), widget.NewLabel("Select Store:"), storeWidget),
		container.New(layout.NewFormLayout(), widget.NewLabel("Select Model:"), productWidget),
		container.New(layout.NewFormLayout(), widget.NewLabel("Bark Notification Address"), barkWidget),

		container.NewBorder(nil, nil,
			createActionButtons(areaWidget, storeWidget, productWidget, barkWidget),
			createControlButtons(),
		),

		services.Listen.Logs,
		layout.NewSpacer(),
		createVersionLabel(),
	))

	view.Window.Resize(fyne.NewSize(1000, 800))
	services.Listen.Run()
	view.Window.ShowAndRun()
}

// initMP3Player 初始化 MP3 播放器 (Initialize MP3 player)
func initMP3Player() {
	SampleRate := beep.SampleRate(44100)
	speaker.Init(SampleRate, SampleRate.N(time.Second/10))
}

// initFyneApp 初始化 Fyne 应用 (Initialize Fyne App)
func initFyneApp() {
	view.App = app.NewWithID("apple-store-helper")
	view.App.Settings().SetTheme(&theme.MyTheme{})
	view.Window = view.App.NewWindow("Apple Store Helper")
}

// 加载用户设置缓存 (Load user settings cache)
func loadUserSettingsCache(areaWidget *widget.RadioGroup, storeWidget *widget.Select, productWidget *widget.Select, barkNotifyWidget *widget.Entry) {
	settings, err := services.LoadSettings()
	if err == nil {
		areaWidget.SetSelected(settings.SelectedArea)
		storeWidget.SetSelected(settings.SelectedStore)
		productWidget.SetSelected(settings.SelectedProduct)
		services.Listen.SetListenItems(settings.ListenItems)
		barkNotifyWidget.Text = settings.BarkNotifyUrl
		services.Listen.BarkNotifyUrl = settings.BarkNotifyUrl
	} else {
		areaWidget.SetSelected(services.Listen.Area.Title)
	}
}

// 创建动作按钮 (Create action buttons)
func createActionButtons(areaWidget *widget.RadioGroup, storeWidget *widget.Select, productWidget *widget.Select, barkNotifyWidget *widget.Entry) *fyne.Container {
	return container.NewHBox(
		widget.NewButton("Add", func() {
			if storeWidget.Selected == "" || productWidget.Selected == "" {
				dialog.ShowError(errors.New("Please select store and model"), view.Window)
			} else {
				services.Listen.Add(areaWidget.Selected, storeWidget.Selected, productWidget.Selected, barkNotifyWidget.Text)
				services.SaveSettings(services.UserSettings{
					SelectedArea:    areaWidget.Selected,
					SelectedStore:   storeWidget.Selected,
					SelectedProduct: productWidget.Selected,
					BarkNotifyUrl:   barkNotifyWidget.Text,
					ListenItems:     services.Listen.GetListenItems(),
				})
			}
		}),
		widget.NewButton("Reset", func() {
			services.Listen.Clean()
			services.ClearSettings()
		}),
		widget.NewButton("Try it out (prompt sound when in stock)", func() {
			go services.Listen.AlertMp3()
		}),
		widget.NewButton("Try Bark notification", func() {
			services.Listen.BarkNotifyUrl = barkNotifyWidget.Text
			services.Listen.SendPushNotificationByBark("In stock reminder (test)", "This is a test reminder. Clicking on the notification will redirect you to the relevant link", "https://www.apple.com.cn/")
		}),
	)
}

// 创建控制按钮 (Create control buttons)
func createControlButtons() *fyne.Container {
	return container.NewHBox(
		widget.NewButton("Start", func() {
			services.Listen.Status.Set(services.Running)
		}),
		widget.NewButton("Pause", func() {
			services.Listen.Status.Set(services.Pause)
		}),
		container.NewCenter(widget.NewLabel("Status:")),
		container.NewCenter(widget.NewLabelWithData(services.Listen.Status)),
	)
}

// createVersionLabel 创建版本标签 (Create version label)
func createVersionLabel() *fyne.Container {
	return container.NewHBox(
		layout.NewSpacer(),
		widget.NewLabel("version: "+common.VERSION),
	)
}
