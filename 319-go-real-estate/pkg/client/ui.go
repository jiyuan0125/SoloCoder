package client

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"realestate/pkg/common"
)

type UI struct {
	client   *Client
	scanner  *bufio.Scanner
	userID   string
	userType string
}

func NewUI(client *Client) *UI {
	return &UI{
		client:  client,
		scanner: bufio.NewScanner(os.Stdin),
	}
}

func (ui *UI) readLine(prompt string) string {
	fmt.Print(prompt)
	ui.scanner.Scan()
	return strings.TrimSpace(ui.scanner.Text())
}

func (ui *UI) readInt(prompt string) int {
	for {
		input := ui.readLine(prompt)
		if input == "" {
			return 0
		}
		val, err := strconv.Atoi(input)
		if err == nil {
			return val
		}
		fmt.Println("请输入有效的数字")
	}
}

func (ui *UI) readFloat(prompt string) float64 {
	for {
		input := ui.readLine(prompt)
		if input == "" {
			return 0
		}
		val, err := strconv.ParseFloat(input, 64)
		if err == nil {
			return val
		}
		fmt.Println("请输入有效的数字")
	}
}

func (ui *UI) Start() error {
	fmt.Println("=== 房产信息发布平台 ===")
	fmt.Println()
	fmt.Println("请选择身份:")
	fmt.Println("1. 房东")
	fmt.Println("2. 租客/买家")
	fmt.Println("3. 退出")

	choice := ui.readLine("请输入选项: ")

	switch choice {
	case "1":
		return ui.landlordMenu()
	case "2":
		return ui.userMenu()
	case "3":
		fmt.Println("再见!")
		return nil
	default:
		fmt.Println("无效选项")
		return ui.Start()
	}
}

func (ui *UI) landlordMenu() error {
	fmt.Println()
	ui.userID = ui.readLine("请输入您的房东ID: ")
	if ui.userID == "" {
		fmt.Println("ID不能为空")
		return ui.landlordMenu()
	}
	ui.userType = "landlord"

	for {
		fmt.Println()
		fmt.Println("=== 房东功能菜单 ===")
		fmt.Println("1. 发布新房源")
		fmt.Println("2. 查看我的房源")
		fmt.Println("3. 编辑房源")
		fmt.Println("4. 下架房源")
		fmt.Println("5. 标记房源为成交")
		fmt.Println("6. 返回上级")

		choice := ui.readLine("请输入选项: ")

		switch choice {
		case "1":
			ui.createProperty()
		case "2":
			ui.listMyProperties()
		case "3":
			ui.editProperty()
		case "4":
			ui.offlineProperty()
		case "5":
			ui.sellProperty()
		case "6":
			return ui.Start()
		default:
			fmt.Println("无效选项")
		}
	}
}

func (ui *UI) createProperty() {
	fmt.Println()
	fmt.Println("=== 发布新房源 ===")

	community := ui.readLine("小区名称: ")
	if community == "" {
		fmt.Println("小区名称不能为空")
		return
	}

	fmt.Println()
	fmt.Println("户型选项:")
	for i, t := range common.ValidHouseTypes {
		fmt.Printf("%d. %s\n", i+1, t)
	}
	houseTypeChoice := ui.readInt("请选择户型编号: ")
	if houseTypeChoice < 1 || houseTypeChoice > len(common.ValidHouseTypes) {
		fmt.Println("无效选项")
		return
	}
	houseType := common.ValidHouseTypes[houseTypeChoice-1]

	area := ui.readFloat("面积(平方米): ")
	if area <= 0 {
		fmt.Println("面积必须大于0")
		return
	}

	floor := ui.readInt("楼层: ")
	if floor <= 0 {
		fmt.Println("楼层必须大于0")
		return
	}

	orientation := ui.readLine("朝向: ")

	fmt.Println()
	fmt.Println("价格类型:")
	fmt.Println("1. 月租")
	fmt.Println("2. 售价")
	fmt.Println("3. 面议")

	priceTypeChoice := ui.readInt("请选择价格类型: ")
	var priceType common.PriceType
	var price float64

	switch priceTypeChoice {
	case 1:
		priceType = common.PriceTypeRent
		price = ui.readFloat("月租金: ")
		if price <= 0 {
			fmt.Println("价格必须大于0")
			return
		}
	case 2:
		priceType = common.PriceTypeSell
		price = ui.readFloat("售价: ")
		if price <= 0 {
			fmt.Println("价格必须大于0")
			return
		}
	case 3:
		priceType = common.PriceTypeNegotiable
	default:
		fmt.Println("无效选项")
		return
	}

	contact := ui.readLine("联系方式: ")
	if contact == "" {
		fmt.Println("联系方式不能为空")
		return
	}

	req := CreatePropertyRequest{
		LandlordID:  ui.userID,
		Community:   community,
		HouseType:   houseType,
		Area:        area,
		Floor:       floor,
		Orientation: orientation,
		PriceType:   priceType,
		Price:       price,
		Contact:     contact,
	}

	property, err := ui.client.CreateProperty(req)
	if err != nil {
		fmt.Printf("发布失败: %v\n", err)
		return
	}

	fmt.Printf("发布成功! 房源ID: %s\n", property.ID)
}

func (ui *UI) listMyProperties() {
	properties, err := ui.client.GetLandlordProperties(ui.userID)
	if err != nil {
		fmt.Printf("获取房源列表失败: %v\n", err)
		return
	}

	if len(properties) == 0 {
		fmt.Println("您还没有发布任何房源")
		return
	}

	fmt.Println()
	fmt.Println("=== 我的房源列表 ===")
	for i, p := range properties {
		statusText := "上架中"
		if p.Status == common.StatusOffline {
			statusText = "已下架"
		} else if p.Status == common.StatusSold {
			statusText = "已成交"
		}

		priceText := ui.formatPrice(p)

		fmt.Printf("\n[%d] ID: %s\n", i+1, p.ID)
		fmt.Printf("    小区: %s\n", p.Community)
		fmt.Printf("    户型: %s | 面积: %.1f㎡ | 楼层: %d层\n", p.HouseType, p.Area, p.Floor)
		fmt.Printf("    朝向: %s | 价格: %s\n", p.Orientation, priceText)
		fmt.Printf("    联系方式: %s\n", p.Contact)
		fmt.Printf("    状态: %s\n", statusText)
	}
}

func (ui *UI) formatPrice(p *common.Property) string {
	switch p.PriceType {
	case common.PriceTypeRent:
		return fmt.Sprintf("%.0f元/月", p.Price)
	case common.PriceTypeSell:
		return fmt.Sprintf("%.0f万元", p.Price)
	case common.PriceTypeNegotiable:
		return "面议"
	default:
		return fmt.Sprintf("%.0f", p.Price)
	}
}

func (ui *UI) editProperty() {
	propertyID := ui.readLine("请输入要编辑的房源ID: ")
	if propertyID == "" {
		return
	}

	property, err := ui.client.GetProperty(propertyID)
	if err != nil {
		fmt.Printf("获取房源信息失败: %v\n", err)
		return
	}

	if property.LandlordID != ui.userID {
		fmt.Println("您只能编辑自己发布的房源")
		return
	}

	if property.Status == common.StatusSold {
		fmt.Println("已成交的房源不能编辑")
		return
	}

	fmt.Println()
	fmt.Println("=== 编辑房源 ===")
	fmt.Println("(直接回车跳过不修改)")

	req := UpdatePropertyRequest{}

	community := ui.readLine(fmt.Sprintf("小区名称 [%s]: ", property.Community))
	if community != "" {
		req.Community = community
	}

	fmt.Println()
	fmt.Println("户型选项:")
	for i, t := range common.ValidHouseTypes {
		fmt.Printf("%d. %s\n", i+1, t)
	}
	houseTypeChoice := ui.readLine(fmt.Sprintf("请选择户型编号 [%s]: ", property.HouseType))
	if houseTypeChoice != "" {
		idx, _ := strconv.Atoi(houseTypeChoice)
		if idx >= 1 && idx <= len(common.ValidHouseTypes) {
			req.HouseType = common.ValidHouseTypes[idx-1]
		}
	}

	areaStr := ui.readLine(fmt.Sprintf("面积(平方米) [%.1f]: ", property.Area))
	if areaStr != "" {
		area, _ := strconv.ParseFloat(areaStr, 64)
		if area > 0 {
			req.Area = area
		}
	}

	floorStr := ui.readLine(fmt.Sprintf("楼层 [%d]: ", property.Floor))
	if floorStr != "" {
		floor, _ := strconv.Atoi(floorStr)
		if floor > 0 {
			req.Floor = floor
		}
	}

	orientation := ui.readLine(fmt.Sprintf("朝向 [%s]: ", property.Orientation))
	if orientation != "" {
		req.Orientation = orientation
	}

	fmt.Println()
	fmt.Println("价格类型:")
	fmt.Println("1. 月租")
	fmt.Println("2. 售价")
	fmt.Println("3. 面议")
	priceTypeChoice := ui.readLine("请选择价格类型 (跳过不修改): ")
	if priceTypeChoice != "" {
		switch priceTypeChoice {
		case "1":
			req.PriceType = common.PriceTypeRent
			price := ui.readFloat("月租金: ")
			if price > 0 {
				req.Price = price
			}
		case "2":
			req.PriceType = common.PriceTypeSell
			price := ui.readFloat("售价: ")
			if price > 0 {
				req.Price = price
			}
		case "3":
			req.PriceType = common.PriceTypeNegotiable
		}
	}

	contact := ui.readLine(fmt.Sprintf("联系方式 [%s]: ", property.Contact))
	if contact != "" {
		req.Contact = contact
	}

	_, err = ui.client.UpdateProperty(propertyID, req)
	if err != nil {
		fmt.Printf("编辑失败: %v\n", err)
		return
	}

	fmt.Println("编辑成功!")
}

func (ui *UI) offlineProperty() {
	propertyID := ui.readLine("请输入要下架的房源ID: ")
	if propertyID == "" {
		return
	}

	property, err := ui.client.GetProperty(propertyID)
	if err != nil {
		fmt.Printf("获取房源信息失败: %v\n", err)
		return
	}

	if property.LandlordID != ui.userID {
		fmt.Println("您只能下架自己发布的房源")
		return
	}

	if property.Status == common.StatusSold {
		fmt.Println("已成交的房源不能下架")
		return
	}

	err = ui.client.OfflineProperty(propertyID)
	if err != nil {
		fmt.Printf("下架失败: %v\n", err)
		return
	}

	fmt.Println("房源已下架")
}

func (ui *UI) sellProperty() {
	propertyID := ui.readLine("请输入要标记为成交的房源ID: ")
	if propertyID == "" {
		return
	}

	property, err := ui.client.GetProperty(propertyID)
	if err != nil {
		fmt.Printf("获取房源信息失败: %v\n", err)
		return
	}

	if property.LandlordID != ui.userID {
		fmt.Println("您只能操作自己发布的房源")
		return
	}

	err = ui.client.SellProperty(propertyID)
	if err != nil {
		fmt.Printf("操作失败: %v\n", err)
		return
	}

	fmt.Println("房源已标记为成交")
}

func (ui *UI) userMenu() error {
	fmt.Println()
	ui.userID = ui.readLine("请输入您的用户ID: ")
	if ui.userID == "" {
		fmt.Println("ID不能为空")
		return ui.userMenu()
	}
	ui.userType = "user"

	for {
		fmt.Println()
		fmt.Println("=== 用户功能菜单 ===")
		fmt.Println("1. 按条件筛选房源")
		fmt.Println("2. 收藏房源")
		fmt.Println("3. 取消收藏")
		fmt.Println("4. 查看我的收藏")
		fmt.Println("5. 返回上级")

		choice := ui.readLine("请输入选项: ")

		switch choice {
		case "1":
			ui.filterProperties()
		case "2":
			ui.addFavorite()
		case "3":
			ui.removeFavorite()
		case "4":
			ui.listFavorites()
		case "5":
			return ui.Start()
		default:
			fmt.Println("无效选项")
		}
	}
}

func (ui *UI) filterProperties() {
	fmt.Println()
	fmt.Println("=== 筛选房源 ===")
	fmt.Println("(直接回车表示不限制)")

	var filter common.FilterRequest

	minPriceStr := ui.readLine("最低价格: ")
	if minPriceStr != "" {
		minPrice, _ := strconv.ParseFloat(minPriceStr, 64)
		filter.MinPrice = &minPrice
	}

	maxPriceStr := ui.readLine("最高价格: ")
	if maxPriceStr != "" {
		maxPrice, _ := strconv.ParseFloat(maxPriceStr, 64)
		filter.MaxPrice = &maxPrice
	}

	minAreaStr := ui.readLine("最小面积(㎡): ")
	if minAreaStr != "" {
		minArea, _ := strconv.ParseFloat(minAreaStr, 64)
		filter.MinArea = &minArea
	}

	maxAreaStr := ui.readLine("最大面积(㎡): ")
	if maxAreaStr != "" {
		maxArea, _ := strconv.ParseFloat(maxAreaStr, 64)
		filter.MaxArea = &maxArea
	}

	fmt.Println()
	fmt.Println("户型选项 (输入编号,多个用逗号分隔):")
	for i, t := range common.ValidHouseTypes {
		fmt.Printf("%d. %s\n", i+1, t)
	}
	houseTypeStr := ui.readLine("选择户型: ")
	if houseTypeStr != "" {
		parts := strings.Split(houseTypeStr, ",")
		for _, p := range parts {
			idx, _ := strconv.Atoi(strings.TrimSpace(p))
			if idx >= 1 && idx <= len(common.ValidHouseTypes) {
				filter.HouseTypes = append(filter.HouseTypes, common.ValidHouseTypes[idx-1])
			}
		}
	}

	fmt.Println()
	fmt.Println("楼层档位 (输入编号,多个用逗号分隔):")
	fmt.Println("1. 低层(1-3层)")
	fmt.Println("2. 中层(4-10层)")
	fmt.Println("3. 高层(11层以上)")
	floorLevelStr := ui.readLine("选择楼层档位: ")
	if floorLevelStr != "" {
		parts := strings.Split(floorLevelStr, ",")
		for _, p := range parts {
			idx, _ := strconv.Atoi(strings.TrimSpace(p))
			switch idx {
			case 1:
				filter.FloorLevels = append(filter.FloorLevels, common.FloorLow)
			case 2:
				filter.FloorLevels = append(filter.FloorLevels, common.FloorMiddle)
			case 3:
				filter.FloorLevels = append(filter.FloorLevels, common.FloorHigh)
			}
		}
	}

	properties, err := ui.client.FilterProperties(filter)
	if err != nil {
		fmt.Printf("筛选失败: %v\n", err)
		return
	}

	if len(properties) == 0 {
		fmt.Println("未找到符合条件的房源")
		return
	}

	fmt.Println()
	fmt.Println("=== 筛选结果 ===")
	for i, p := range properties {
		priceText := ui.formatPrice(p)
		floorLevel := common.GetFloorLevel(p.Floor)
		floorText := common.FloorLevelMap[floorLevel]

		fmt.Printf("\n[%d] ID: %s\n", i+1, p.ID)
		fmt.Printf("    小区: %s\n", p.Community)
		fmt.Printf("    户型: %s | 面积: %.1f㎡ | %s(%d层)\n", p.HouseType, p.Area, floorText, p.Floor)
		fmt.Printf("    朝向: %s | 价格: %s\n", p.Orientation, priceText)
		fmt.Printf("    联系方式: %s\n", p.Contact)

		isFavorited, _ := ui.client.CheckFavorite(ui.userID, p.ID)
		if isFavorited {
			fmt.Printf("    [已收藏]\n")
		}
	}
}

func (ui *UI) addFavorite() {
	propertyID := ui.readLine("请输入要收藏的房源ID: ")
	if propertyID == "" {
		return
	}

	err := ui.client.AddFavorite(ui.userID, propertyID)
	if err != nil {
		fmt.Printf("收藏失败: %v\n", err)
		return
	}

	fmt.Println("收藏成功!")
}

func (ui *UI) removeFavorite() {
	propertyID := ui.readLine("请输入要取消收藏的房源ID: ")
	if propertyID == "" {
		return
	}

	err := ui.client.RemoveFavorite(ui.userID, propertyID)
	if err != nil {
		fmt.Printf("取消收藏失败: %v\n", err)
		return
	}

	fmt.Println("已取消收藏")
}

func (ui *UI) listFavorites() {
	items, err := ui.client.GetFavorites(ui.userID)
	if err != nil {
		fmt.Printf("获取收藏列表失败: %v\n", err)
		return
	}

	if len(items) == 0 {
		fmt.Println("您还没有收藏任何房源")
		return
	}

	fmt.Println()
	fmt.Println("=== 我的收藏 ===")
	for i, item := range items {
		if item.PropertyRemoved {
			fmt.Printf("\n[%d] 房源ID: %s\n", i+1, item.PropertyID)
			fmt.Printf("    [房源已移除]\n")
		} else if item.Property != nil {
			p := item.Property
			priceText := ui.formatPrice(p)

			fmt.Printf("\n[%d] ID: %s\n", i+1, p.ID)
			fmt.Printf("    小区: %s\n", p.Community)
			fmt.Printf("    户型: %s | 面积: %.1f㎡ | 楼层: %d层\n", p.HouseType, p.Area, p.Floor)
			fmt.Printf("    朝向: %s | 价格: %s\n", p.Orientation, priceText)
			fmt.Printf("    联系方式: %s\n", p.Contact)

			if item.StatusText != "" {
				fmt.Printf("    [%s]\n", item.StatusText)
			}
		}
	}
}
